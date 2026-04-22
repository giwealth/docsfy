// Package kroki renders fenced code blocks as diagrams via a Kroki-compatible HTTP API.
package kroki

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// DefaultURL is the public Kroki instance.
const DefaultURL = "https://kroki.io"

// diagramTypes lists Kroki diagram_type values (see https://kroki.io/).
var diagramTypes = map[string]struct{}{
	"actdiag": {}, "blockdiag": {}, "bpmn": {}, "bytefield": {}, "c4plantuml": {},
	"d2": {}, "dbml": {}, "ditaa": {}, "erd": {}, "excalidraw": {}, "graphviz": {},
	"mermaid": {}, "nomnoml": {}, "nwdiag": {}, "packetdiag": {}, "pikchr": {},
	"plantuml": {}, "rackdiag": {}, "seqdiag": {}, "structurizr": {}, "svgbob": {},
	"symbolator": {}, "tikz": {}, "vega": {}, "vegalite": {}, "wavedrom": {}, "wireviz": {},
}

var typeAliases = map[string]string{
	"vega-lite": "vegalite",
	"dot":       "graphviz",
	"c4":        "c4plantuml",
}

var svgCache sync.Map // key: hex(sha256), value: []byte

// NormalizeType maps fence language / alias to Kroki diagram_type.
func NormalizeType(lang string) (string, bool) {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang == "" {
		return "", false
	}
	if a, ok := typeAliases[lang]; ok {
		lang = a
	}
	if _, ok := diagramTypes[lang]; ok {
		return lang, true
	}
	return "", false
}

// TypeFromCodeClass parses goldmark "language-xxx" class and returns Kroki type.
func TypeFromCodeClass(class string) (string, bool) {
	for _, p := range strings.Fields(class) {
		if strings.HasPrefix(p, "language-") {
			raw := strings.TrimPrefix(p, "language-")
			return NormalizeType(raw)
		}
	}
	return "", false
}

type postBody struct {
	DiagramSource string `json:"diagram_source"`
	DiagramType   string `json:"diagram_type"`
	OutputFormat  string `json:"output_format"`
}

// FetchSVG calls Kroki POST / and returns raw SVG bytes.
func FetchSVG(ctx context.Context, baseURL, diagramType, source string, client *http.Client) ([]byte, error) {
	if client == nil {
		client = http.DefaultClient
	}
	base := strings.TrimSuffix(strings.TrimSpace(baseURL), "/")
	if base == "" {
		return nil, fmt.Errorf("kroki: empty base URL")
	}
	u := base + "/"
	body, err := json.Marshal(postBody{
		DiagramSource: source,
		DiagramType:   diagramType,
		OutputFormat:  "svg",
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "image/svg+xml")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kroki: %s: %s", resp.Status, truncate(string(b), 200))
	}
	return b, nil
}

func fetchSVGCached(ctx context.Context, baseURL, diagramType, source string, client *http.Client) ([]byte, error) {
	h := sha256.Sum256([]byte(diagramType + "\x00" + source))
	key := hex.EncodeToString(h[:])
	if v, ok := svgCache.Load(key); ok {
		if b, ok2 := v.([]byte); ok2 {
			return b, nil
		}
	}
	b, err := FetchSVG(ctx, baseURL, diagramType, source, client)
	if err != nil {
		return nil, err
	}
	svgCache.Store(key, b)
	return b, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// Embed replaces supported <pre><code class="language-xxx"> blocks in an HTML fragment with Kroki output.
func Embed(ctx context.Context, htmlFragment, baseURL string, krokiDisabled bool, client *http.Client) (string, error) {
	if krokiDisabled || strings.TrimSpace(baseURL) == "" {
		return htmlFragment, nil
	}
	if client == nil {
		client = &http.Client{Timeout: 90 * time.Second}
	}

	wrapped := `<!DOCTYPE html><html><head><meta charset="utf-8"></head><body><div id="docsfy-kroki-root">` +
		htmlFragment + `</div></body></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(wrapped))
	if err != nil {
		return "", err
	}
	root := doc.Find("#docsfy-kroki-root").First()
	if root.Length() == 0 {
		return htmlFragment, nil
	}

	root.Find("pre").Each(func(_ int, pre *goquery.Selection) {
		code := pre.Find("code").First()
		if code.Length() == 0 {
			return
		}
		cls, ok := code.Attr("class")
		if !ok {
			return
		}
		diagramType, ok := TypeFromCodeClass(cls)
		if !ok {
			return
		}
		src := code.Text()
		if strings.TrimSpace(src) == "" {
			return
		}

		svg, err := fetchSVGCached(ctx, baseURL, diagramType, src, client)
		if err != nil {
			log.Printf("kroki [%s]: %v", diagramType, err)
			return
		}
		svg = bytes.TrimSpace(svg)
		if bytes.HasPrefix(svg, []byte("<?xml")) {
			if i := bytes.IndexByte(svg, '>'); i >= 0 {
				svg = bytes.TrimSpace(svg[i+1:])
			}
		}
		b64 := base64.StdEncoding.EncodeToString(svg)
		frag := fmt.Sprintf(
			`<div class="kroki-diagram" data-kroki-type="%s"><img src="data:image/svg+xml;base64,%s" alt="%s diagram" loading="lazy"/></div>`,
			diagramType, b64, diagramType,
		)
		pre.ReplaceWithHtml(frag)
	})

	out, err := root.Html()
	if err != nil {
		return "", err
	}
	return out, nil
}
