package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"

	_ "github.com/mattn/go-sqlite3"
)

type Server struct {
	DB  *sql.DB
	Log *log.Logger
}

func (s *Server) Serve(ctx context.Context, addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v0/albums", s.wrapRequestHandler(s.handleGetAlbums))
	server := &http.Server{Addr: addr, Handler: mux}
	go func() {
		<-ctx.Done()
		server.Close()
	}()
	nl, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.Log.Printf("Listening on %s", addr)
	return server.Serve(nl)
}

func (s *Server) wrapRequestHandler(handler func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		err := handler(w, r)
		if err != nil {
			s.Log.Printf("Internal error on request for %s %s: %s", r.Method, r.URL, err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]any{
				"user_error":     "An unexpected error occurred",
				"internal_error": err.Error(),
			})
			return
		}
	}
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
		DB:  db,
		Log: log.Default(),
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
