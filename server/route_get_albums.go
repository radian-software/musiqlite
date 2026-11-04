package main

import (
	"encoding/json"
	"net/http"
)

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
