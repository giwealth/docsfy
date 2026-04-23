package webembed

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed templates/*.tmpl assets/* assets/vendor/*
var embeddedFiles embed.FS

func EmbeddedFiles() fs.FS {
	return embeddedFiles
}

func ParseEmbeddedTemplates() (*template.Template, error) {
	tfs, err := fs.Sub(embeddedFiles, "templates")
	if err != nil {
		return nil, err
	}
	return template.ParseFS(tfs, "*.tmpl")
}

func EmbeddedAssetsFS() (fs.FS, error) {
	return fs.Sub(embeddedFiles, "assets")
}

func CopyEmbeddedAssets(dst string) error {
	afs, err := EmbeddedAssetsFS()
	if err != nil {
		return err
	}
	return fs.WalkDir(afs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(dst, path)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := fs.ReadFile(afs, path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
}

func HasDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func ErrNoEmbeddedAssets() error {
	return fmt.Errorf("embedded assets unavailable")
}
