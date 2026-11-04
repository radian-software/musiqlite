package main

import (
	"context"
	"fmt"
	"strings"
)

func quoteSingle(text string) string {
	return `'` + strings.ReplaceAll(text, `'`, `''`) + `'`
}

func quoteDouble(text string) string {
	return `"` + strings.ReplaceAll(text, `"`, `""`) + `"`
}

func (s *Server) getAllowedTagNames(ctx context.Context) (map[string]bool, error) {
	if s.cacheData.AllowedTagNames == nil {
		rows, err := s.DB.QueryContext(ctx, "SELECT name FROM pragma_table_info('song')")
		if err != nil {
			return nil, fmt.Errorf("query db for listing song tag names: %w", err)
		}
		allowedTagNames := map[string]bool{}
		for rows.Next() {
			var name string
			err := rows.Scan(&name)
			if err != nil {
				return nil, fmt.Errorf("scan from db for listing song tag names: %w", err)
			}
			allowedTagNames[name] = true
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("scan from db for listing song tag names: %w", err)
		}
		s.cacheData.AllowedTagNames = allowedTagNames
	}
	return s.cacheData.AllowedTagNames, nil
}
