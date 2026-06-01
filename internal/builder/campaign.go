package builder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"time"

	"github.com/tricky-bits/builder/internal/helpers"
	"github.com/tricky-bits/builder/internal/markdown"
	"github.com/tricky-bits/builder/internal/theme"
)

// stagePayload is the per-stage record encoded into the page's `payloads`
// blob. The client-side decoder in progress.js unmarshals an ordered array
// of these and reveals each entry when its predecessor is solved.
type stagePayload struct {
	Slug       string `json:"slug"`
	Title      string `json:"title"`
	Difficulty int    `json:"difficulty"`
	ETA        int    `json:"eta"`
	Href       string `json:"href"`
}

// encodeStagePayloads builds the ordered per-stage payload array for c,
// JSON-marshals it, and returns the result run through the builder's
// encoder (passthrough when the obfuscation key is empty).
func encodeStagePayloads(b *Builder, c *Campaign) (string, error) {
	payloads := make([]stagePayload, 0, len(c.Stages))
	current := c.StartSlug
	for current != "" {
		s, ok := c.Stages[current]
		if !ok {
			break
		}
		payloads = append(payloads, stagePayload{
			Slug:       current,
			Title:      s.Frontmatter.Title,
			Difficulty: s.Frontmatter.Difficulty,
			ETA:        s.Frontmatter.ETAMinutes,
			Href:       theme.JoinURL(b.config.Site.BasePath, "campaigns", c.Frontmatter.Slug, current),
		})
		current = s.Frontmatter.Next
	}
	raw, err := json.Marshal(payloads)
	if err != nil {
		return "", err
	}
	return b.encoder.Encode(string(raw)), nil
}

// CampaignFrontmatter represents YAML metadata for campaign.md.
//
// It contains campaign-level metadata used for discovery (category), user-facing
// characteristics (difficulty), publication tracking, theming, and display ordering.
// Campaign-level settings may be inherited by contained stages according to the
// project's theme resolution rules.
type CampaignFrontmatter struct {
	// Title is the human-readable campaign title (required).
	Title string `yaml:"title"`

	// Category of the campaign (required).
	Category string `yaml:"category"`

	// Slug is the logical route identifier for the campaign (optional). If omitted,
	// it is typically derived from the directory name.
	Slug string `yaml:"slug,omitempty"`

	// Difficulty is an optional difficulty indicator for the campaign.
	Difficulty int `yaml:"difficulty,omitempty"`

	// PublishedAt is the campaign publication timestamp (optional).
	PublishedAt *time.Time `yaml:"published_at,omitempty"`

	// LastUpdatedAt is the campaign last update timestamp (optional).
	LastUpdatedAt *time.Time `yaml:"last_update_at,omitempty"`

	// Theme is an optional theme override at campaign level.
	// Effective theme resolution is stage.theme > campaign.theme > global.theme.
	Theme string `yaml:"theme,omitempty"`

	// Order determines the display order of campaigns in listings.
	// Lower values appear first. If not specified, campaigns are ordered by slug.
	Order int `yaml:"order,omitempty"`

	// Featured marks the campaign for the home-page Featured block. Featured
	// campaigns appear in the scroll-snap strip regardless of completion state.
	Featured bool `yaml:"featured,omitempty"`

	// Abstract is a short summary displayed on the site index instead of the full
	// body (optional).
	Abstract string `yaml:"abstract,omitempty"`

	// CompletionMessage is the message displayed when a campaign is completed (optional).
	CompletionMessage string `yaml:"completion_message"`

	// Assets is a list of file names (e.g., images) that should be copied alongside the
	// campaign. These files should be present in the campaign directory.
	Assets []string `yaml:"assets,omitempty"`
}

