package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func queryParamBool(val string) (bool, error) {
	switch val {
	case "", "0", "n", "no", "off":
		return false, nil
	case "1", "y", "yes", "on":
		return true, nil
	}
	return false, fmt.Errorf("not a valid boolean: %s", val)
}

func (s *Server) handlePutPlaylist(w http.ResponseWriter, r *http.Request) error {
	name := r.URL.Query().Get("name")
	if name == "" {
		return clientErrorf("no playlist name specified")
	}
	isUpNext, err := queryParamBool(r.URL.Query().Get("is_upnext"))
	if err != nil {
		return clientErrorf("is_upnext: %w", err)
	}
	oidStrings := r.URL.Query()["song"]
	oids := []int64{}
	for _, oidString := range oidStrings {
		oid, err := strconv.ParseInt(oidString, 10, 64)
		if err != nil {
			return clientErrorf("not a valid song oid: %s", oidString)
		}
		oids = append(oids, oid)
	}
	uniqueOIDs := map[int64]bool{}
	for _, oid := range oids {
		uniqueOIDs[oid] = true
	}
	txn, err := s.DB.BeginTx(r.Context(), nil)
	if err != nil {
		return err
	}
	defer txn.Rollback()
	queryPart1 := `SELECT COUNT(*) FROM song WHERE oid IN (`
	queryPart2s := []string{}
	for oid := range uniqueOIDs {
		queryPart2s = append(queryPart2s, fmt.Sprintf("%d", oid))
	}
	queryPart3 := `)`
	query := queryPart1 + strings.Join(queryPart2s, ", ") + queryPart3
	row := txn.QueryRowContext(r.Context(), query)
	numExisting := -1
	err = row.Scan(&numExisting)
	if err != nil {
		return fmt.Errorf("verifying songs for new playlist: %w", err)
	}
	if numExisting > len(uniqueOIDs) {
		return fmt.Errorf("oid query for new playlist somehow returned duplicates")
	}
	if numExisting < len(uniqueOIDs) {
		return clientErrorf("one or more songs for playlist do not exist")
	}
	_, err = txn.ExecContext(r.Context(), fmt.Sprintf(`DELETE FROM playlist_song WHERE playlist_oid = (SELECT oid FROM playlist WHERE name = %s)`, quoteSingle(name)))
	if err != nil {
		return fmt.Errorf("deleting songs in existing playlist: %w", err)
	}
	_, err = txn.ExecContext(r.Context(), fmt.Sprintf(`DELETE FROM playlist WHERE name = %s`, quoteSingle(name)))
	if err != nil {
		return fmt.Errorf("deleting existing playlist: %w", err)
	}
	playbackPosition := "NULL"
	if isUpNext {
		playbackPosition = "0"
	}
	row = txn.QueryRowContext(r.Context(), fmt.Sprintf(`INSERT INTO playlist (name, is_upnext_playlist, playback_position) VALUES (%s, %t, %s) RETURNING oid`, quoteSingle(name), isUpNext, playbackPosition))
	newOID := -1
	err = row.Scan(&newOID)
	if err != nil {
		return fmt.Errorf("creating playlist: %w", err)
	}
	if len(oids) > 0 {
		queryPart1 = `INSERT INTO playlist_song (playlist_oid, song_oid, ordering) VALUES `
		queryPart2s = []string{}
		for idx, oid := range oids {
			queryPart2s = append(queryPart2s, fmt.Sprintf(`(%d, %d, %d)`, newOID, oid, idx))
		}
		query = queryPart1 + strings.Join(queryPart2s, ", ")
		_, err = txn.ExecContext(r.Context(), query)
		if err != nil {
			return fmt.Errorf("adding songs to playlist: %w", err)
		}
	}
	err = txn.Commit()
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
