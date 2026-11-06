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
	mux.HandleFunc("GET /api/v0/playlist", s.wrapRequestHandler(s.handleGetPlaylist))
	return mux
}

type returnableError struct {
	UserMessage    string
	InternalError  error
	ErrorCode      string
	HTTPStatusCode int
}

func expectedError(errorCode string, description string, httpCode int) *returnableError {
	return &returnableError{
		// The error itself isn't a bug in the client, but if
		// the client fails to handle the error code and shows
		// the error to the user, then it is a client bug.
		UserMessage:    "An unexpected client-side error occurred. This is a bug",
		InternalError:  fmt.Errorf("client error (%d, %s): %s", httpCode, errorCode, description),
		ErrorCode:      errorCode,
		HTTPStatusCode: httpCode,
	}
}

func clientError(err error, code int) *returnableError {
	return &returnableError{
		UserMessage:    "An unexpected client-side error occurred. This is a bug",
		InternalError:  fmt.Errorf("client error (%d): %w", code, err),
		ErrorCode:      "unexpected_client_error",
		HTTPStatusCode: code,
	}
}

func clientErrorf(format string, a ...any) *returnableError {
	return clientError(fmt.Errorf(format, a...), http.StatusBadRequest)
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
		ErrorCode:      "unexpected_server_error",
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
				"error_code":     returnErr.ErrorCode,
				"internal_error": returnErr.InternalError.Error(),
			})
			return
		}
	}
}