// Campaign is the in-memory representation of a fully loaded campaign.
//
// A campaign corresponds to a directory containing a campaign.md file and one
// or more stage files under a stages/ subdirectory. The Campaign aggregates
// parsed and rendered campaign content, associated stages, and derived data
// computed during loading/validation (e.g., the start stage slug).
type Campaign struct {
	// SourceDir is the full filesystem path to the campaign directory.
	SourceDir string

	// Frontmatter contains the parsed YAML metadata from campaign.md.
	Frontmatter CampaignFrontmatter

	// Content is the rendered HTML body of campaign.md.
	Content template.HTML

	// Abstract is the rendered HTML representation of Frontmatter.Abstract.
	Abstract template.HTML

	// CompletionMessage is the rendered HTML message displayed upon completion.
	CompletionMessage template.HTML

	// Stages contains the campaign stages keyed by stage slug (unique within the
	// campaign). This enables direct lookup when resolving navigation (e.g., Next).
	Stages map[string]*Stage

	// StartSlug is the slug of the stage marked as the entry point
	// (i.e., stage frontmatter.start == true). Derived during loading/validation.
	StartSlug string

	// HasFeaturedImage is true whenever the campaign has any featured image
	// to render — either a real featured.png next to the campaign source or
	// a theme-provided fallback assigned during the build.
	HasFeaturedImage bool

	// FeaturedImageURL is the absolute URL of the image to render (real or
	// fallback). Empty when neither a real featured image nor a fallback is
	// available.
	FeaturedImageURL string
}

// LoadAllCampaigns iterates over subdirectories of basepath, loads each as a
// campaign, and returns them keyed by campaign slug.
func LoadAllCampaigns(basepath string) (map[string]*Campaign, error) {
	entries, err := os.ReadDir(basepath)
	if err != nil {
		return nil, err
	}

	campaigns := make(map[string]*Campaign)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		campaign, err := LoadCampaign(filepath.Join(basepath, e.Name()))
		if err != nil {
			return nil, err
		}

		campaigns[campaign.Frontmatter.Slug] = campaign
	}

	return campaigns, nil
}

// LoadCampaign loads and validates a campaign from basepath. basepath must be a
// directory containing campaign.md and a stages/ subdirectory. Returns a
// descriptive error when campaign.md is missing.
func LoadCampaign(basepath string) (*Campaign, error) {
	campaign := &Campaign{
		SourceDir: basepath,
	}

	campaignFile := filepath.Join(basepath, "campaign.md")
	if _, err := os.Stat(campaignFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("[%s] campaign.md not found in %s", filepath.Base(basepath), basepath)
	}

	// Parse campaign.md: render body to buf, decode frontmatter
	var contentBuf bytes.Buffer
	if err := markdown.Parse(campaignFile, &contentBuf, &campaign.Frontmatter); err != nil {
		return nil, fmt.Errorf("[%s] parse campaign.md: %w", filepath.Base(basepath), err)
	}
	campaign.Content = template.HTML(contentBuf.String())

	// Derive slug from directory name if not provided
	if campaign.Frontmatter.Slug == "" {
		campaign.Frontmatter.Slug = helpers.DeriveSlug(filepath.Base(basepath))
	}

	// Render abstract markdown to HTML if provided
	if campaign.Frontmatter.Abstract != "" {
		rendered, err := markdown.Render(campaign.Frontmatter.Abstract)
		if err != nil {
			return nil, fmt.Errorf("[%s] render campaign abstract: %w", filepath.Base(basepath), err)
		}
		campaign.Abstract = rendered
	}

	// Render completion message if provided
	if campaign.Frontmatter.CompletionMessage != "" {
		rendered, err := markdown.Render(campaign.Frontmatter.CompletionMessage)
		if err != nil {
			return nil, fmt.Errorf("[%s] render campaign completion message: %w", filepath.Base(basepath), err)
		}
		campaign.CompletionMessage = rendered
	}

	// Discover and load stages
	var err error
	campaign.Stages, err = ReadAllStages(filepath.Join(basepath, "stages"))
	if err != nil {
		return nil, fmt.Errorf("[%s] discover stages: %w", filepath.Base(basepath), err)
	}

	if err := campaign.Validate(); err != nil {
		return nil, err
	}

	return campaign, nil
}

