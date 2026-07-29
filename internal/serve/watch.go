package serve

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const debounceInterval = 300 * time.Millisecond

// Watcher watches input_dir (permanently, for the process lifetime), the
// active theme's dir (swappable via UpdateThemeDir when the config's theme
// name changes), and the config file's parent dir.
//
// The config file itself is watched via its parent directory rather than its
// literal path: editors that save-via-rename replace the inode on every
// save, which silently kills a watch on the direct file path but survives a
// directory watch filtered by basename.
type Watcher struct {
	fsw *fsnotify.Watcher

	// Set once in newWatcher, never mutated after: safe to read without a lock.
	inputDir   string
	configDir  string
	configName string

	// Mutated by UpdateThemeDir/debounce after construction: guarded by mu.
	mu        sync.Mutex
	themeDirs []string
	themeDir  string
	timer     *time.Timer

	onChange func()
}

// newWatcher sets up all fsnotify watches but does not start processing
// events yet — call Start once the caller has whatever it needs (e.g. the
// Watcher itself, to reference from the onChange callback) fully wired up.
func newWatcher(inputDir, configPath, themeDir string) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if _, err := addRecursive(fsw, inputDir); err != nil {
		fsw.Close()
		return nil, err
	}

	themeDirs, err := addRecursive(fsw, themeDir)
	if err != nil {
		fsw.Close()
		return nil, err
	}

	configDir := filepath.Dir(configPath)
	if err := fsw.Add(configDir); err != nil {
		fsw.Close()
		return nil, err
	}

	w := &Watcher{
		fsw:        fsw,
		inputDir:   inputDir,
		themeDirs:  themeDirs,
		themeDir:   themeDir,
		configDir:  configDir,
		configName: filepath.Base(configPath),
	}

	return w, nil
}

// Start begins processing fs events, invoking onChange (debounced) for every
// relevant change.
func (w *Watcher) Start(onChange func()) {
	w.onChange = onChange
	go w.loop()
}

func (w *Watcher) loop() {
	for {
		select {
		case ev, ok := <-w.fsw.Events:
			if !ok {
				return
			}
			w.handleEvent(ev)
		case err, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
			slog.Error("watch error", "err", err)
		}
	}
}

func (w *Watcher) handleEvent(ev fsnotify.Event) {
	if ev.Op&fsnotify.Create != 0 {
		if info, err := os.Stat(ev.Name); err == nil && info.IsDir() {
			if _, err := addRecursive(w.fsw, ev.Name); err != nil {
				slog.Error("failed to watch new directory", "dir", ev.Name, "err", err)
			}
		}
	}

	if w.relevant(ev.Name) {
		w.debounce()
	}
}

// relevant reports whether a changed path is inside one of the trees this
// Watcher cares about. Membership is checked explicitly (rather than by
// excluding the config dir) so a config file that happens to live inside the
// input dir or theme dir still triggers correctly.
func (w *Watcher) relevant(name string) bool {
	if within(w.inputDir, name) {
		return true
	}

	w.mu.Lock()
	themeDir := w.themeDir
	w.mu.Unlock()
	if within(themeDir, name) {
		return true
	}

	return filepath.Dir(name) == w.configDir && filepath.Base(name) == w.configName
}

func within(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func (w *Watcher) debounce() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.timer != nil {
		w.timer.Stop()
	}
	w.timer = time.AfterFunc(debounceInterval, w.onChange)
}

// UpdateThemeDir stops watching the previous theme directory and starts
// watching newThemeDir, called after a rebuild picks up a config change that
// switched the active theme.
func (w *Watcher) UpdateThemeDir(newThemeDir string) error {
	w.mu.Lock()
	oldDirs := w.themeDirs
	w.mu.Unlock()

	for _, d := range oldDirs {
		_ = w.fsw.Remove(d) // best-effort: dir may already be gone
	}

	newDirs, err := addRecursive(w.fsw, newThemeDir)
	if err != nil {
		return err
	}

	w.mu.Lock()
	w.themeDirs = newDirs
	w.themeDir = newThemeDir
	w.mu.Unlock()

	return nil
}

func (w *Watcher) Close() error {
	return w.fsw.Close()
}

// addRecursive walks root and adds a watch for every directory in the tree,
// since fsnotify does not watch recursively on its own.
func addRecursive(fsw *fsnotify.Watcher, root string) ([]string, error) {
	var dirs []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if addErr := fsw.Add(path); addErr != nil {
				return addErr
			}
			dirs = append(dirs, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dirs, nil
}
