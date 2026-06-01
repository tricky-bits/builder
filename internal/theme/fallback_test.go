package theme

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
