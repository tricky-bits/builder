package theme

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromPath_DiscoversFallbacks(t *testing.T) {
	themesDir, err := filepath.Abs("../../testdata/themes")
	require.NoError(t, err)

	th, err := LoadFromPath("fixture", themesDir)
	require.NoError(t, err)
	assert.Equal(t, []string{"fallback-a.png", "fallback-b.png", "fallback-c.png"}, th.FeaturedImageFallbacks)
}

func TestLoadFromPath_NoFallbacksDirYieldsEmptyList(t *testing.T) {
	// Build a fixture-like theme on the fly without a fallbacks dir.
	themesDir := t.TempDir()
	themeDir := filepath.Join(themesDir, "noinfo")
	tmplDir := filepath.Join(themeDir, "templates")
	staticDir := filepath.Join(themeDir, "static")
	require.NoError(t, os.MkdirAll(tmplDir, 0o755))
	require.NoError(t, os.MkdirAll(staticDir, 0o755))

	for _, name := range []string{"home.html", "campaign.html", "stage.html", "page.html", "404.html"} {
		require.NoError(t, os.WriteFile(filepath.Join(tmplDir, name), []byte(`{{define "title"}}x{{end}}{{define "content"}}x{{end}}`), 0o644))
	}

	th, err := LoadFromPath("noinfo", themesDir)
	require.NoError(t, err)
	assert.Empty(t, th.FeaturedImageFallbacks)
}

func TestLoadFromPath_EmptyFallbacksDirYieldsEmptyList(t *testing.T) {
	themesDir := t.TempDir()
	themeDir := filepath.Join(themesDir, "emptyfb")
	tmplDir := filepath.Join(themeDir, "templates")
	fbDir := filepath.Join(themeDir, "static", "fallbacks")
	require.NoError(t, os.MkdirAll(tmplDir, 0o755))
	require.NoError(t, os.MkdirAll(fbDir, 0o755))

	for _, name := range []string{"home.html", "campaign.html", "stage.html", "page.html", "404.html"} {
		require.NoError(t, os.WriteFile(filepath.Join(tmplDir, name), []byte(`{{define "title"}}x{{end}}{{define "content"}}x{{end}}`), 0o644))
	}

	th, err := LoadFromPath("emptyfb", themesDir)
	require.NoError(t, err)
	assert.Empty(t, th.FeaturedImageFallbacks)
}
