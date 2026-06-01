package markdown

import (
	"html/template"
	"strings"
)

// RenderHeroTitle renders the hero title source markdown, strips the
// surrounding paragraph wrapper produced by goldmark, and rewrites any
// <strong>...</strong> spans into <span class="tb-gradient">...</span>
// so authors can mark accent words with **bold** in IndexTitle.
func RenderHeroTitle(src string) (template.HTML, error) {
	if strings.TrimSpace(src) == "" {
		return "", nil
	}
	rendered, err := Render(src)
	if err != nil {
		return "", err
	}
	s := strings.TrimSpace(string(rendered))
	s = strings.TrimPrefix(s, "<p>")
	s = strings.TrimSuffix(s, "</p>")
	s = strings.ReplaceAll(s, "<strong>", `<span class="tb-gradient">`)
	s = strings.ReplaceAll(s, "</strong>", "</span>")
	return template.HTML(s), nil
}
