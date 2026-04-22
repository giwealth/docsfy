package render

import (
	"strings"
	"testing"
)

func TestMermaidFenceClass(t *testing.T) {
	md := "```mermaid\nflowchart LR\nA-->B\n```"
	r, err := MarkdownToHTML(md)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.HTML, "language-mermaid") && !strings.Contains(r.HTML, "mermaid") {
		t.Fatalf("unexpected HTML: %s", r.HTML)
	}
	t.Log(r.HTML)
}
