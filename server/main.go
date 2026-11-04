package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func quoteSingle(text string) string {
	return `'` + strings.ReplaceAll(text, `'`, `''`) + `'`
}

func quoteDouble(text string) string {
	return `"` + strings.ReplaceAll(text, `"`, `""`) + `"`
}

type Server struct {
	DB        *sql.DB
	ObjectDir string
	Log       *log.Logger
}

func (s *Server) Serve(ctx context.Context, addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v0/media/{filename}", s.wrapRequestHandler(s.handleGetMedia))
	mux.HandleFunc("/api/v0/albums", s.wrapRequestHandler(s.handleGetAlbums))
	mux.HandleFunc("/api/v0/groups", s.wrapRequestHandler(s.handleGetGroups))
	server := &http.Server{Addr: addr, Handler: mux}
	go func() {
		<-ctx.Done()
		server.Close()
	}()
	nl, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("binding to address %s: %w", addr, err)
	}
	s.Log.Printf("Listening on %s", addr)
	return server.Serve(nl)
}

type returnableError struct {
	UserMessage    string
	InternalError  error
	HTTPStatusCode int
}

func clientError(err error, code int) *returnableError {
	return &returnableError{
		UserMessage:    "An unexpected client-side error occurred. This is a bug",
		InternalError:  fmt.Errorf("client error (%d): %w", code, err),
		HTTPStatusCode: code,
	}
}

func (h *returnableError) Error() string {
	return h.InternalError.Error()
}

func getReturnableError(err error) *returnableError {
	if err, yes := errors.AsType[*returnableError](err); yes {
		return err
	}
	return &returnableError{
		UserMessage:    "An unexpected server-side error occurred. This is a bug",
		InternalError:  fmt.Errorf("unknown error: %w", err),
		HTTPStatusCode: http.StatusInternalServerError,
	}
}

func (s *Server) wrapRequestHandler(handler func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		err := handler(w, r)
		if err != nil {
			returnErr := getReturnableError(err)
			if returnErr.HTTPStatusCode == http.StatusInternalServerError {
				s.Log.Printf("Internal error on request for %s %s: %s", r.Method, r.URL, returnErr.InternalError)
			}
			w.WriteHeader(returnErr.HTTPStatusCode)
			json.NewEncoder(w).Encode(map[string]any{
				"user_message":   returnErr.UserMessage,
				"internal_error": returnErr.InternalError.Error(),
			})
			return
		}
	}
}

var validMediaRegexp = regexp.MustCompile(`^([a-z0-9]{8,128})(\.[a-z0-9]{2,4})$`)

func (s *Server) handleGetMedia(w http.ResponseWriter, r *http.Request) error {
	mediaFilename := r.PathValue("filename")
	mediaFilenameMatch := validMediaRegexp.FindStringSubmatch(mediaFilename)
	if mediaFilenameMatch == nil {
		return clientError(fmt.Errorf("not a valid media filename: %s", mediaFilename), http.StatusBadRequest)
	}
	digest, ext := mediaFilenameMatch[1], mediaFilenameMatch[2]
	objectPath := path.Join(s.ObjectDir, digest[:2], digest[2:])
	file, err := os.Open(objectPath)
	if os.IsNotExist(err) {
		return clientError(fmt.Errorf("no such media: %s", digest), http.StatusNotFound)
	} else if err != nil {
		return err
	}
	defer file.Close()
	st, err := file.Stat()
	if err != nil {
		return err
	}
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		return clientError(fmt.Errorf("unsupported file extension: %s", ext), http.StatusBadRequest)
	}
	w.Header().Set("content-type", mimeType)
	w.Header().Set("content-length", fmt.Sprintf("%d", st.Size()))
	_, err = io.Copy(w, file)
	if err != nil {
		s.Log.Printf("Error while writing media response: %s", err.Error())
	}
	return nil
}

