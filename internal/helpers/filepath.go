package helpers

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ListFiles walks path recursively and returns file paths whose extension
// matches ext (e.g. ".md"). If path does not exist, it returns nil without
// an error. Returns an error if path exists but is not a directory.
func ListFiles(path, ext string) ([]string, error) {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("stat %q: %w", path, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", path)
	}

	var out []string
	if err := filepath.Walk(path, func(p string, fi os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if fi == nil || fi.IsDir() {
			return nil
		}
		if filepath.Ext(fi.Name()) != ext {
			return nil
		}
		out = append(out, p)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk %q: %w", path, err)
	}

	return out, nil
}
