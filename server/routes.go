package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func (s *Server) getHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v0/media/{filename}", s.wrapRequestHandler(s.handleGetMedia))
	mux.HandleFunc("GET /api/v0/albums", s.wrapRequestHandler(s.handleGetAlbums))
	mux.HandleFunc("GET /api/v0/groups", s.wrapRequestHandler(s.handleGetGroups))
	mux.HandleFunc("GET /api/v0/songs", s.wrapRequestHandler(s.handleGetSongs))
	mux.HandleFunc("PUT /api/v0/playlist", s.wrapRequestHandler(s.handlePutPlaylist))
	mux.HandleFunc("GET /api/v0/playlists", s.wrapRequestHandler(s.handleGetPlaylists))
	return mux
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
