package theme

import (
	"fmt"
	"path/filepath"
)

// Manager loads and caches themes from a theme root directory.
// When OutputDir is set, static assets are automatically copied to the output
// directory the first time each theme is loaded.
type Manager struct {
	ThemeRoot string // absolute path to the "themes" directory
	OutputDir string // destination for asset copies; empty disables auto-copy
	cache     map[string]*Theme
}

// NewManager constructs a Manager that loads themes from themeRoot and copies
// their static assets to outputDir on first use. Pass an empty outputDir to
// disable automatic asset copying (e.g. during validation-only flows).
func NewManager(themeRoot, outputDir string) *Manager {
	absRoot, err := filepath.Abs(themeRoot)
	if err != nil {
		absRoot = themeRoot
	}
	return &Manager{
		ThemeRoot: absRoot,
		OutputDir: outputDir,
		cache:     make(map[string]*Theme),
	}
}

// Load resolves and loads the first non-empty theme name from the provided list,
// applying the same precedence (e.g. stage > campaign > global).
// The resolved theme is cached so subsequent calls with the same name are free.
// If OutputDir is set, static assets are copied to the output directory on
// first load.
func (m *Manager) Load(themes ...string) (*Theme, error) {
	name := ""
	for _, candidate := range themes {
		if candidate != "" {
			name = candidate
			break
		}
	}
	if name == "" {
		return nil, fmt.Errorf("no theme configured")
	}

	if t, ok := m.cache[name]; ok {
		return t, nil
	}

	t, err := LoadFromPath(name, m.ThemeRoot)
	if err != nil {
		return nil, fmt.Errorf("load theme %q: %w", name, err)
	}

	if m.OutputDir != "" {
		if err := t.CopyAssets(m.OutputDir); err != nil {
			return nil, fmt.Errorf("copy assets for theme %q: %w", name, err)
		}
	}

	m.cache[name] = t
	return t, nil
}
