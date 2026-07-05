// Package theme provides theme management functionality for loading HTML templates
// and static assets. Themes are organized in a directory structure where each theme has:
//
//   - templates/: HTML template files (*.tmpl) that define the structure and layout
//   - static/: Static assets (CSS, JS, images, etc.) that are copied to the output
//
// The Manager handles loading themes and caching them. Themes can be resolved at
// different levels (global, campaign, stage) with stage-level themes taking precedence.
package theme

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tricky-bits/builder/internal/helpers"
)

// Config is the theme portion of the global config.
type Config struct {
	// Name is the default theme name (e.g., "default").
	Name string `yaml:"name"`
}

// Theme represents a loaded theme: parsed templates and the path to static assets.
type Theme struct {
	Name string

	// TemplateSets contains one parsed template set per required entry template.
	// Each set bundles the entry template together with all helper templates.
	// Key is the entry template filename (e.g., "stage.html").
	TemplateSets map[string]*template.Template

	// StaticDir is the path to the theme's static asset directory
	// (themes/<name>/static). Empty string means no static assets.
	StaticDir string

	// FeaturedImageFallbacks is the sorted list of image filenames (any
	// FeaturedImageExtensions extension) discovered in <StaticDir>/fallbacks/,
	// used as campaign featured-image fallbacks.
	// Nil when the directory is missing or empty.
	FeaturedImageFallbacks []string

	// copied tracks whether CopyAssets has already run for this theme.
	copied bool
}

// Validate verifies the theme provides all required entry templates.
func (t *Theme) Validate() error {
	var missing []string
	for _, required := range requiredTemplates {
		tmplSet, ok := t.TemplateSets[required]
		if !ok || tmplSet.Lookup(required) == nil {
			missing = append(missing, required)
		}
	}
	if len(missing) > 0 {
		if t.Name == "" {
			return fmt.Errorf("theme (unnamed) missing required templates: %v", missing)
		}
		return fmt.Errorf("theme %q missing required templates: %v", t.Name, missing)
	}
	return nil
}

// CopyAssets copies the theme's static assets to <outputDir>/assets/<themeName>/.
// It is idempotent: subsequent calls for the same theme instance are no-ops.
func (t *Theme) CopyAssets(outputDir string) error {
	if t.copied || t.StaticDir == "" {
		return nil
	}

	targetDir := filepath.Join(outputDir, "assets", t.Name)

	err := filepath.Walk(t.StaticDir, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(t.StaticDir, srcPath)
		if err != nil {
			return fmt.Errorf("get relative path for %s: %w", srcPath, err)
		}

		dstPath := filepath.Join(targetDir, rel)

		if info.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
			return fmt.Errorf("create parent directory for %s: %w", dstPath, err)
		}

		return helpers.CopyFile(srcPath, dstPath)
	})
	if err != nil {
		return fmt.Errorf("copy assets for theme %s: %w", t.Name, err)
	}

	t.copied = true
	return nil
}

// requiredTemplates declares the entry templates the generator must be able to execute.
// This is intentionally unexported: external code must not modify this slice, as it
// is used during theme validation.
var requiredTemplates = []string{
	"home.html",
	"campaigns.html",
	"campaign.html",
	"stage.html",
	"404.html",
	"page.html",
}

// Render executes the named template against data, writing the result to w.
func (t *Theme) Render(w io.Writer, templateName string, data any) error {
	tmplSet, ok := t.TemplateSets[templateName]
	if !ok {
		return fmt.Errorf("theme %q: template %q not found", t.Name, templateName)
	}
	tmpl := tmplSet.Lookup(templateName)
	if tmpl == nil {
		return fmt.Errorf("theme %q: template %q could not be looked up", t.Name, templateName)
	}
	return tmpl.Execute(w, data)
}

