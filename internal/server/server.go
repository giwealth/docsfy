package server

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/giwealth/docsfy/internal/config"
	"github.com/giwealth/docsfy/internal/content"
	"github.com/giwealth/docsfy/internal/kroki"
	"github.com/giwealth/docsfy/internal/render"
	"github.com/giwealth/docsfy/internal/search"
	"github.com/giwealth/docsfy/internal/site"
)

type state struct {
	mu         sync.RWMutex
	cfg        config.Config
	tmpl       *template.Template
	docs       []content.Document
	nav        []site.NavItem
	index      []search.Item
	clients    map[chan string]struct{}
	httpClient *http.Client
}

func Run(cfg config.Config) error {
	s := &state{
		cfg:        cfg,
		clients:    map[chan string]struct{}{},
		httpClient: &http.Client{Timeout: 90 * time.Second},
	}
	if err := s.rebuildAll(); err != nil {
		return err
	}
	go s.watch()

	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(cfg.AssetsDir))))
	mux.HandleFunc("/search-index.json", s.handleSearchIndex)
	mux.HandleFunc("/__livereload", s.handleLiveReload)
	mux.HandleFunc("/", s.handlePage)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("serve on http://localhost%s", addr)
	return http.ListenAndServe(addr, mux)
}

func (s *state) rebuildAll() error {
	docs, err := content.Scan(s.cfg.DocsDir)
	if err != nil {
		return err
	}
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
		embedded, err := kroki.Embed(ctx, docs[i].HTML, s.cfg.KrokiURL, s.cfg.KrokiDisabled, s.httpClient)
		if err != nil {
			return err
		}
		docs[i].HTML = embedded
	}
	tmpl, err := site.ParseTemplates(s.cfg.TemplatesDir)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.docs = docs
	s.nav = site.BuildNav(docs)
	s.index = search.BuildIndex(docs)
	s.tmpl = tmpl
	s.mu.Unlock()
	return nil
}

func (s *state) watch() {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("watcher init failed: %v", err)
		return
	}
	defer w.Close()

	_ = addWatchRecursive(w, s.cfg.DocsDir)
	_ = addWatchRecursive(w, s.cfg.TemplatesDir)
	_ = addWatchRecursive(w, s.cfg.AssetsDir)

	debounce := time.NewTimer(time.Hour)
	debounce.Stop()
	pending := false

	for {
		select {
		case <-debounce.C:
			pending = false
			if err := s.rebuildAll(); err != nil {
				log.Printf("rebuild failed: %v", err)
				continue
			}
			s.broadcast("reload")
		case evt, ok := <-w.Events:
			if !ok {
				return
			}
			if evt.Op&(fsnotify.Create) != 0 {
				if info, err := os.Stat(evt.Name); err == nil && info.IsDir() {
					_ = addWatchRecursive(w, evt.Name)
				}
			}
			if !pending {
				debounce.Reset(300 * time.Millisecond)
				pending = true
			}
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			log.Printf("watch error: %v", err)
		}
	}
}

func addWatchRecursive(w *fsnotify.Watcher, root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return w.Add(path)
		}
		return nil
	})
}

func (s *state) handleSearchIndex(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(s.index); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *state) handleLiveReload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "stream unsupported", http.StatusInternalServerError)
		return
	}
	ch := make(chan string, 8)
	s.mu.Lock()
	s.clients[ch] = struct{}{}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.clients, ch)
		s.mu.Unlock()
	}()

	for {
		select {
		case <-r.Context().Done():
			return
		case msg := <-ch:
			if _, err := fmt.Fprintf(w, "data: %s\n\n", msg); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (s *state) broadcast(msg string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for ch := range s.clients {
		select {
		case ch <- msg:
		default:
		}
	}
}

func (s *state) handlePage(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	if path == "//" {
		path = "/"
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	var doc *content.Document
	for i := range s.docs {
		if s.docs[i].RoutePath == path {
			doc = &s.docs[i]
			break
		}
	}
	if doc == nil {
		http.NotFound(w, r)
		return
	}

	data := site.PageData{
		SiteTitle: s.cfg.SiteTitle,
		PageTitle: doc.Title,
		Current:   doc.RoutePath,
		Content:   template.HTML(doc.HTML),
		Nav:       s.nav,
	}

	if err := s.tmpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
