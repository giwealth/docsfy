package search

import (
	"encoding/json"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/giwealth/docsfy/internal/content"
)

type Item struct {
	Title   string `json:"title"`
	Route   string `json:"route"`
	Content string `json:"content"`
}

func BuildIndex(docs []content.Document) []Item {
	items := make([]Item, 0, len(docs))
	for _, d := range docs {
		title := d.Title
		if title == "" {
			title = filepath.Base(d.RelPath)
		}
		body := d.PlainText
		if utf8.RuneCountInString(body) > 160 {
			body = truncateRunes(body, 160)
		}
		items = append(items, Item{
			Title:   title,
			Route:   d.RoutePath,
			Content: body,
		})
	}
	return items
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	count := 0
	for i := range s {
		if count == max {
			return s[:i]
		}
		count++
	}
	return s
}

func Write(path string, items []Item) error {
	b, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
