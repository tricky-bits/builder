package builder

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tricky-bits/builder/internal/theme"
)

func newTestBuilderForCampaignsPage(t *testing.T, campaigns map[string]*Campaign) *Builder {
	t.Helper()

	themesDir, err := filepath.Abs("../../themes")
	require.NoError(t, err)
	outputDir := t.TempDir()

	return &Builder{
		config: &Config{
			Build: BuildConfig{Theme: "base", OutputDir: outputDir},
			Site:  SiteConfig{BasePath: "/", DocumentTitle: "Test Site"},
		},
		themeMgr:  theme.NewManager(themesDir, outputDir),
		campaigns: campaigns,
		logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

func TestRenderCampaignsPage(t *testing.T) {
	t.Run("writes campaigns/index.html with all campaigns and aggregated tags", func(t *testing.T) {
		campaigns := map[string]*Campaign{
			"featured-one": {
				Frontmatter: CampaignFrontmatter{Title: "Featured One", Slug: "featured-one", Category: "crypto", Featured: true},
				StartSlug:   "stage-01",
				Stages: map[string]*Stage{
					"stage-01": {Frontmatter: StageFrontmatter{Tags: []string{"go", "crypto"}}},
				},
			},
			"plain-two": {
				Frontmatter: CampaignFrontmatter{Title: "Plain Two", Slug: "plain-two", Category: "web"},
				StartSlug:   "stage-01",
				Stages: map[string]*Stage{
					"stage-01": {Frontmatter: StageFrontmatter{Tags: []string{"web", "crypto"}}},
				},
			},
		}
		b := newTestBuilderForCampaignsPage(t, campaigns)

		require.NoError(t, RenderCampaignsPage(b))

		outputPath := filepath.Join(b.config.Build.OutputDir, "campaigns", "index.html")
		content, err := os.ReadFile(outputPath)
		require.NoError(t, err)

		assert.Contains(t, string(content), "Featured One")
		assert.Contains(t, string(content), "Plain Two")
		// Tags aggregated across all campaigns, deduplicated and sorted.
		assert.Contains(t, string(content), "crypto")
		assert.Contains(t, string(content), "go")
		assert.Contains(t, string(content), "web")
	})

	t.Run("no campaigns still renders a valid page", func(t *testing.T) {
		b := newTestBuilderForCampaignsPage(t, map[string]*Campaign{})

		require.NoError(t, RenderCampaignsPage(b))

		outputPath := filepath.Join(b.config.Build.OutputDir, "campaigns", "index.html")
		_, err := os.ReadFile(outputPath)
		require.NoError(t, err)
	})

	t.Run("theme load failure is propagated", func(t *testing.T) {
		b := newTestBuilderForCampaignsPage(t, nil)
		b.config.Build.Theme = "does-not-exist"

		err := RenderCampaignsPage(b)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "load theme")
	})
}
