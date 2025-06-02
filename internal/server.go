package internal

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

type Server struct {
	clients      map[*websocket.Conn]bool
	clientsMutex sync.Mutex
	upgrader     websocket.Upgrader
}

func NewServer() *Server {
	return &Server{
		clients: make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	s.clientsMutex.Lock()
	s.clients[conn] = true
	s.clientsMutex.Unlock()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			s.clientsMutex.Lock()
			delete(s.clients, conn)
			s.clientsMutex.Unlock()
			return
		}
	}
}

func watchDir(dir string, onChange func(), debounceTime time.Duration) {
	var debounceTimer *time.Timer

	resetTimer := func() {
		if debounceTimer != nil {
			debounceTimer.Stop()
		}

		debounceTimer = time.AfterFunc(debounceTime, func() {
			onChange()
		})
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if strings.HasPrefix(filepath.Base(event.Name), ".") {
					break
				}

				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
					resetTimer()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Watcher error:", err)
			}
		}
	}()

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			err = watcher.Add(path)
			if err != nil {
				log.Fatalln(err)
			}
		}
		return nil
	})

	<-done
}

func (s *Server) broadcastReload() {
	s.clientsMutex.Lock()
	for client := range s.clients {
		client.WriteMessage(websocket.TextMessage, []byte("reload"))
	}
	s.clientsMutex.Unlock()
}

const reloadScript = `
<script>
if ('WebSocket' in window) {
	const ws = new WebSocket('ws://' + window.location.host + '/ws');

	ws.onmessage = (event) => {
		if (event.data === "reload") {
			console.log("Reloading page...");
			location.reload();
		}
	};
}
</script>
`

func handlerWithReloadInjection(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join("dist", r.URL.Path)

	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		path = filepath.Join(path, "index.html")
	}

	if strings.HasSuffix(path, ".html") {
		file, err := os.Open(path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer file.Close()

		var buf bytes.Buffer
		_, err = io.Copy(&buf, file)
		if err != nil {
			http.Error(w, "Failed to read file", 500)
			return
		}

		content := strings.Replace(buf.String(), "</body>", reloadScript+"</body>", 1)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(content))
	} else {
		http.ServeFile(w, r, path)
	}
}

func (s *Server) ServeAndWatch(dirToServe string, dirToWatch string, onFileChange func(), quit chan struct{}) {
	go func() {
		http.HandleFunc("/", handlerWithReloadInjection)
		http.HandleFunc("/ws", s.handleWebSocket)
		log.Println("Now serving on http://localhost:3000. Watching for changes...")

		err := http.ListenAndServe(":3000", nil)
		if err != nil {
			log.Fatalln(err)
		}
	}()

	go watchDir(dirToWatch, func() {
		onFileChange()
		s.broadcastReload()
	}, 300*time.Millisecond)

	<-quit
}
