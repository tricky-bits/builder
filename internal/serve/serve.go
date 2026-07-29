// Package serve implements `builder serve`: build once into a scratch
// directory, serve it locally, and rebuild (with a livereload push) whenever
// the config file, input dir, or active theme dir change.
package serve

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/tricky-bits/builder/internal/builder"
)

// Options configures the serve command.
type Options struct {
	ConfigPath string
	Bind       string
	Port       int
}

// Run builds once, serves the result locally, and rebuilds on changes until
// ctx is cancelled. A failing first build returns an error immediately
// (nothing is served or watched); a failing later rebuild is logged and the
// last good build keeps being served.
func Run(ctx context.Context, opts Options) error {
	outDir, basePath, cfg, err := buildOnce(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("initial build: %w", err)
	}

	state := newSiteState()
	state.swap(outDir, basePath)

	bc := newBroadcaster()

	addr := net.JoinHostPort(opts.Bind, strconv.Itoa(opts.Port))
	srv := &http.Server{
		Addr:    addr,
		Handler: newHandler(state, bc),
	}

	go func() {
		slog.Info("serving", "url", "http://"+addr+basePath+"/")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("http server error", "err", err)
		}
	}()

	themeDir := filepath.Join(cfg.Build.ThemesDir, cfg.Build.Theme)

	w, err := newWatcher(cfg.Build.InputDir, opts.ConfigPath, themeDir)
	if err != nil {
		shutdown(srv, nil, outDir)
		return fmt.Errorf("start watcher: %w", err)
	}

	w.Start(func() {
		newOut, newBase, newCfg, err := buildOnce(opts.ConfigPath)
		if err != nil {
			slog.Error("rebuild failed, keeping previous build", "err", err)
			return
		}

		if prevDir := state.swap(newOut, newBase); prevDir != "" {
			os.RemoveAll(prevDir)
		}
		bc.Broadcast([]byte("reload"))
		slog.Info("rebuilt", "dir", newOut)

		if newThemeDir := filepath.Join(newCfg.Build.ThemesDir, newCfg.Build.Theme); newThemeDir != themeDir {
			if err := w.UpdateThemeDir(newThemeDir); err != nil {
				slog.Error("failed to update theme watch", "err", err)
			} else {
				themeDir = newThemeDir
			}
		}
	})

	<-ctx.Done()

	dir, _ := state.snapshot()
	shutdown(srv, w, dir)

	return nil
}

func shutdown(srv *http.Server, w *Watcher, activeDir string) {
	shCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shCtx)

	if w != nil {
		w.Close()
	}

	if activeDir != "" {
		os.RemoveAll(activeDir)
	}
}

// buildOnce re-reads the config file from disk, builds into a fresh temp
// directory (never touching the configured output_dir), and returns where it
// landed.
func buildOnce(configPath string) (string, string, *builder.Config, error) {
	cfg, err := builder.ReadConfigFile(configPath)
	if err != nil {
		return "", "", nil, err
	}

	outDir, err := os.MkdirTemp("", "builder-serve-*")
	if err != nil {
		return "", "", nil, err
	}

	cfg.Build.OutputDir = outDir

	b, err := builder.New(cfg)
	if err != nil {
		os.RemoveAll(outDir)
		return "", "", nil, err
	}
	if err := b.Load(); err != nil {
		os.RemoveAll(outDir)
		return "", "", nil, err
	}
	if err := b.Build(); err != nil {
		os.RemoveAll(outDir)
		return "", "", nil, err
	}

	return outDir, cfg.Site.BasePath, cfg, nil
}
