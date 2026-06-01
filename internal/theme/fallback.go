package theme

import (
	"hash/fnv"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// PickFallback selects a deterministic candidate filename for slug.
// Returns "" when candidates is empty. The hash input is just the slug,
// so the same slug against the same candidate list always returns the
// same entry. Candidates must already be in a stable order.
func PickFallback(slug string, candidates []string) string {
	if len(candidates) == 0 {
		return ""
	}
	h := fnv.New32a()
	h.Write([]byte(slug))
	return candidates[int(h.Sum32())%len(candidates)]
}

// discoverFallbacks scans <staticDir>/fallbacks/ for *.png files and
// returns their filenames sorted alphabetically. A missing or empty
// static dir, missing fallbacks dir, or empty fallbacks dir yields nil.
func discoverFallbacks(staticDir string) ([]string, error) {
	if staticDir == "" {
		return nil, nil
	}
	dir := filepath.Join(staticDir, "fallbacks")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(e.Name()), ".png") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}
