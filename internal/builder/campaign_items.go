package builder

import (
	"fmt"
	"html/template"
	"sort"
	"time"

	"github.com/tricky-bits/builder/internal/markdown"
)

// campaignItem is the per-campaign summary used by the index and campaigns
// listing pages.
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

// buildCampaignItems sorts all loaded campaigns (by Order then Slug) and
// builds their campaignItem summaries, split by featured flag. It also
// returns the full sorted list (both featured and non-featured combined)
// and the set of all tags used across every campaign.
func buildCampaignItems(b *Builder) (all, featured, nonFeatured []campaignItem, allTags []string, err error) {
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

	all = make([]campaignItem, 0, len(sorted))
	featured = make([]campaignItem, 0, len(sorted))
	nonFeatured = make([]campaignItem, 0, len(sorted))
	tagsSet := make(map[string]bool)

	for _, c := range sorted {
		var etaMinutes int
		itemTagsSet := make(map[string]bool)
		for _, s := range c.Stages {
			etaMinutes += s.Frontmatter.ETAMinutes
			for _, tag := range s.Frontmatter.Tags {
				itemTagsSet[tag] = true
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

		_, title, titleErr := markdown.HeroTitle(c.Frontmatter.Title)
		if titleErr != nil {
			return nil, nil, nil, nil, fmt.Errorf("[%s] render campaign hero title: %w", c.Frontmatter.Slug, titleErr)
		}

		item := campaignItem{
			Slug:              c.Frontmatter.Slug,
			Title:             title,
			Category:          c.Frontmatter.Category,
			Abstract:          c.Abstract,
			Order:             c.Frontmatter.Order,
			Difficulty:        c.Frontmatter.Difficulty,
			ETAMinutes:        etaMinutes,
			Tags:              sortedStringSet(itemTagsSet),
			PublishedAt:       c.Frontmatter.PublishedAt,
			LastUpdatedAt:     c.Frontmatter.LastUpdatedAt,
			StageCount:        len(c.Stages),
			StageStartSlug:    c.StartSlug,
			StageOrderedSlugs: stageOrderedSlugs,
			HasFeaturedImage:  c.HasFeaturedImage,
			FeaturedImageURL:  c.FeaturedImageURL,
		}

		all = append(all, item)
		if c.Frontmatter.Featured {
			featured = append(featured, item)
		} else {
			nonFeatured = append(nonFeatured, item)
		}
	}

	return all, featured, nonFeatured, sortedStringSet(tagsSet), nil
}
