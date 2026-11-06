package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (s *Server) handleGetPlaylist(w http.ResponseWriter, r *http.Request) error {
	allowedTagNames, err := s.getAllowedTagNames(r.Context())
	if err != nil {
		return err
	}
	name := r.URL.Query().Get("name")
	if name == "" {
		return clientErrorf("no playlist name")
	}
	includedTags := r.URL.Query()["tag"]
	for _, tag := range includedTags {
		if tag == "" {
			return clientErrorf("empty tag name")
		}
		if !allowedTagNames[tag] {
			return clientErrorf("requested tag does not exist: %s", tag)
		}
	}
	if len(includedTags) == 0 {
		return clientErrorf("no tags requested")
	}
	queryPart1 := "SELECT "
	queryPart2s := []string{}
	for _, tag := range includedTags {
		queryPart2s = append(queryPart2s, fmt.Sprintf(`s.%s`, quoteDouble(tag)))
	}
	queryPart3 := fmt.Sprintf(" FROM song AS s JOIN playlist_song AS ps ON s.oid = ps.song_oid JOIN playlist AS p ON ps.playlist_oid = p.oid WHERE p.name = %s", quoteSingle(name))
	query := queryPart1 + strings.Join(queryPart2s, ", ") + queryPart3
	fmt.Println(query)
	rows, err := s.DB.QueryContext(r.Context(), query)
	if err != nil {
		return fmt.Errorf("query db for songs: %w", err)
	}
	songsInfo := []any{}
	for rows.Next() {
		tagValues := make([]any, len(includedTags))
		cols := make([]any, len(includedTags))
		for idx := range includedTags {
			cols[idx] = &tagValues[idx]
		}
		err := rows.Scan(cols...)
		if err != nil {
			return fmt.Errorf("scan from db for songs: %w", err)
		}
		songInfo := map[string]any{}
		for idx, name := range includedTags {
			songInfo[name] = tagValues[idx]
		}
		songsInfo = append(songsInfo, songInfo)
	}
	if rows.Err() != nil {
		return fmt.Errorf("scan from db for songs: %w", err)
	}
	json.NewEncoder(w).Encode(songsInfo)
	return nil
}
