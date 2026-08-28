package handler

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // للسماح بأي عميل Frontend بالاتصال
	},
}

type WebSocketHandler struct {
	clients   map[*websocket.Conn]bool
	broadcast chan any
	mu        sync.RWMutex
}

func NewWebSocketHandler() *WebSocketHandler {
	h := &WebSocketHandler{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan any, 100),
	}
	go h.listen()
	return h
}

// ServeWS يرفع اتصال HTTP العادي إلى اتصال WebSocket مستمر
func (h *WebSocketHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket Upgrade Error: %v", err)
		return
	}

	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, conn)
		h.mu.Unlock()
		_ = conn.Close()
	}()

	// قراءة رسائل Ping/Pong من العميل للحفاظ على الاتصال حياً
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// BroadcastEvent يرسل تحديثاً فورياً لجميع المتصلين (مثلاً عند تعديل تذكرة أو لوحة أجايل)
func (h *WebSocketHandler) BroadcastEvent(event any) {
	h.broadcast <- event
}

func (h *WebSocketHandler) listen() {
	for msg := range h.broadcast {
		h.mu.RLock()
		for client := range h.clients {
			err := client.WriteJSON(msg)
			if err != nil {
				_ = client.Close()
				delete(h.clients, client)
			}
		}
		h.mu.RUnlock()
	}
}
