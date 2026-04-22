package render

import (
	"bytes"
	stdhtml "html"
	"regexp"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldhtml "github.com/yuin/goldmark/renderer/html"
)

var htmlTagRE = regexp.MustCompile(`<[^>]*>`)
var headingRE = regexp.MustCompile(`(?s)<h([2-3]) id="([^"]+)".*?>(.*?)</h[2-3]>`)

type Result struct {
	HTML      string
	PlainText string
	Title     string
	Sections  []Section
}

type Section struct {
	Title  string
	Anchor string
	Level  int
}

func MarkdownToHTML(markdown string) (Result, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			meta.Meta,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			goldhtml.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		return Result{}, err
	}
	htmlStr := buf.String()

	title := firstHeading(markdown)
	plain := strings.TrimSpace(htmlTagRE.ReplaceAllString(htmlStr, " "))
	plain = strings.Join(strings.Fields(plain), " ")
	return Result{
		HTML:      htmlStr,
		PlainText: plain,
		Title:     title,
		Sections:  extractSections(htmlStr),
	}, nil
}

func firstHeading(markdown string) string {
	lines := strings.Split(markdown, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

func extractSections(htmlStr string) []Section {
	matches := headingRE.FindAllStringSubmatch(htmlStr, -1)
	sections := make([]Section, 0, len(matches))
	for _, m := range matches {
		if len(m) < 4 {
			continue
		}
		anchor := strings.TrimSpace(m[2])
		title := strings.TrimSpace(htmlTagRE.ReplaceAllString(m[3], ""))
		title = stdhtml.UnescapeString(title)
		level, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		if anchor == "" || title == "" {
			continue
		}
		sections = append(sections, Section{
			Title:  title,
			Anchor: anchor,
			Level:  level,
		})
	}
	return sections
}
