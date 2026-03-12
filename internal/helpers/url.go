package helpers

import (
	"path/filepath"
	"regexp"
	"strings"
)

// DeriveSlug derives a URL-safe slug from a filename.
func DeriveSlug(filename string) string {
	// Remove extension
	name := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))

	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace spaces and underscores with hyphens
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")

	// Remove non-alphanumeric characters except hyphens
	reg := regexp.MustCompile(`[^a-z0-9-]+`)
	name = reg.ReplaceAllString(name, "")

	// Remove consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	name = reg.ReplaceAllString(name, "-")

	// Trim hyphens from edges
	name = strings.Trim(name, "-")

	return name
}
