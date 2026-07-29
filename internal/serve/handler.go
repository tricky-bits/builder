package serve

import (
	"bytes"
	"maps"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

// siteState holds the currently-served output directory and base path,
// swapped atomically after each successful rebuild so in-flight requests
// never see a half-written directory or a dir/basePath pair that doesn't
// belong together.
type siteState struct {
	current atomic.Pointer[siteSnapshot]
}

type siteSnapshot struct {
	dir      string
	basePath string
}

func newSiteState() *siteState {
	s := &siteState{}
	s.current.Store(&siteSnapshot{})
	return s
}

// swap installs a new dir/basePath and returns the previous dir, so the
// caller can remove it once the swap is live.
func (s *siteState) swap(dir, basePath string) string {
	prev := s.current.Swap(&siteSnapshot{dir: dir, basePath: basePath})
	return prev.dir
}

func (s *siteState) snapshot() (string, string) {
	cur := s.current.Load()
	return cur.dir, cur.basePath
}

// fileServerHandler serves the current site dir, honoring Site.BasePath.
//
// basePath == "" must skip http.StripPrefix entirely: StripPrefix("", h)
// is a no-op TrimPrefix, so every request's path stays unchanged length and
// StripPrefix's length check makes it 404 everything. builder.New collapses
// base_path "/" down to "", which is the common case, so this guard matters.
func fileServerHandler(state *siteState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dir, basePath := state.snapshot()
		fs := http.FileServer(http.Dir(dir))
		if basePath == "" {
			fs.ServeHTTP(w, r)
			return
		}
		http.StripPrefix(basePath, fs).ServeHTTP(w, r)
	}
}

const livereloadScript = `<script>(function(){
function connect(){
	var ws=new WebSocket((location.protocol==="https:"?"wss://":"ws://")+location.host+"/__builder_livereload");
	ws.onmessage=function(){location.reload();};
	ws.onclose=function(){setTimeout(connect,1000);};
}
connect();
})();</script>`

func looksLikeHTML(p string) bool {
	return p == "/" || strings.HasSuffix(p, "/") || strings.HasSuffix(p, ".html")
}

// injectScript inserts the livereload script before </body>, falling back to
// </html>, then to a plain append if neither tag is present.
func injectScript(body []byte) []byte {
	script := []byte(livereloadScript)
	for _, tag := range [][]byte{[]byte("</body>"), []byte("</html>")} {
		if idx := bytes.LastIndex(body, tag); idx != -1 {
			out := make([]byte, 0, len(body)+len(script))
			out = append(out, body[:idx]...)
			out = append(out, script...)
			out = append(out, body[idx:]...)
			return out
		}
	}
	return append(body, script...)
}

// livereloadInject buffers HTML responses and injects the livereload script.
// Non-HTML paths pass through untouched to avoid buffering assets.
func livereloadInject(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !looksLikeHTML(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		rec := httptest.NewRecorder()
		next.ServeHTTP(rec, r)

		body := rec.Body.Bytes()
		if rec.Code == http.StatusOK {
			body = injectScript(body)
		}

		maps.Copy(w.Header(), rec.Header())
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		w.WriteHeader(rec.Code)
		w.Write(body)
	})
}

// broadcaster upgrades /__builder_livereload connections and pushes a reload
// message to every connected client after a successful rebuild.
type broadcaster struct {
	mu       sync.Mutex
	clients  map[*websocket.Conn]struct{}
	upgrader websocket.Upgrader
}

func newBroadcaster() *broadcaster {
	return &broadcaster{
		clients: make(map[*websocket.Conn]struct{}),
		upgrader: websocket.Upgrader{
			// Dev server only, bound to localhost by default: no CSRF-style
			// origin risk worth the friction of checking it.
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (b *broadcaster) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := b.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	b.mu.Lock()
	b.clients[conn] = struct{}{}
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		delete(b.clients, conn)
		b.mu.Unlock()
		conn.Close()
	}()

	// Block reading so we notice the client going away; we never expect
	// incoming messages from the injected script.
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (b *broadcaster) Broadcast(msg []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for conn := range b.clients {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			conn.Close()
			delete(b.clients, conn)
		}
	}
}

func newHandler(state *siteState, bc *broadcaster) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/__builder_livereload", bc)
	mux.Handle("/", livereloadInject(fileServerHandler(state)))
	return mux
}
