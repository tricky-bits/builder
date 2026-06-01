package builder

import "github.com/BurntSushi/toml"

// Config is the root configuration for a site build.
// It contains build/runtime settings (filesystem, theme) and template-visible
// site settings (base path, title, navigation, etc.).
type Config struct {
	Build BuildConfig `toml:"build"`
	Site  SiteConfig  `toml:"site"`
}

// BuildConfig contains build-time configuration used by the renderer pipeline.
// This data should not be passed to html/template directly.
type BuildConfig struct {
	// InputDir is the root directory containing source content (pages, posts, assets).
	InputDir string `toml:"input_dir"`

	// OutputDir is the destination directory where rendered files are written.
	OutputDir string `toml:"output_dir"`

	// ThemesDir is the directory where themes are located.
	ThemesDir string `toml:"themes_dir"`

	// Theme is the active theme name (resolved within ThemesDir).
	Theme string `toml:"theme"`

	// ObfuscationKey is the XOR key used to obfuscate future-stage metadata,
	// hint contents, and completion messages in rendered HTML. When empty,
	// obfuscation is disabled (payloads ship as plaintext) — dev mode.
	ObfuscationKey string `toml:"obfuscation_key,omitempty"`
}

// SiteConfig contains global site settings used during rendering and exposed to templates.
type SiteConfig struct {
	// Name is the project name.
	Name string `toml:"name"`

	// BasePath is the URL prefix under which the site is served (e.g. "/" or "/docs").
	BasePath string `toml:"base_path"`

	// DocumentTitle is the document title displayed in templates.
	DocumentTitle string `toml:"document_title"`

	// IndexTitle is the title on the index page.
	IndexTitle string `toml:"index_title"`

	// IndexSubtitle is the subtitle on the index page.
	IndexSubtitle string `toml:"index_subtitle"`

	// Copyright is a site-wide copyright notice.
	Copyright string `toml:"copyright,omitempty"`

	// FooterSlogan is the slogan displayed in the footer.
	FooterSlogan string `toml:"footer_slogan"`

	// NavItems are the navbar entries rendered in declared order.
	NavItems []NavItem `toml:"nav_items,omitempty"`

	// FooterItems are the footer link entries rendered in declared order.
	FooterItems []NavItem `toml:"footer_items,omitempty"`
}

// NavItem represents a single link entry used in the navbar or footer.
// This is template-visible by design.
type NavItem struct {
	// Label is the text shown to users (e.g., "Home", "About").
	Label string `toml:"label"`

	// Href is the target URL. Internal hrefs are resolved against Site.BasePath
	// by the template; External hrefs are emitted as-is.
	Href string `toml:"href"`

	// Icon is an optional icon identifier.
	Icon string `toml:"icon,omitempty"`

	// External marks the link as off-site. When true, the template emits
	// target="_blank" rel="noopener noreferrer" and skips BasePath resolution.
	External bool `toml:"external,omitempty"`
}

// ReadConfigFile decodes a TOML configuration file at path into a Config.
func ReadConfigFile(path string) (*Config, error) {
	var config Config

	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
