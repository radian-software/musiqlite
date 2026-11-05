package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *Server) handleGetPlaylists(w http.ResponseWriter, r *http.Request) error {
	rows, err := s.DB.QueryContext(r.Context(), "SELECT name, is_upnext_playlist FROM playlist ORDER BY is_upnext_playlist, name COLLATE NOCASE ASC")
	if err != nil {
		return fmt.Errorf("query db for playlists: %w", err)
	}
	playlists := []any{}
	for rows.Next() {
		var name string
		var isUpNext bool
		err := rows.Scan(&name, &isUpNext)
		if err != nil {
			return fmt.Errorf("scan from db for playlists: %w", err)
		}
		playlists = append(playlists, map[string]any{
			"name":               name,
			"is_upnext_playlist": isUpNext,
		})
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("scan from db for playlists: %w", err)
	}
	json.NewEncoder(w).Encode(playlists)
	return nil
}
