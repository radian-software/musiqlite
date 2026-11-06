package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (s *Server) handleGetSongs(w http.ResponseWriter, r *http.Request) error {
	allowedTagNames, err := s.getAllowedTagNames(r.Context())
	if err != nil {
		return err
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
	tagFilters := map[string]string{}
	for key, val := range r.URL.Query() {
		if !strings.HasPrefix(key, "filter-") {
			continue
		}
		tag := strings.TrimPrefix(key, "filter-")
		if !allowedTagNames[tag] {
			return clientErrorf("requested filter tag does not exist: %s", tag)
		}
		if len(val) < 1 {
			return clientErrorf("no value for filter on tag: %s", tag)
		}
		if len(val) > 1 {
			return clientErrorf("more than one filter on same tag: %s", tag)
		}
		tagFilters[tag] = val[0]
	}
	sort := r.URL.Query().Get("sort")
	if sort != "" && !allowedTagNames[sort] {
		return clientErrorf("requested sort tag does not exist: %s", sort)
	}
	queryPart1 := "SELECT "
	queryPart2s := []string{}
	for _, tag := range includedTags {
		queryPart2s = append(queryPart2s, quoteDouble(tag))
	}
	queryPart3 := " FROM song WHERE "
	queryPart4s := []string{}
	for key, val := range tagFilters {
		queryPart4s = append(queryPart4s, fmt.Sprintf("%s = %s", quoteDouble(key), quoteSingle(val)))
	}
	queryPart5 := ""
	if sort != "" {
		queryPart5 = fmt.Sprintf(" ORDER BY %s COLLATE NOCASE ASC", quoteDouble(sort))
	}
	query := queryPart1 + strings.Join(queryPart2s, ", ") + queryPart3 + strings.Join(queryPart4s, ", ") + queryPart5
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
