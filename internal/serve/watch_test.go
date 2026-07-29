package serve

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithin(t *testing.T) {
	tests := []struct {
		name       string
		root, path string
		want       bool
	}{
		{"direct child", "/a/b", "/a/b/c.txt", true},
		{"nested child", "/a/b", "/a/b/c/d.txt", true},
		{"same as root", "/a/b", "/a/b", true},
		{"sibling", "/a/b", "/a/c/d.txt", false},
		{"parent", "/a/b", "/a", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, within(tt.root, tt.path))
		})
	}
}

func TestAddRecursive_WatchesEveryDir(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "sub", "nested"), 0o755))

	fsw, err := fsnotify.NewWatcher()
	require.NoError(t, err)
	defer fsw.Close()

	dirs, err := addRecursive(fsw, root)
	require.NoError(t, err)
	assert.Len(t, dirs, 3) // root, sub, sub/nested
}

func TestAddRecursive_NonexistentRoot(t *testing.T) {
	fsw, err := fsnotify.NewWatcher()
	require.NoError(t, err)
	defer fsw.Close()

	_, err = addRecursive(fsw, filepath.Join(t.TempDir(), "does-not-exist"))
	assert.Error(t, err)
}

func newTestWatcher(t *testing.T, inputDir, themeDir string) *Watcher {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "config.toml")
	require.NoError(t, os.WriteFile(configPath, []byte("x"), 0o644))

	w, err := newWatcher(inputDir, configPath, themeDir)
	require.NoError(t, err)
	t.Cleanup(func() { w.Close() })
	return w
}

func TestNewWatcher_Success(t *testing.T) {
	w := newTestWatcher(t, t.TempDir(), t.TempDir())
	assert.NotNil(t, w)
}

func TestNewWatcher_BadInputDir(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	require.NoError(t, os.WriteFile(configPath, []byte("x"), 0o644))

	_, err := newWatcher(filepath.Join(t.TempDir(), "missing"), configPath, t.TempDir())
	assert.Error(t, err)
}

func TestWatcher_Relevant(t *testing.T) {
	inputDir := t.TempDir()
	themeDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "config.toml")
	require.NoError(t, os.WriteFile(configPath, []byte("x"), 0o644))

	w, err := newWatcher(inputDir, configPath, themeDir)
	require.NoError(t, err)
	defer w.Close()

	assert.True(t, w.relevant(filepath.Join(inputDir, "campaigns", "foo.md")))
	assert.True(t, w.relevant(filepath.Join(themeDir, "templates", "stage.html")))
	assert.True(t, w.relevant(configPath))
	assert.False(t, w.relevant(filepath.Join(filepath.Dir(configPath), "unrelated.txt")))
	assert.False(t, w.relevant("/some/unrelated/path.md"))
}

func TestUpdateThemeDir_Success(t *testing.T) {
	oldTheme := t.TempDir()
	newTheme := t.TempDir()
	w := newTestWatcher(t, t.TempDir(), oldTheme)

	require.NoError(t, w.UpdateThemeDir(newTheme))

	assert.True(t, w.relevant(filepath.Join(newTheme, "foo")))
	assert.False(t, w.relevant(filepath.Join(oldTheme, "foo")))
}

func TestUpdateThemeDir_BadDir(t *testing.T) {
	w := newTestWatcher(t, t.TempDir(), t.TempDir())

	err := w.UpdateThemeDir(filepath.Join(t.TempDir(), "missing"))
	assert.Error(t, err)
}

func TestWatcher_DebounceTriggersOnChangeAfterFileWrite(t *testing.T) {
	inputDir := t.TempDir()
	w := newTestWatcher(t, inputDir, t.TempDir())

	changed := make(chan struct{}, 1)
	w.Start(func() {
		select {
		case changed <- struct{}{}:
		default:
		}
	})

	require.NoError(t, os.WriteFile(filepath.Join(inputDir, "new.md"), []byte("hello"), 0o644))

	select {
	case <-changed:
	case <-time.After(2 * time.Second):
		t.Fatal("onChange not called within timeout after file write")
	}
}
