package builder

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

// RenderCampaignsPage generates the "/campaigns" listing page (all campaigns,
// split into not-yet-finished and finished client-side, with a tag filter)
// and writes it to <outputDir>/campaigns/index.html.
func RenderCampaignsPage(b *Builder) error {
	t, err := b.themeMgr.Load(b.config.Build.Theme)
	if err != nil {
		return fmt.Errorf("[campaigns.md] load theme: %w", err)
	}

	all, _, _, allTags, err := buildCampaignItems(b)
	if err != nil {
		return fmt.Errorf("[campaigns.md] build campaign items: %w", err)
	}

	data := struct {
		Site      SiteConfig
		Campaigns []campaignItem
		Tags      []string
	}{
		Site:      b.config.Site,
		Campaigns: all,
		Tags:      allTags,
	}

	var buffer bytes.Buffer
	if err := t.Render(&buffer, "campaigns.html", data); err != nil {
		return fmt.Errorf("[campaigns.md] render campaigns page: %w", err)
	}

	outputDir := filepath.Join(b.config.Build.OutputDir, "campaigns")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("[campaigns.md] create campaigns output dir: %w", err)
	}

	outputPath := filepath.Join(outputDir, "index.html")
	if err := os.WriteFile(outputPath, buffer.Bytes(), 0o644); err != nil {
		return fmt.Errorf("[campaigns.md] write rendered campaigns page: %w", err)
	}

	b.logger.Info("built campaigns page")
	return nil
}
