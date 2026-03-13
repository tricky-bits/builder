package theme

import (
	"path"
	"strings"
)

// JoinURL constructs a root-relative URL from a base path and additional elements.
// Absolute URLs (starting with http:// or https://) are returned unchanged.
// The result always starts with a leading slash.
func JoinURL(base string, elems ...string) string {
	// If base is an absolute URL, return it unchanged
	if strings.HasPrefix(base, "http://") || strings.HasPrefix(base, "https://") {
		return base
	}

	joined := path.Join(append([]string{base}, elems...)...)

	// Ensure leading slash
	if !strings.HasPrefix(joined, "/") {
		joined = "/" + joined
	}

	return joined
}
