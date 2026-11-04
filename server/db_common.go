package main

import "strings"

func quoteSingle(text string) string {
	return `'` + strings.ReplaceAll(text, `'`, `''`) + `'`
}

func quoteDouble(text string) string {
	return `"` + strings.ReplaceAll(text, `"`, `""`) + `"`
}
