package builder

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/tricky-bits/builder/internal/helpers"
	"github.com/tricky-bits/builder/internal/markdown"
)

// PageFrontmatter represents the YAML frontmatter declared at the top of a page
// markdown file. It contains author-provided metadata consumed by the rendering
// pipeline (routing, theming, navigation, and publishing timestamps).
type PageFrontmatter struct {
	// Title is the human-readable page title (required).
	Title string `yaml:"title"`

	// Slug is the logical route identifier for the page (optional). If omitted,
	// it is derived from the filename.
	Slug string `yaml:"slug,omitempty"`

	// Theme optionally overrides the global theme for this page.
	Theme string `yaml:"theme,omitempty"`

	// MenuLabel, when set, indicates that the page should be included in the
	// generated navigation menu using this label.
	MenuLabel string `yaml:"menu_label,omitempty"`

	// MenuOrder controls the relative ordering of the page in navigation menus.
	// Lower values appear first.
	MenuOrder int `yaml:"menu_order,omitempty"`

	// MenuIcon is an optional icon identifier for the navigation menu entry.
	MenuIcon string `yaml:"menu_icon,omitempty"`

	// PublishedAt is the page publication timestamp (optional).
	PublishedAt *time.Time `yaml:"published_at,omitempty"`

	// LastUpdatedAt is the page last update timestamp (optional).
	LastUpdatedAt *time.Time `yaml:"last_updated_at,omitempty"`
}

// Page represents a parsed and rendered page produced from a markdown file.
type Page struct {
	// Filename is the source markdown file path (or name), used primarily for
	// traceability and error reporting.
	Filename string

	// Frontmatter contains the parsed YAML header for the page.
	Frontmatter PageFrontmatter

	// Content is the rendered HTML body of the page, intended to be injected into
	// a larger HTML template.
	Content template.HTML
}

// ReadAllPages discovers and loads all page markdown files under the page
// directory rooted at basepath. Returns an error if any file fails to parse or
// if two pages share the same slug.
func ReadAllPages(basepath string) (map[string]*Page, error) {
	filenames, err := helpers.ListFiles(basepath, ".md")
	if err != nil {
		return nil, fmt.Errorf("unable to list page files: %w", err)
	}

	pages := make(map[string]*Page)
	for _, filename := range filenames {
		page, err := ReadPageFile(filename)
		if err != nil {
			return nil, fmt.Errorf("unable to read page file %q: %w", filename, err)
		}

		if _, exists := pages[page.Frontmatter.Slug]; exists {
			return nil, fmt.Errorf("[%s] duplicate page slug: %s", filename, page.Frontmatter.Slug)
		}

		pages[page.Frontmatter.Slug] = page
	}

	return pages, nil
}

// ReadPageFile reads and parses a page markdown file from the given filename.
// It extracts the YAML frontmatter, renders the markdown content to HTML,
// derives the slug if not provided, and validates the page structure.
func ReadPageFile(filename string) (*Page, error) {
	page := &Page{
		Filename: filename,
	}

	var contentBuf bytes.Buffer
	if err := markdown.Parse(filename, &contentBuf, &page.Frontmatter); err != nil {
		return nil, fmt.Errorf("unable to parse markdown file: %w", err)
	}

	page.Content = template.HTML(contentBuf.String())

	// Derive slug from frontmatter or filename
	if page.Frontmatter.Slug == "" {
		page.Frontmatter.Slug = helpers.DeriveSlug(filename)
	}

	if err := page.Validate(); err != nil {
		return nil, err
	}

	return page, nil
}

// Validate checks whether the Page is structurally valid and ready to render.
func (p *Page) Validate() error {
	if p.Frontmatter.Title == "" {
		return fmt.Errorf("title is required")
	}

	if p.Frontmatter.Slug == "" {
		return fmt.Errorf("slug is required")
	}

	if p.Frontmatter.MenuOrder < 0 {
		return fmt.Errorf("menu_order must be non-negative, got %d", p.Frontmatter.MenuOrder)
	}

	return nil
}