// Validate checks campaign-level invariants: required fields, exactly one start
// stage, valid asset files, and a valid linear stage chain. Sets c.StartSlug.
func (c *Campaign) Validate() error {
	if c.Frontmatter.Title == "" {
		return fmt.Errorf("[%s] title is required", c.Frontmatter.Slug)
	}
	if c.Frontmatter.Category == "" {
		return fmt.Errorf("[%s] category is required", c.Frontmatter.Slug)
	}

	// Find exactly one start stage
	var startSlug string
	for slug, stage := range c.Stages {
		if stage.Frontmatter.Start {
			if startSlug != "" {
				return fmt.Errorf("[%s] multiple start stages: %s and %s", c.Frontmatter.Slug, startSlug, slug)
			}
			startSlug = slug
		}
	}
	if startSlug == "" {
		return fmt.Errorf("[%s] no start stage found", c.Frontmatter.Slug)
	}
	c.StartSlug = startSlug

	if err := c.ValidateLinearChain(); err != nil {
		return err
	}

	// Validate asset files exist
	for _, assetFile := range c.Frontmatter.Assets {
		assetPath := filepath.Join(c.SourceDir, assetFile)
		if _, err := os.Stat(assetPath); os.IsNotExist(err) {
			return fmt.Errorf("[%s] asset file not found: %s", c.Frontmatter.Slug, assetPath)
		}
	}

	return nil
}

// ValidateLinearChain ensures that all stages form a single linear chain
// reachable from StartSlug. Returns an error if a cycle is detected or if
// any stage is unreachable (orphaned).
func (c *Campaign) ValidateLinearChain() error {
	if c.StartSlug == "" {
		return fmt.Errorf("[%s] cannot validate linear chain: no start stage", c.Frontmatter.Slug)
	}

	visited := make(map[string]bool)
	current := c.StartSlug

	for current != "" {
		if visited[current] {
			return fmt.Errorf("[%s] cycle detected in stage chain at: %s", c.Frontmatter.Slug, current)
		}
		visited[current] = true

		stage, exists := c.Stages[current]
		if !exists {
			return fmt.Errorf("[%s] stage not found: %s", c.Frontmatter.Slug, current)
		}

		current = stage.Frontmatter.Next
	}

	// Check for orphaned stages not reachable from start
	if len(visited) != len(c.Stages) {
		var orphans []string
		for slug := range c.Stages {
			if !visited[slug] {
				orphans = append(orphans, slug)
			}
		}
		return fmt.Errorf("[%s] orphaned stages not reachable from start: %v", c.Frontmatter.Slug, orphans)
	}

	return nil
}

// Build renders the campaign index page and all its stages to the output directory.
func (c *Campaign) Build(b *Builder) error {
	outputDir := filepath.Join(b.config.Build.OutputDir, "campaigns", c.Frontmatter.Slug)

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("[%s] create campaign output directory: %w", c.Frontmatter.Slug, err)
	}

	if err := c.buildFeaturedImage(b); err != nil {
		return err
	}

	if err := c.buildCampaignPage(b, outputDir); err != nil {
		return err
	}

	if err := c.buildCampaignAssets(b, outputDir); err != nil {
		return err
	}

	// Build all stages, propagating errors.
	for _, s := range c.Stages {
		if err := s.Build(b, c); err != nil {
			return err
		}
	}

	b.logger.Info("built campaign", "slug", c.Frontmatter.Slug, "stages", len(c.Stages))
	return nil
}

// buildFeaturedImage resolves the campaign's featured image, setting
// HasFeaturedImage and FeaturedImageURL. A real featured.png next to the
// campaign source wins: it is added to the asset list (so it gets copied) and
// its absolute URL is resolved. Otherwise a deterministic theme fallback is
// picked using the effective theme (campaign.theme > global.theme). When the
// theme exposes no fallbacks, the flag stays false and templates fall back to
// the gradient placeholder. Must run before buildCampaignPage (reads the
// flag/URL) and buildCampaignAssets (copies featured.png).
func (c *Campaign) buildFeaturedImage(b *Builder) error {
	featuredPath := filepath.Join(c.SourceDir, "featured.png")
	if _, err := os.Stat(featuredPath); err == nil {
		if !slices.Contains(c.Frontmatter.Assets, "featured.png") {
			c.Frontmatter.Assets = append(c.Frontmatter.Assets, "featured.png")
		}
		c.HasFeaturedImage = true
		c.FeaturedImageURL = theme.JoinURL(b.config.Site.BasePath, "campaigns", c.Frontmatter.Slug, "featured.png")
		return nil
	}

	t, err := b.themeMgr.Load(c.Frontmatter.Theme, b.config.Build.Theme)
	if err != nil {
		return fmt.Errorf("[%s] resolve theme for fallback: %w", c.Frontmatter.Slug, err)
	}
	pick := theme.PickFallback(c.Frontmatter.Slug, t.FeaturedImageFallbacks)
	if pick == "" {
		return nil
	}
	c.FeaturedImageURL = theme.JoinURL(b.config.Site.BasePath, "assets", t.Name, "fallbacks", pick)
	c.HasFeaturedImage = true
	return nil
}

