package markdown

import (
	"html"
	"html/template"
	"regexp"
	"strings"
)

// heroTagRE matches any HTML tag, used to reduce rendered hero markup to text.
var heroTagRE = regexp.MustCompile(`<[^>]+>`)

// HeroTitle renders hero title markdown once and returns both the gradient
// HTML form and the plain-text form. The HTML form strips the paragraph
// wrapper produced by goldmark and rewrites <strong>...</strong> spans into
// <span class="tb-gradient">...</span> so authors can mark accent words with
// **bold**. The text form removes all markup; use it wherever the title is
// shown as literal text (the <title> tag, breadcrumbs, JSON payloads) so the
// bold markers do not leak through. Returns zero values for blank input.
func HeroTitle(src string) (template.HTML, string, error) {
	if strings.TrimSpace(src) == "" {
		return "", "", nil
	}
	rendered, err := Render(src)
	if err != nil {
		return "", "", err
	}

	h := strings.TrimSpace(string(rendered))
	h = strings.TrimPrefix(h, "<p>")
	h = strings.TrimSuffix(h, "</p>")
	h = strings.ReplaceAll(h, "<strong>", `<span class="tb-gradient">`)
	h = strings.ReplaceAll(h, "</strong>", "</span>")

	text := heroTagRE.ReplaceAllString(string(rendered), "")
	text = html.UnescapeString(text)
	text = strings.TrimSpace(text)

	return template.HTML(h), text, nil
}