func (s *Server) handleGetGroups(w http.ResponseWriter, r *http.Request) error {
	partition := r.URL.Query().Get("partition")
	if partition == "" {
		return clientError(errors.New("missing query parameter: partition"), http.StatusBadRequest)
	}
	includedTags := r.URL.Query()["tag"]
	for _, tag := range includedTags {
		if tag == "" {
			return clientError(errors.New("empty tag name"), http.StatusBadRequest)
		}
	}
	sort := r.URL.Query().Get("sort")
	rows, err := s.DB.QueryContext(r.Context(), fmt.Sprintf("SELECT %[1]s, COUNT(*) FROM song GROUP BY %[1]s", quoteDouble(partition)))
	if err != nil {
		return fmt.Errorf("query db for counting songs by partition: %w", err)
	}
	songCounts := map[string]int{}
	for rows.Next() {
		var name string
		var count int
		err := rows.Scan(&name, &count)
		if err != nil {
			return fmt.Errorf("scan from db for counting songs by partition: %w", err)
		}
		songCounts[name] = count
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("scan from db for counting songs by partition: %w", err)
	}
	rows, err = s.DB.QueryContext(r.Context(), "SELECT name FROM pragma_table_info('song')")
	if err != nil {
		return fmt.Errorf("query db for listing song tag names: %w", err)
	}
	allowedTagNames := map[string]bool{}
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			return fmt.Errorf("scan from db for listing song tag names: %w", err)
		}
		allowedTagNames[name] = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("scan from db for listing song tag names: %w", err)
	}
	for _, tagName := range append(includedTags, partition, sort) {
		if tagName != "" && !allowedTagNames[tagName] {
			return clientError(fmt.Errorf("requested tag does not exist: %s", partition), http.StatusUnprocessableEntity)
		}
	}
	groups := []any{}
	if len(includedTags) > 0 {
		queryPart1 := "WITH cte AS (SELECT *, "
		queryPart2s := []string{}
		for _, name := range includedTags {
			queryPart2s = append(queryPart2s, fmt.Sprintf("COUNT(*) OVER (PARTITION BY %s, %s) AS %s", quoteDouble(partition), quoteDouble(name), quoteDouble("_"+name+"_count")))
		}
		queryPart3 := fmt.Sprintf(" FROM song) SELECT DISTINCT %s, ", quoteDouble(partition))
		queryPart4s := []string{}
		for _, name := range includedTags {
			queryPart4s = append(queryPart4s, fmt.Sprintf("FIRST_VALUE(%[1]s) OVER (PARTITION BY %[2]s ORDER BY %[3]s DESC), %[3]s", quoteDouble(name), quoteDouble(partition), quoteDouble("_"+name+"_count")))
		}
		queryPart5 := " FROM cte"
		queryPart6 := ""
		if sort != "" {
			queryPart6 = fmt.Sprintf(" ORDER BY %s COLLATE NOCASE ASC", quoteDouble(sort))
		}
		query := queryPart1 + strings.Join(queryPart2s, ", ") + queryPart3 + strings.Join(queryPart4s, ", ") + queryPart5 + queryPart6
		rows, err = s.DB.QueryContext(r.Context(), query)
		if err != nil {
			return fmt.Errorf("query db for counting tags within partitions: %w", err)
		}
		for rows.Next() {
			var partitionKey string
			tagValues := make([]any, len(includedTags))
			tagCounts := make([]int, len(includedTags))
			cols := make([]any, 2*len(includedTags)+1)
			cols[0] = &partitionKey
			for idx := range includedTags {
				cols[2*idx+1] = &tagValues[idx]
				cols[2*idx+2] = &tagCounts[idx]
			}
			err := rows.Scan(cols...)
			if err != nil {
				return fmt.Errorf("scan from db for counting tags within partitions: %w", err)
			}
			partitionTotal := songCounts[partitionKey]
			tagsInfo := map[string]any{}
			for idx, name := range includedTags {
				tagsInfo[name] = map[string]any{
					"most_common_value": tagValues[idx],
					"most_common_count": tagCounts[idx],
				}
			}
			groups = append(groups, map[string]any{
				"partition_tag_value": partitionKey,
				"song_count":          partitionTotal,
				"tag_values":          tagsInfo,
			})
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("scan from db for counting tags within partitions: %w", err)
		}
	} else {
		for partitionKey, partitionTotal := range songCounts {
			groups = append(groups, map[string]any{
				"partition_tag_value": partitionKey,
				"song_count":          partitionTotal,
				"tag_values":          map[string]any{},
			})
		}
	}
	json.NewEncoder(w).Encode(groups)
	return nil
}

func (s *Server) handleGetAlbums(w http.ResponseWriter, r *http.Request) error {
	rows, err := s.DB.QueryContext(r.Context(), "SELECT DISTINCT album FROM song ORDER BY coalesce(album_sort, album) COLLATE NOCASE ASC")
	if err != nil {
		return err
	}
	albums := []string{}
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			return err
		}
		albums = append(albums, name)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	json.NewEncoder(w).Encode(albums)
	return nil
}

func mainE() error {
	// Do the command-line argument parsing
	dbFile := flag.String("db", "library.sqlite3", "path to sqlite database")
	filestore := flag.String("filestore", "filestore", "path to object store directory")
	addr := flag.String("addr", ":8080", "host and port to listen for http on")
	flag.Parse()

	// Open the database
	if _, err := os.Stat(*dbFile); err != nil {
		return err
	}
	db, err := sql.Open("sqlite3", *dbFile)
	if err != nil {
		return err
	}
	defer db.Close()

	// Verify the object store
	objDir := path.Join(*filestore, "objects")
	if stat, err := os.Stat(objDir); err != nil {
		return err
	} else if !stat.IsDir() {
		return fmt.Errorf("not a directory: %s", objDir)
	}

	// Set up a context that will be canceled on signal
	ctx, stop := context.WithCancelCause(context.Background())
	defer stop(errors.New("application closing"))
	appSignal := make(chan os.Signal, 3)
	signal.Notify(appSignal, os.Interrupt)
	go func() {
		<-appSignal
		stop(errors.New("interrupt signal received"))
	}()

	// Create the server struct
	s := &Server{
		DB:        db,
		ObjectDir: objDir,
		Log:       log.Default(),
	}

	// Start handling requests
	err = s.Serve(ctx, *addr)
	if ctxErr := context.Cause(ctx); ctxErr != nil {
		err = ctxErr
	}
	return err
}

func main() {
	err := mainE()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %s\n", err.Error())
	}
}