// buildCampaignPage renders the campaign index HTML template and writes index.html to outputDir.
func (c *Campaign) buildCampaignPage(b *Builder, outputDir string) error {
	t, err := b.themeMgr.Load(c.Frontmatter.Theme, b.config.Build.Theme)
	if err != nil {
		return fmt.Errorf("[%s] load theme: %w", c.Frontmatter.Slug, err)
	}

	payloads, err := encodeStagePayloads(b, c)
	if err != nil {
		return fmt.Errorf("[%s] encode stage payloads: %w", c.Frontmatter.Slug, err)
	}

	// Aggregate tags, authors, and ETA across stages.
	authorsSet := make(map[string]bool)
	tagsSet := make(map[string]bool)
	var etaMinutes int
	for _, s := range c.Stages {
		authorsSet[s.Frontmatter.Author] = true
		etaMinutes += s.Frontmatter.ETAMinutes
		for _, tag := range s.Frontmatter.Tags {
			tagsSet[tag] = true
		}
	}

	type campaignData struct {
		Slug                 string
		Title                string
		Category             string
		Theme                string
		Content              template.HTML
		Abstract             template.HTML
		EncCompletionMessage string
		Order                int
		Difficulty           int
		ETAMinutes           int
		Tags                 []string
		Authors              []string
		PublishedAt          *time.Time
		LastUpdatedAt        *time.Time
		StageStartSlug       string
		StageCount           int
		Payloads             string
		HasFeaturedImage     bool
		FeaturedImageURL     string
	}

	encCompletion := ""
	if c.CompletionMessage != "" {
		encCompletion = b.encoder.Encode(string(c.CompletionMessage))
	}

	data := struct {
		Site     SiteConfig
		Campaign campaignData
		K        string
	}{
		Site: b.config.Site,
		Campaign: campaignData{
			Slug:                 c.Frontmatter.Slug,
			Title:                c.Frontmatter.Title,
			Category:             c.Frontmatter.Category,
			Theme:                c.Frontmatter.Theme,
			Content:              c.Content,
			Abstract:             c.Abstract,
			EncCompletionMessage: encCompletion,
			Order:                c.Frontmatter.Order,
			Difficulty:           c.Frontmatter.Difficulty,
			ETAMinutes:           etaMinutes,
			Tags:                 sortedStringSet(tagsSet),
			Authors:              sortedStringSet(authorsSet),
			PublishedAt:          c.Frontmatter.PublishedAt,
			LastUpdatedAt:        c.Frontmatter.LastUpdatedAt,
			StageStartSlug:       c.StartSlug,
			StageCount:           len(c.Stages),
			Payloads:             payloads,
			HasFeaturedImage:     c.HasFeaturedImage,
			FeaturedImageURL:     c.FeaturedImageURL,
		},
		K: b.config.Build.ObfuscationKey,
	}

	var buffer bytes.Buffer
	if err := t.Render(&buffer, "campaign.html", data); err != nil {
		return fmt.Errorf("[%s] render campaign: %w", c.Frontmatter.Slug, err)
	}

	outputPath := filepath.Join(outputDir, "index.html")
	if err := os.WriteFile(outputPath, buffer.Bytes(), 0o644); err != nil {
		return fmt.Errorf("[%s] write rendered campaign: %w", c.Frontmatter.Slug, err)
	}

	return nil
}

// buildCampaignAssets copies the campaign's asset files to outputDir.
func (c *Campaign) buildCampaignAssets(b *Builder, outputDir string) error {
	for _, assetFile := range c.Frontmatter.Assets {
		srcPath := filepath.Join(c.SourceDir, assetFile)
		dstPath := filepath.Join(outputDir, assetFile)
		if err := helpers.CopyFile(srcPath, dstPath); err != nil {
			return fmt.Errorf("[%s] copy asset file %q: %w", c.Frontmatter.Slug, assetFile, err)
		}
		b.logger.Info("copied asset file", "campaign", c.Frontmatter.Slug, "file", assetFile)
	}

	return nil
}

// sortedStringSet returns the keys of m sorted alphabetically.
func sortedStringSet(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
