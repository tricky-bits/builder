package markdown

import (
	"bytes"
	"html/template"
	"io"
	"os"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/frontmatter"
)

var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM, &frontmatter.Extender{}),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		html.WithUnsafe(),
	),
)

// Parse opens a file, parses it using goldmark.Convert, and retrieves
// frontmatter using the goldmark-frontmatter package into fm.
func Parse(filename string, writer io.Writer, fm any) error {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	// Create parser context
	ctx := parser.NewContext()

	// Convert markdown to HTML and write to writer
	if err := md.Convert(content, writer, parser.WithContext(ctx)); err != nil {
		return err
	}

	// Extract frontmatter from context
	if fmData := frontmatter.Get(ctx); fmData != nil && fm != nil {
		// Decode frontmatter into the provided struct
		if err := fmData.Decode(fm); err != nil {
			return err
		}
	}

	return nil
}

// Render converts a markdown string to HTML and returns it as template.HTML.
func Render(content string) (template.HTML, error) {
	var buf bytes.Buffer
	if err := md.Convert([]byte(content), &buf); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}
