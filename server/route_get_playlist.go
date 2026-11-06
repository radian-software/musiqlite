package main

import (
	"database/sql"
	"encoding/json"
	"errors"
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
	row := s.DB.QueryRowContext(r.Context(), fmt.Sprintf("SELECT oid, is_upnext_playlist, playback_position FROM playlist WHERE name = %s", quoteSingle(name)))
	var playlistOID int64
	var isUpNext bool
	var playbackPosition *int
	err = row.Scan(&playlistOID, &isUpNext, &playbackPosition)
	if errors.Is(err, sql.ErrNoRows) {
		return expectedError("no_such_playlist", "no such playlist", http.StatusNotFound)
	} else if err != nil {
		return fmt.Errorf("query db for playlist info: %w", err)
	}
	queryPart1 := "SELECT ps.oid, "
	queryPart2s := []string{}
	for _, tag := range includedTags {
		queryPart2s = append(queryPart2s, fmt.Sprintf(`s.%s`, quoteDouble(tag)))
	}
	queryPart3 := fmt.Sprintf(" FROM song AS s JOIN playlist_song AS ps ON s.oid = ps.song_oid WHERE ps.playlist_oid = %d", playlistOID)
	query := queryPart1 + strings.Join(queryPart2s, ", ") + queryPart3
	fmt.Println(query)
	rows, err := s.DB.QueryContext(r.Context(), query)
	if err != nil {
		return fmt.Errorf("query db for songs: %w", err)
	}
	songsInfo := []any{}
	for rows.Next() {
		var entryOID int64
		tagValues := make([]any, len(includedTags))
		cols := make([]any, len(includedTags)+1)
		cols[0] = &entryOID
		for idx := range includedTags {
			cols[idx+1] = &tagValues[idx]
		}
		err := rows.Scan(cols...)
		if err != nil {
			return fmt.Errorf("scan from db for songs: %w", err)
		}
		songInfo := map[string]any{
			"entry_oid": entryOID,
		}
		for idx, name := range includedTags {
			songInfo[name] = tagValues[idx]
		}
		songsInfo = append(songsInfo, songInfo)
	}
	if rows.Err() != nil {
		return fmt.Errorf("scan from db for songs: %w", err)
	}
	result := map[string]any{
		"is_upnext": isUpNext,
		"songs":     songsInfo,
	}
	if isUpNext {
		result["playback_position"] = playbackPosition
	}
	json.NewEncoder(w).Encode(result)
	return nil
}
