package theme

import (
	"fmt"
	"path/filepath"
)

// Manager loads and caches themes from a theme root directory.
type Manager struct {
	ThemeRoot string // absolute path to the "themes" directory
	cache     map[string]*Theme
}

// NewManager constructs a Manager that loads themes from themeRoot.
func NewManager(themeRoot string) *Manager {
	absRoot, err := filepath.Abs(themeRoot)
	if err != nil {
		absRoot = themeRoot
	}
	return &Manager{
		ThemeRoot: absRoot,
		cache:     make(map[string]*Theme),
	}
}

// Load resolves and loads the first non-empty theme name from the provided list,
// applying the same precedence as Resolve (e.g. stage > campaign > global).
// The resolved theme is cached so subsequent calls with the same name are free.
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

	m.cache[name] = t
	return t, nil
}
