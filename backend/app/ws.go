package app

import (
	"net/http"
	"prometheus/config"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Hub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
}

var DemoHub = &Hub{clients: map[*websocket.Conn]struct{}{}}

func (h *Hub) add(c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = struct{}{}
}

func (h *Hub) remove(c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, c)
	c.Close()
}

func (h *Hub) Broadcast(event string) {
	h.mu.Lock()
	conns := make([]*websocket.Conn, 0, len(h.clients))
	for c := range h.clients {
		conns = append(conns, c)
	}
	h.mu.Unlock()

	payload := []byte(`{"event":"` + event + `"}`)
	for _, c := range conns {
		if err := c.WriteMessage(websocket.TextMessage, payload); err != nil {
			h.remove(c)
		}
	}
}

func (app *App) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	DemoHub.add(conn)
	go func() {
		defer DemoHub.remove(conn)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
}

func (app *App) ModeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if config.DemoMode {
		w.Write([]byte(`{"demo":true}`))
	} else {
		w.Write([]byte(`{"demo":false}`))
	}
}
