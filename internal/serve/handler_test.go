package serve

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInjectScript(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"body tag", "<html><body>hi</body></html>"},
		{"html fallback", "<html>hi</html>"},
		{"neither tag", "just some text"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := injectScript([]byte(c.body))
			if !bytes.Contains(out, []byte(livereloadScript)) {
				t.Fatalf("script not injected into %q", c.body)
			}
			if !strings.HasPrefix(string(out), c.body[:min(len(c.body), 4)]) {
				t.Fatalf("original content mangled: %q", out)
			}
		})
	}

	// </body> case: script must land before </body>, not after.
	out := injectScript([]byte("<html><body>hi</body></html>"))
	if idx := bytes.Index(out, []byte(livereloadScript)); idx == -1 || idx > bytes.Index(out, []byte("</body>")) {
		t.Fatalf("script not inserted before </body>: %q", out)
	}
}

func TestFileServerHandlerEmptyBasePath(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("root"), 0o644); err != nil {
		t.Fatal(err)
	}

	state := newSiteState()
	state.swap(dir, "")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	fileServerHandler(state).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("basePath=\"\" must not 404 (StripPrefix(\"\") footgun): got %d", rec.Code)
	}
}

func TestFileServerHandlerNonEmptyBasePath(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("under docs"), 0o644); err != nil {
		t.Fatal(err)
	}

	state := newSiteState()
	state.swap(dir, "/docs")

	req := httptest.NewRequest(http.MethodGet, "/docs/", nil)
	rec := httptest.NewRecorder()
	fileServerHandler(state).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for prefixed path, got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	rec2 := httptest.NewRecorder()
	fileServerHandler(state).ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unprefixed path when basePath is set, got %d", rec2.Code)
	}
}

func TestLooksLikeHTML(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/", true},
		{"/foo/", true},
		{"/foo.html", true},
		{"/foo.png", false},
		{"/foo", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, looksLikeHTML(tt.path), tt.path)
	}
}

func TestLivereloadInject_InjectsScriptForHTML(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body>hi</body></html>"))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	livereloadInject(next).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "__builder_livereload")
}

func TestLivereloadInject_PassesThroughNonHTML(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("body { color: red }"))
	})

	req := httptest.NewRequest(http.MethodGet, "/style.css", nil)
	rec := httptest.NewRecorder()
	livereloadInject(next).ServeHTTP(rec, req)

	assert.Equal(t, "body { color: red }", rec.Body.String())
	assert.NotContains(t, rec.Body.String(), "__builder_livereload")
}

func TestBroadcaster_BroadcastsToConnectedClient(t *testing.T) {
	bc := newBroadcaster()
	srv := httptest.NewServer(bc)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	require.Eventually(t, func() bool {
		bc.mu.Lock()
		defer bc.mu.Unlock()
		return len(bc.clients) == 1
	}, time.Second, 10*time.Millisecond, "client never registered")

	bc.Broadcast([]byte("reload"))

	conn.SetReadDeadline(time.Now().Add(time.Second))
	_, msg, err := conn.ReadMessage()
	require.NoError(t, err)
	assert.Equal(t, "reload", string(msg))
}

func TestBroadcaster_ServeHTTP_RejectsNonWebsocketRequest(t *testing.T) {
	bc := newBroadcaster()

	req := httptest.NewRequest(http.MethodGet, "/__builder_livereload", nil)
	rec := httptest.NewRecorder()
	bc.ServeHTTP(rec, req)

	assert.NotEqual(t, http.StatusSwitchingProtocols, rec.Code)
	bc.mu.Lock()
	defer bc.mu.Unlock()
	assert.Empty(t, bc.clients)
}

func TestNewHandler_ServesFileWithInjectedScript(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html><body>hi</body></html>"), 0o644))

	state := newSiteState()
	state.swap(dir, "")

	srv := httptest.NewServer(newHandler(state, newBroadcaster()))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "__builder_livereload")
}
