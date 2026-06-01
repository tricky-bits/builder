// Package builder provides the core builder engine for loading and rendering
// campaigns, stages, and pages. It manages content discovery, theme resolution,
// and coordinates rendering through the theme system.
package builder

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/lmittmann/tint"
	"github.com/tricky-bits/builder/internal/obfuscate"
	"github.com/tricky-bits/builder/internal/theme"
)

// Builder represents a configured builder instance with loaded campaigns,
// pages, and theme management.
type Builder struct {
	config    *Config
	themeMgr  *theme.Manager
	campaigns map[string]*Campaign
	pages     map[string]*Page
	logger    *slog.Logger
	encoder   *obfuscate.Encoder
}

// New creates a new Builder instance with the given configuration.
// It sets up a tint-coloured slog logger on stderr, validates the global theme,
// and returns a ready-to-use Builder.
func New(cfg *Config) (*Builder, error) {
	// Remove trailing slash from base path
	cfg.Site.BasePath = strings.TrimSuffix(cfg.Site.BasePath, "/")

	logger := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelInfo}))

	// Pass OutputDir so the manager auto-copies assets on first theme load.
	themeMgr := theme.NewManager(cfg.Build.ThemesDir, cfg.Build.OutputDir)

	// Verify global theme exists.
	if _, err := themeMgr.Load(cfg.Build.Theme); err != nil {
		return nil, fmt.Errorf("load global theme %s: %w", cfg.Build.Theme, err)
	}

	return &Builder{
		config:    cfg,
		themeMgr:  themeMgr,
		campaigns: make(map[string]*Campaign),
		logger:    logger,
		encoder:   obfuscate.New(cfg.Build.ObfuscationKey),
	}, nil
}

// Load discovers and loads all campaigns and pages from the configured input directory.
func (b *Builder) Load() error {
	var err error

	b.campaigns, err = LoadAllCampaigns(filepath.Join(b.config.Build.InputDir, "campaigns"))
	if err != nil {
		return err
	}
	b.logger.Info("loaded campaigns", "count", len(b.campaigns))

	b.pages, err = ReadAllPages(filepath.Join(b.config.Build.InputDir, "pages"))
	if err != nil {
		return err
	}
	b.logger.Info("loaded pages", "count", len(b.pages))

	return nil
}

// Build orchestrates the full build: output directory, campaigns, pages,
// and the index page. Theme assets are copied automatically by the theme manager
// the first time each theme is loaded.
func (b *Builder) Build() error {
	if err := os.MkdirAll(b.config.Build.OutputDir, 0o755); err != nil {
		return err
	}

	if err := b.BuildAllCampaigns(); err != nil {
		return err
	}

	if err := b.BuildAllPages(); err != nil {
		return err
	}

	if err := b.BuildIndexPage(); err != nil {
		return err
	}

	return nil
}

// BuildAllCampaigns renders every loaded campaign and its stages.
func (b *Builder) BuildAllCampaigns() error {
	for _, campaign := range b.campaigns {
		if err := campaign.Build(b); err != nil {
			return err
		}
	}
	return nil
}

// BuildAllPages renders every loaded page.
func (b *Builder) BuildAllPages() error {
	b.logger.Info("rendering pages", "count", len(b.pages))
	for _, page := range b.pages {
		if err := page.Build(b); err != nil {
			return err
		}
	}
	return nil
}

// BuildIndexPage renders the site index page.
func (b *Builder) BuildIndexPage() error {
	return RenderIndexPage(b)
}
