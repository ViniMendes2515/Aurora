package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// LightStateEvent é a mensagem enviada a todos os clientes WS quando uma luz muda
type LightStateEvent struct {
	Type     string `json:"type"`      // "light_state_changed"
	LightID  string `json:"light_id"`
	Name     string `json:"name"`
	Location string `json:"location"`
	State    string `json:"state"` // "on" | "off"
}

// Hub gerencia todas as conexões WebSocket ativas
type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]struct{}

	upgrader websocket.Upgrader
}

// NewHub cria um novo Hub
func NewHub() *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]struct{}),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// ServeWS faz o upgrade HTTP → WS e registra o cliente
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] Upgrade error: %v", err)
		return
	}

	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()

	log.Printf("[WS] Client connected (%d total)", h.clientCount())

	// Goroutine que detecta desconexão
	go func() {
		defer func() {
			h.mu.Lock()
			delete(h.clients, conn)
			h.mu.Unlock()
			conn.Close()
			log.Printf("[WS] Client disconnected (%d remaining)", h.clientCount())
		}()
		for {
			// Lê mensagens (apenas para detectar fechamento)
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}

// Broadcast envia um evento de mudança de estado para todos os clientes conectados
func (h *Hub) Broadcast(event LightStateEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("[WS] Failed to marshal event: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("[WS] Failed to send to client: %v", err)
		}
	}
}

// BroadcastState implementa application.StatePublisher
func (h *Hub) BroadcastState(lightID, name, location, state string) {
	h.Broadcast(LightStateEvent{
		Type:     "light_state_changed",
		LightID:  lightID,
		Name:     name,
		Location: location,
		State:    state,
	})
}

func (h *Hub) clientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
