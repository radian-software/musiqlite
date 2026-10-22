package main

import (
	"context"
	"database/sql"
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

func mainE() error {
	// Do the command-line argument parsing
	dbFile := flag.String("db", "library.sqlite3", "path to sqlite database")
	addr := flag.String("addr", ":8080", "host and port to listen for http on")
	flag.Parse()

	// Open the database
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