// LoadFromPath loads and validates a theme from themesDir/name.
func LoadFromPath(name, themesDir string) (*Theme, error) {
	// Guard against path traversal: verify the resolved theme directory
	// is still rooted under themesDir after filepath.Join cleans the path.
	themeDir := filepath.Join(themesDir, name)
	rel, err := filepath.Rel(themesDir, themeDir)
	if err != nil || strings.HasPrefix(rel, "..") {
		return nil, fmt.Errorf("invalid theme name %q", name)
	}

	templatesDir := filepath.Join(themeDir, "templates")
	staticDir := filepath.Join(themeDir, "static")

	if _, err := os.Stat(templatesDir); err != nil {
		return nil, fmt.Errorf("templates directory not found: %s", templatesDir)
	}

	funcMap := template.FuncMap{
		"urlJoin": JoinURL,
		"add":     func(a, b int) int { return a + b },
		"json": func(v any) (template.JS, error) {
			b, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return template.JS(b), nil
		},
		"formatETA": func(minutes int) string {
			if minutes <= 0 {
				return ""
			}
			h := minutes / 60
			m := minutes % 60
			if h == 0 {
				return fmt.Sprintf("%dm", m)
			}
			if m == 0 {
				return fmt.Sprintf("%dh", h)
			}
			return fmt.Sprintf("%dh %dm", h, m)
		},
		"seq": func(n int) []int {
			if n <= 0 {
				return nil
			}
			out := make([]int, n)
			for i := range out {
				out[i] = i
			}
			return out
		},
	}

	// Collect all .tmpl files from the templates directory.
	var allFiles []string
	err = filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".html") {
			allFiles = append(allFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk templates directory: %w", err)
	}
	if len(allFiles) == 0 {
		return nil, fmt.Errorf("no templates found in %s", templatesDir)
	}

	// Separate required (entry) templates from helper templates.
	//
	// Unlike Jinja2's block inheritance, Go's html/template uses an additive
	// model: all templates parsed into the same set share a single namespace.
	// If two files both define a template named "content", the last one parsed
	// silently overrides the first — there is no error. To prevent helpers from
	// accidentally overriding blocks defined by the entry template, we give each
	// required template its own isolated template set. Helpers are parsed first
	// (establishing shared partials), then the required template is parsed last
	// so its definitions always win.
	requiredSet := make(map[string]bool, len(requiredTemplates))
	for _, r := range requiredTemplates {
		requiredSet[r] = true
	}

	var helperFiles, requiredFiles []string
	for _, f := range allFiles {
		if requiredSet[filepath.Base(f)] {
			requiredFiles = append(requiredFiles, f)
		} else {
			helperFiles = append(helperFiles, f)
		}
	}

	// Build one template set per required template.
	// Each set contains all helper templates (parsed first) plus the required
	// template (parsed last), ensuring the required template's block definitions
	// take precedence over any same-named definitions in helpers.
	templateSets := make(map[string]*template.Template, len(requiredFiles))
	for _, requiredFile := range requiredFiles {
		relPath, err := filepath.Rel(templatesDir, requiredFile)
		if err != nil {
			return nil, fmt.Errorf("get relative path for %s: %w", requiredFile, err)
		}
		templateName := filepath.ToSlash(relPath)

		tmpl := template.New("").Funcs(funcMap)

		for _, helperFile := range helperFiles {
			helperRel, err := filepath.Rel(templatesDir, helperFile)
			if err != nil {
				return nil, fmt.Errorf("get relative path for helper %s: %w", helperFile, err)
			}
			content, err := os.ReadFile(helperFile)
			if err != nil {
				return nil, fmt.Errorf("read helper template %s: %w", helperRel, err)
			}
			if _, err = tmpl.New(filepath.ToSlash(helperRel)).Parse(string(content)); err != nil {
				return nil, fmt.Errorf("parse helper template %s: %w", helperRel, err)
			}
		}

		content, err := os.ReadFile(requiredFile)
		if err != nil {
			return nil, fmt.Errorf("read required template %s: %w", templateName, err)
		}
		if _, err = tmpl.New(templateName).Parse(string(content)); err != nil {
			return nil, fmt.Errorf("parse required template %s: %w", templateName, err)
		}

		templateSets[filepath.Base(templateName)] = tmpl
	}

	// Resolve static directory: leave empty if it doesn't exist.
	if stat, err := os.Stat(staticDir); err != nil || !stat.IsDir() {
		staticDir = ""
	}

	fallbacks, err := discoverFallbacks(staticDir)
	if err != nil {
		return nil, err
	}

	t := &Theme{
		Name:                   name,
		TemplateSets:           templateSets,
		StaticDir:              staticDir,
		FeaturedImageFallbacks: fallbacks,
	}

	if err := t.Validate(); err != nil {
		return nil, err
	}

	return t, nil
}
