package builder

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/tricky-bits/builder/internal/markdown"
)

// RenderIndexPage generates the site index page and writes it to the output directory.
func RenderIndexPage(b *Builder) error {
	t, err := b.themeMgr.Load(b.config.Build.Theme)
	if err != nil {
		return fmt.Errorf("[index.md] load theme: %w", err)
	}

	all, featured, _, _, err := buildCampaignItems(b)
	if err != nil {
		return fmt.Errorf("[index.md] build campaign items: %w", err)
	}

	heroTitleHTML, _, err := markdown.HeroTitle(b.config.Site.IndexTitle)
	if err != nil {
		return fmt.Errorf("[index.md] render hero title: %w", err)
	}

	data := struct {
		Site SiteConfig
		Hero struct {
			TitleHTML    template.HTML
			SubtitleText string
		}
		Featured   []campaignItem
		TotalCount int
	}{
		Site:       b.config.Site,
		Featured:   featured,
		TotalCount: len(all),
	}
	data.Hero.TitleHTML = heroTitleHTML
	data.Hero.SubtitleText = b.config.Site.IndexSubtitle

	var buffer bytes.Buffer
	if err := t.Render(&buffer, "home.html", data); err != nil {
		return fmt.Errorf("[index.md] render index page: %w", err)
	}

	outputPath := filepath.Join(b.config.Build.OutputDir, "index.html")
	if err := os.WriteFile(outputPath, buffer.Bytes(), 0o644); err != nil {
		return fmt.Errorf("[index.md] write rendered index page: %w", err)
	}

	b.logger.Info("built index page")
	return nil
}
