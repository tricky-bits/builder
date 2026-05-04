package builder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadConfigFile_NavAndFooterItems(t *testing.T) {
	const tomlSrc = `
[build]
input_dir = "in"
output_dir = "out"
themes_dir = "themes"
theme = "base"

[site]
name = "Test Site"
base_path = "/"
document_title = "Test"
index_title = "Hi"
index_subtitle = "Sub"
footer_slogan = "Slogan"

[[site.nav_items]]
label = "Home"
href = "/index.html"

[[site.nav_items]]
label = "About"
href = "/about.html"
icon = "user"

[[site.footer_items]]
label = "About"
href = "/about.html"

[[site.footer_items]]
label = "GitHub"
href = "https://github.com/tricky-bits/builder"
external = true
`

	dir := t.TempDir()
	path := filepath.Join(dir, "tbb.toml")
	require.NoError(t, os.WriteFile(path, []byte(tomlSrc), 0o644))

	cfg, err := ReadConfigFile(path)
	require.NoError(t, err)

	assert.Equal(t, []NavItem{
		{Label: "Home", Href: "/index.html"},
		{Label: "About", Href: "/about.html", Icon: "user"},
	}, cfg.Site.NavItems)

	assert.Equal(t, []NavItem{
		{Label: "About", Href: "/about.html"},
		{Label: "GitHub", Href: "https://github.com/tricky-bits/builder", External: true},
	}, cfg.Site.FooterItems)
}

func TestReadConfigFile_EmptyNavAndFooter(t *testing.T) {
	const tomlSrc = `
[build]
input_dir = "in"
output_dir = "out"
themes_dir = "themes"
theme = "base"

[site]
name = "Test Site"
base_path = "/"
document_title = "Test"
index_title = "Hi"
index_subtitle = "Sub"
footer_slogan = "Slogan"
`

	dir := t.TempDir()
	path := filepath.Join(dir, "tbb.toml")
	require.NoError(t, os.WriteFile(path, []byte(tomlSrc), 0o644))

	cfg, err := ReadConfigFile(path)
	require.NoError(t, err)
	assert.Nil(t, cfg.Site.NavItems)
	assert.Nil(t, cfg.Site.FooterItems)
}
