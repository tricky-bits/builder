package serve

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// repoRoot resolves the module root from this test file's location, since
// buildOnce/Run resolve config-relative paths (input_dir, themes_dir) against
// the process's working directory, and the real example/ fixture's paths
// assume that's the repo root.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Join(filepath.Dir(file), "..", "..")
}

func TestBuildOnce_Success(t *testing.T) {
	t.Chdir(repoRoot(t))

	outDir, basePath, cfg, err := buildOnce("example/config.toml")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(outDir) })

	assert.DirExists(t, outDir)
	assert.FileExists(t, filepath.Join(outDir, "index.html"))
	assert.Equal(t, "", basePath) // base_path "/" collapses to ""
	assert.NotNil(t, cfg)
}

func TestBuildOnce_BadConfigPath(t *testing.T) {
	_, _, _, err := buildOnce(filepath.Join(t.TempDir(), "missing.toml"))
	assert.Error(t, err)
}

func TestRun_BadConfigPathReturnsErrorImmediately(t *testing.T) {
	err := Run(context.Background(), Options{
		ConfigPath: filepath.Join(t.TempDir(), "missing.toml"),
		Bind:       "127.0.0.1",
		Port:       0,
	})
	assert.Error(t, err)
}

func TestRun_ServesAndShutsDownCleanly(t *testing.T) {
	t.Chdir(repoRoot(t))

	const port = 18173 // fixed, unlikely-to-collide port for this one test

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- Run(ctx, Options{ConfigPath: "example/config.toml", Bind: "127.0.0.1", Port: port})
	}()

	url := fmt.Sprintf("http://127.0.0.1:%d/", port)
	var resp *http.Response
	var getErr error
	require.Eventually(t, func() bool {
		resp, getErr = http.Get(url)
		return getErr == nil
	}, 3*time.Second, 50*time.Millisecond)
	require.NoError(t, getErr)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	cancel()

	select {
	case runErr := <-done:
		assert.NoError(t, runErr)
	case <-time.After(3 * time.Second):
		t.Fatal("Run did not return after context cancellation")
	}
}
