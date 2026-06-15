package theme

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverFallbacks_AcceptsAllFeaturedExtensions(t *testing.T) {
	staticDir := t.TempDir()
	fallbacksDir := filepath.Join(staticDir, "fallbacks")
	require.NoError(t, os.MkdirAll(fallbacksDir, 0o755))

	// One accepted file per supported extension, plus a non-image that must
	// be ignored and a subdirectory that must be skipped.
	for _, name := range []string{
		"a.avif", "b.webp", "c.jpg", "d.jpeg", "e.png", "f.gif",
		"g.PNG", // case-insensitive extension still accepted
		"notes.txt", "cover.svg",
	} {
		require.NoError(t, os.WriteFile(filepath.Join(fallbacksDir, name), []byte("x"), 0o644))
	}
	require.NoError(t, os.MkdirAll(filepath.Join(fallbacksDir, "nested"), 0o755))

	got, err := discoverFallbacks(staticDir)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"a.avif", "b.webp", "c.jpg", "d.jpeg", "e.png", "f.gif", "g.PNG",
	}, got)
}

func TestIsFeaturedImageExt(t *testing.T) {
	for _, ext := range []string{".avif", ".webp", ".jpg", ".jpeg", ".png", ".gif", ".PNG", ".JpG"} {
		assert.True(t, isFeaturedImageExt(ext), ext)
	}
	for _, ext := range []string{".txt", ".svg", ".bmp", "", ".pngx"} {
		assert.False(t, isFeaturedImageExt(ext), ext)
	}
}

func TestPickFallback(t *testing.T) {
	t.Run("empty candidates returns empty string", func(t *testing.T) {
		assert.Equal(t, "", PickFallback("any-slug", nil))
		assert.Equal(t, "", PickFallback("any-slug", []string{}))
	})

	t.Run("single candidate always picked", func(t *testing.T) {
		got := PickFallback("any-slug", []string{"only.png"})
		assert.Equal(t, "only.png", got)
	})

	t.Run("deterministic across calls", func(t *testing.T) {
		candidates := []string{"a.png", "b.png", "c.png", "d.png", "e.png"}
		first := PickFallback("my-campaign", candidates)
		for range 10 {
			assert.Equal(t, first, PickFallback("my-campaign", candidates))
		}
	})

	t.Run("different slugs against same list can yield different picks", func(t *testing.T) {
		candidates := []string{"a.png", "b.png", "c.png", "d.png", "e.png"}
		seen := make(map[string]bool)
		for _, slug := range []string{"alpha", "beta", "gamma", "delta", "epsilon", "zeta", "eta"} {
			seen[PickFallback(slug, candidates)] = true
		}
		assert.Greater(t, len(seen), 1, "expected at least two distinct picks across varied slugs")
	})

	t.Run("pick always in candidate list", func(t *testing.T) {
		candidates := []string{"a.png", "b.png", "c.png"}
		set := map[string]bool{"a.png": true, "b.png": true, "c.png": true}
		for _, slug := range []string{"x", "y", "z", "foo", "bar", "baz", "long-slug-name"} {
			assert.True(t, set[PickFallback(slug, candidates)])
		}
	})
}
