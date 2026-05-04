package builder

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/tricky-bits/builder/internal/markdown"
)

// RenderIndexPage generates the site index page and writes it to the output directory.
func RenderIndexPage(b *Builder) error {
	t, err := b.themeMgr.Load(b.config.Build.Theme)
	if err != nil {
		return fmt.Errorf("[index.md] load theme: %w", err)
	}

	// Sort campaigns by Order then Slug.
	sorted := make([]*Campaign, 0, len(b.campaigns))
	for _, c := range b.campaigns {
		sorted = append(sorted, c)
	}
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Frontmatter.Order != sorted[j].Frontmatter.Order {
			return sorted[i].Frontmatter.Order < sorted[j].Frontmatter.Order
		}
		return sorted[i].Frontmatter.Slug < sorted[j].Frontmatter.Slug
	})

	type campaignItem struct {
		Slug              string
		Title             string
		Category          string
		Abstract          template.HTML
		Order             int
		Difficulty        int
		ETAMinutes        int
		Tags              []string
		PublishedAt       *time.Time
		LastUpdatedAt     *time.Time
		StageCount        int
		StageStartSlug    string
		StageOrderedSlugs []string
		HasFeaturedImage  bool
		FeaturedImageURL  string
	}

	featured := make([]campaignItem, 0, len(sorted))
	nonFeatured := make([]campaignItem, 0, len(sorted))
	for _, c := range sorted {
		var etaMinutes int
		tagsSet := make(map[string]bool)
		for _, s := range c.Stages {
			etaMinutes += s.Frontmatter.ETAMinutes
			for _, tag := range s.Frontmatter.Tags {
				tagsSet[tag] = true
			}
		}

		stageOrderedSlugs := make([]string, 0, len(c.Stages))
		current := c.StartSlug
		for current != "" {
			stageOrderedSlugs = append(stageOrderedSlugs, current)
			stage, ok := c.Stages[current]
			if !ok {
				break
			}
			current = stage.Frontmatter.Next
		}

		item := campaignItem{
			Slug:              c.Frontmatter.Slug,
			Title:             c.Frontmatter.Title,
			Category:          c.Frontmatter.Category,
			Abstract:          c.Abstract,
			Order:             c.Frontmatter.Order,
			Difficulty:        c.Frontmatter.Difficulty,
			ETAMinutes:        etaMinutes,
			Tags:              sortedStringSet(tagsSet),
			PublishedAt:       c.Frontmatter.PublishedAt,
			LastUpdatedAt:     c.Frontmatter.LastUpdatedAt,
			StageCount:        len(c.Stages),
			StageStartSlug:    c.StartSlug,
			StageOrderedSlugs: stageOrderedSlugs,
			HasFeaturedImage:  c.HasFeaturedImage,
			FeaturedImageURL:  c.FeaturedImageURL,
		}
		if c.Frontmatter.Featured {
			featured = append(featured, item)
		} else {
			nonFeatured = append(nonFeatured, item)
		}
	}

	heroTitleHTML, err := markdown.RenderHeroTitle(b.config.Site.IndexTitle)
	if err != nil {
		return fmt.Errorf("[index.md] render hero title: %w", err)
	}

	data := struct {
		Site SiteConfig
		Hero struct {
			TitleHTML    template.HTML
			SubtitleText string
		}
		Featured    []campaignItem
		NonFeatured []campaignItem
	}{
		Site:        b.config.Site,
		Featured:    featured,
		NonFeatured: nonFeatured,
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
