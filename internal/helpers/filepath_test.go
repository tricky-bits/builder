package helpers

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListFiles(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(dir string) // creates files under dir
		ext       string
		wantFiles []string // relative to dir; nil means expect nil result
		wantErr   bool
	}{
		{
			name:      "non-existent directory returns nil",
			setup:     func(dir string) {},
			ext:       ".md",
			wantFiles: nil,
		},
		{
			name: "empty directory returns nil",
			setup: func(dir string) {
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "target"), 0o755))
			},
			ext:       ".md",
			wantFiles: nil,
		},
		{
			name: "returns matching files",
			setup: func(dir string) {
				d := filepath.Join(dir, "target")
				require.NoError(t, os.MkdirAll(d, 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(d, "a.md"), []byte("a"), 0o644))
				require.NoError(t, os.WriteFile(filepath.Join(d, "b.md"), []byte("b"), 0o644))
			},
			ext:       ".md",
			wantFiles: []string{"a.md", "b.md"},
		},
		{
			name: "ignores non-matching extensions",
			setup: func(dir string) {
				d := filepath.Join(dir, "target")
				require.NoError(t, os.MkdirAll(d, 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(d, "a.md"), []byte("a"), 0o644))
				require.NoError(t, os.WriteFile(filepath.Join(d, "b.txt"), []byte("b"), 0o644))
				require.NoError(t, os.WriteFile(filepath.Join(d, "c.go"), []byte("c"), 0o644))
			},
			ext:       ".md",
			wantFiles: []string{"a.md"},
		},
		{
			name: "recurses into subdirectories",
			setup: func(dir string) {
				d := filepath.Join(dir, "target")
				sub := filepath.Join(d, "sub")
				require.NoError(t, os.MkdirAll(sub, 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(d, "a.md"), []byte("a"), 0o644))
				require.NoError(t, os.WriteFile(filepath.Join(sub, "b.md"), []byte("b"), 0o644))
			},
			ext:       ".md",
			wantFiles: []string{"a.md", filepath.Join("sub", "b.md")},
		},
		{
			name: "path is not a directory returns error",
			setup: func(dir string) {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "target"), []byte("file"), 0o644))
			},
			ext:     ".md",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			tt.setup(tmp)

			var path string
			if tt.name == "non-existent directory returns nil" {
				path = filepath.Join(tmp, "does-not-exist")
			} else {
				path = filepath.Join(tmp, "target")
			}

			got, err := ListFiles(path, tt.ext)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			if tt.wantFiles == nil {
				assert.Nil(t, got)
				return
			}

			// Build expected absolute paths
			want := make([]string, len(tt.wantFiles))
			for i, f := range tt.wantFiles {
				want[i] = filepath.Join(path, f)
			}

			sort.Strings(got)
			sort.Strings(want)
			assert.Equal(t, want, got)
		})
	}
}
