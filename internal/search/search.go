package search

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/giwealth/docsfy/internal/content"
)

type Item struct {
	Title   string `json:"title"`
	Route   string `json:"route"`
	Content string `json:"content"`
}

var headingRE = regexp.MustCompile(`^(##|###)\s+(.+)$`)
var mdNoiseRE = regexp.MustCompile("[`*_>#\\[\\]()]")

func BuildIndex(docs []content.Document) []Item {
	items := make([]Item, 0, len(docs)*4)
	for _, d := range docs {
		docTitle := d.Title
		if docTitle == "" {
			docTitle = filepath.Base(d.RelPath)
		}

		if len(d.Sections) == 0 {
			items = append(items, Item{
				Title:   docTitle,
				Route:   d.RoutePath,
				Content: shrinkSnippet(d.PlainText, 160),
			})
			continue
		}

		sectionBodies := extractSectionBodies(d.RawMarkdown)
		added := 0
		for i, sec := range d.Sections {
			route := d.RoutePath + "#" + sec.Anchor
			chunk := ""
			if i < len(sectionBodies) {
				chunk = sectionBodies[i]
			}
			paragraphs := splitParagraphs(chunk)
			if len(paragraphs) == 0 {
				paragraphs = []string{sec.Title}
			}

			for _, p := range paragraphs {
				snippet := shrinkSnippet(p, 160)
				if snippet == "" {
					continue
				}
				items = append(items, Item{
					Title:   docTitle + " / " + sec.Title,
					Route:   route,
					Content: snippet,
				})
				added++
			}
		}
		if added == 0 {
			items = append(items, Item{
				Title:   docTitle,
				Route:   d.RoutePath,
				Content: shrinkSnippet(d.PlainText, 160),
			})
		}
	}
	return items
}

func extractSectionBodies(md string) []string {
	lines := strings.Split(md, "\n")
	bodies := make([]string, 0)
	var current []string
	inSection := false

	for _, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		if headingRE.MatchString(strings.TrimSpace(line)) {
			if inSection {
				bodies = append(bodies, strings.Join(current, "\n"))
				current = current[:0]
			} else {
				inSection = true
			}
			continue
		}
		if inSection {
			current = append(current, line)
		}
	}
	if inSection {
		bodies = append(bodies, strings.Join(current, "\n"))
	}
	return bodies
}

func splitParagraphs(text string) []string {
	blocks := strings.Split(text, "\n\n")
	out := make([]string, 0, len(blocks))
	for _, b := range blocks {
		clean := shrinkSnippet(b, 0)
		if clean != "" {
			out = append(out, clean)
		}
	}
	return out
}

func shrinkSnippet(s string, max int) string {
	s = mdNoiseRE.ReplaceAllString(s, " ")
	s = strings.Join(strings.Fields(s), " ")
	if max > 0 && utf8.RuneCountInString(s) > max {
		s = truncateRunes(s, max)
	}
	return strings.TrimSpace(s)
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
