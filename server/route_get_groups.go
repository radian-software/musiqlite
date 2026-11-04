package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

func (s *Server) handleGetGroups(w http.ResponseWriter, r *http.Request) error {
	partition := r.URL.Query().Get("partition")
	if partition == "" {
		return clientError(errors.New("missing query parameter: partition"), http.StatusBadRequest)
	}
	includedTags := r.URL.Query()["tag"]
	for _, tag := range includedTags {
		if tag == "" {
			return clientError(errors.New("empty tag name"), http.StatusBadRequest)
		}
	}
	sort := r.URL.Query().Get("sort")
	rows, err := s.DB.QueryContext(r.Context(), fmt.Sprintf("SELECT %[1]s, COUNT(*) FROM song GROUP BY %[1]s", quoteDouble(partition)))
	if err != nil {
		return fmt.Errorf("query db for counting songs by partition: %w", err)
	}
	songCounts := map[string]int{}
	for rows.Next() {
		var name string
		var count int
		err := rows.Scan(&name, &count)
		if err != nil {
			return fmt.Errorf("scan from db for counting songs by partition: %w", err)
		}
		songCounts[name] = count
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("scan from db for counting songs by partition: %w", err)
	}
	allowedTagNames, err := s.getAllowedTagNames(r.Context())
	if err != nil {
		return err
	}
	for _, tagName := range append(includedTags, partition, sort) {
		if tagName != "" && !allowedTagNames[tagName] {
			return clientError(fmt.Errorf("requested tag does not exist: %s", partition), http.StatusUnprocessableEntity)
		}
	}
	groups := []any{}
	if len(includedTags) > 0 {
		queryPart1 := "WITH cte AS (SELECT *, "
		queryPart2s := []string{}
		for _, name := range includedTags {
			queryPart2s = append(queryPart2s, fmt.Sprintf("COUNT(*) OVER (PARTITION BY %s, %s) AS %s", quoteDouble(partition), quoteDouble(name), quoteDouble("_"+name+"_count")))
		}
		queryPart3 := fmt.Sprintf(" FROM song) SELECT DISTINCT %s, ", quoteDouble(partition))
		queryPart4s := []string{}
		for _, name := range includedTags {
			queryPart4s = append(queryPart4s, fmt.Sprintf("FIRST_VALUE(%[1]s) OVER (PARTITION BY %[2]s ORDER BY %[3]s DESC), %[3]s", quoteDouble(name), quoteDouble(partition), quoteDouble("_"+name+"_count")))
		}
		queryPart5 := " FROM cte"
		queryPart6 := ""
		if sort != "" {
			queryPart6 = fmt.Sprintf(" ORDER BY %s COLLATE NOCASE ASC", quoteDouble(sort))
		}
		query := queryPart1 + strings.Join(queryPart2s, ", ") + queryPart3 + strings.Join(queryPart4s, ", ") + queryPart5 + queryPart6
		rows, err = s.DB.QueryContext(r.Context(), query)
		if err != nil {
			return fmt.Errorf("query db for counting tags within partitions: %w", err)
		}
		for rows.Next() {
			var partitionKey string
			tagValues := make([]any, len(includedTags))
			tagCounts := make([]int, len(includedTags))
			cols := make([]any, 2*len(includedTags)+1)
			cols[0] = &partitionKey
			for idx := range includedTags {
				cols[2*idx+1] = &tagValues[idx]
				cols[2*idx+2] = &tagCounts[idx]
			}
			err := rows.Scan(cols...)
			if err != nil {
				return fmt.Errorf("scan from db for counting tags within partitions: %w", err)
			}
			partitionTotal := songCounts[partitionKey]
			tagsInfo := map[string]any{}
			for idx, name := range includedTags {
				tagsInfo[name] = map[string]any{
					"most_common_value": tagValues[idx],
					"most_common_count": tagCounts[idx],
				}
			}
			groups = append(groups, map[string]any{
				"partition_tag_value": partitionKey,
				"song_count":          partitionTotal,
				"tag_values":          tagsInfo,
			})
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("scan from db for counting tags within partitions: %w", err)
		}
	} else {
		for partitionKey, partitionTotal := range songCounts {
			groups = append(groups, map[string]any{
				"partition_tag_value": partitionKey,
				"song_count":          partitionTotal,
				"tag_values":          map[string]any{},
			})
		}
	}
	json.NewEncoder(w).Encode(groups)
	return nil
}
