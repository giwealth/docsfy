package build

import (
	"bytes"
	"context"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/giwealth/docsfy/internal/config"
	"github.com/giwealth/docsfy/internal/content"
	"github.com/giwealth/docsfy/internal/kroki"
	"github.com/giwealth/docsfy/internal/render"
	"github.com/giwealth/docsfy/internal/search"
	"github.com/giwealth/docsfy/internal/site"
)

func Run(cfg config.Config) error {
	docs, err := content.Scan(cfg.DocsDir)
	if err != nil {
		return err
	}

	httpClient := &http.Client{Timeout: 90 * time.Second}
	ctx := context.Background()
	for i := range docs {
		r, err := render.MarkdownToHTML(docs[i].RawMarkdown)
		if err != nil {
			return err
		}
		docs[i].HTML = r.HTML
		docs[i].PlainText = r.PlainText
		docs[i].Title = r.Title
		docs[i].Sections = make([]content.Section, 0, len(r.Sections))
		for _, sec := range r.Sections {
			docs[i].Sections = append(docs[i].Sections, content.Section{
				Title:  sec.Title,
				Anchor: sec.Anchor,
			})
		}
		embedded, err := kroki.Embed(ctx, docs[i].HTML, cfg.KrokiURL, cfg.KrokiDisabled, cfg.FrontendMermaid, httpClient)
		if err != nil {
			return err
		}
		docs[i].HTML = embedded
	}

	if err := os.RemoveAll(cfg.OutDir); err != nil {
		return err
	}
	if err := os.MkdirAll(cfg.OutDir, 0o755); err != nil {
		return err
	}

	tmpl, err := site.ParseTemplates(cfg.TemplatesDir)
	if err != nil {
		return err
	}
	nav := site.BuildNav(docs)
	for _, d := range docs {
		data := site.PageData{
			SiteTitle: cfg.SiteTitle,
			PageTitle: d.Title,
			Current:   d.RoutePath,
			Content:   template.HTML(d.HTML),
			Nav:       nav,
		}
		var out bytes.Buffer
		if err := tmpl.ExecuteTemplate(&out, "layout.tmpl", data); err != nil {
			return err
		}

		target := routeToOutPath(cfg.OutDir, d.RoutePath)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, out.Bytes(), 0o644); err != nil {
			return err
		}
	}

	if err := copyDir(cfg.AssetsDir, filepath.Join(cfg.OutDir, "assets")); err != nil {
		return err
	}

	index := search.BuildIndex(docs)
	return search.Write(filepath.Join(cfg.OutDir, "search-index.json"), index)
}

func routeToOutPath(outDir, route string) string {
	if route == "/" {
		return filepath.Join(outDir, "index.html")
	}
	return filepath.Join(outDir, route, "index.html")
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
}
