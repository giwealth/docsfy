package site

import (
	"html/template"
	"path/filepath"
	"sort"
	"strings"

	"github.com/giwealth/docsfy/internal/content"
)

type NavItem struct {
	Title    string
	Route    string
	Children []NavItem
}

type PageData struct {
	SiteTitle string
	PageTitle string
	Current   string
	Content   template.HTML
	Nav       []NavItem
}

func ParseTemplates(templatesDir string) (*template.Template, error) {
	funcMap := template.FuncMap{
		"isActivePrefix": isActivePrefix,
	}
	return template.New("").Funcs(funcMap).ParseGlob(filepath.Join(templatesDir, "*.tmpl"))
}

func isActivePrefix(current, route string) bool {
	if route == "/" {
		return current == "/"
	}
	return strings.HasPrefix(current, route)
}

func BuildNav(docs []content.Document) []NavItem {
	root := &dirNode{name: "", children: map[string]*dirNode{}}
	for i := range docs {
		doc := docs[i]
		parts := strings.Split(filepath.ToSlash(doc.RelPath), "/")
		curr := root
		for _, seg := range parts[:len(parts)-1] {
			if seg == "" || seg == "." {
				continue
			}
			if curr.children[seg] == nil {
				curr.children[seg] = &dirNode{name: seg, children: map[string]*dirNode{}}
			}
			curr = curr.children[seg]
		}
		curr.docs = append(curr.docs, doc)
	}
	return buildDirNav(root, true)
}

type dirNode struct {
	name     string
	docs     []content.Document
	children map[string]*dirNode
}

func buildDirNav(node *dirNode, isRoot bool) []NavItem {
	sortDocs(node.docs)
	subDirs := sortedSubDirs(node)

	mainDoc, hasMain := pickMainDoc(node.docs)
	otherDocs := excludeMainDoc(node.docs, mainDoc, hasMain)

	subMenus := make([]NavItem, 0)
	for _, sub := range subDirs {
		subMenus = append(subMenus, buildDirNav(sub, false)...)
	}

	if isRoot {
		nav := make([]NavItem, 0)
		if hasMain {
			nav = append(nav, docToNav(mainDoc))
		}
		nav = append(nav, docsToNav(otherDocs)...)
		nav = append(nav, subMenus...)
		return nav
	}

	if hasMain {
		main := docToNav(mainDoc)
		main.Children = append(main.Children, docsToNav(otherDocs)...)
		main.Children = append(main.Children, subMenus...)
		return []NavItem{main}
	}

	nav := docsToNav(otherDocs)
	nav = append(nav, subMenus...)
	return nav
}

func sortDocs(docs []content.Document) {
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].RelPath < docs[j].RelPath
	})
}

func sortedSubDirs(node *dirNode) []*dirNode {
	keys := make([]string, 0, len(node.children))
	for k := range node.children {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]*dirNode, 0, len(keys))
	for _, k := range keys {
		out = append(out, node.children[k])
	}
	return out
}

func pickMainDoc(docs []content.Document) (content.Document, bool) {
	var indexDoc content.Document
	hasIndex := false
	for _, d := range docs {
		base := strings.ToLower(filepath.Base(d.RelPath))
		if base == "readme.md" {
			return d, true
		}
		if base == "index.md" {
			indexDoc = d
			hasIndex = true
		}
	}
	if hasIndex {
		return indexDoc, true
	}
	return content.Document{}, false
}

func excludeMainDoc(docs []content.Document, main content.Document, hasMain bool) []content.Document {
	if !hasMain {
		return docs
	}
	out := make([]content.Document, 0, len(docs)-1)
	for _, d := range docs {
		if d.RelPath == main.RelPath {
			continue
		}
		out = append(out, d)
	}
	return out
}

func docsToNav(docs []content.Document) []NavItem {
	out := make([]NavItem, 0, len(docs))
	for _, d := range docs {
		out = append(out, docToNav(d))
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Route < out[j].Route
	})
	return out
}

func docToNav(d content.Document) NavItem {
	title := d.Title
	if title == "" {
		title = strings.TrimSuffix(filepath.Base(d.RelPath), filepath.Ext(d.RelPath))
	}
	item := NavItem{
		Title: title,
		Route: d.RoutePath,
	}
	item.Children = buildSectionNav(d.RoutePath, d.Sections)
	return item
}

func buildSectionNav(baseRoute string, sections []content.Section) []NavItem {
	out := make([]NavItem, 0)
	lastH2 := -1
	for _, sec := range sections {
		n := NavItem{
			Title: sec.Title,
			Route: baseRoute + "#" + sec.Anchor,
		}
		switch sec.Level {
		case 2:
			out = append(out, n)
			lastH2 = len(out) - 1
		case 3:
			if lastH2 >= 0 {
				out[lastH2].Children = append(out[lastH2].Children, n)
			} else {
				// Fallback: when there is no preceding h2, keep h3 visible.
				out = append(out, n)
			}
		}
	}
	return out
}
