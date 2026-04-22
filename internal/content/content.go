package content

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Document struct {
	SourcePath   string
	RelPath      string
	RoutePath    string
	Title        string
	RawMarkdown  string
	HTML         string
	PlainText    string
	Sections     []Section
	LastModified time.Time
}

type Section struct {
	Title  string
	Anchor string
}

func Scan(docsDir string) ([]Document, error) {
	var docs []Document

	err := filepath.WalkDir(docsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		rel, err := filepath.Rel(docsDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		info, err := os.Stat(path)
		if err != nil {
			return err
		}

		docs = append(docs, Document{
			SourcePath:   path,
			RelPath:      rel,
			RoutePath:    toRoute(rel),
			RawMarkdown:  string(raw),
			LastModified: info.ModTime(),
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(docs, func(i, j int) bool {
		return docs[i].RoutePath < docs[j].RoutePath
	})

	return docs, nil
}

func toRoute(rel string) string {
	lower := strings.ToLower(rel)
	if lower == "readme.md" || lower == "index.md" {
		return "/"
	}

	base := strings.TrimSuffix(rel, filepath.Ext(rel))
	base = strings.TrimSuffix(base, "/README")
	base = strings.TrimSuffix(base, "/readme")
	base = filepath.ToSlash(base)
	if base == "" {
		return "/"
	}
	return "/" + strings.Trim(base, "/") + "/"
}
