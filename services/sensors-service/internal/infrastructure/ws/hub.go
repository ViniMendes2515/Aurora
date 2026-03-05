package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// SensorEventType indica o tipo de evento enviado pelo hub
type SensorEventType string

const (
	EventMotionDetected SensorEventType = "motion_detected"
	EventLightUpdated   SensorEventType = "light_updated"
)

// SensorEvent é a mensagem enviada aos clientes WebSocket sempre que
// um novo evento de sensor é registrado (movimento ou luminosidade).
type SensorEvent struct {
	Type      SensorEventType `json:"type"`
	SensorID  string          `json:"sensor_id"`
	SensorType string         `json:"sensor_type"` // "motion" | "light"

	// Para movimento
	DetectedAt string `json:"detected_at,omitempty"`

	// Para luminosidade
	// Importante: não usar omitempty para que o valor 0% também seja enviado.
	Value      float64 `json:"value"`
	Raw        int     `json:"raw"`
	RecordedAt string  `json:"recorded_at"`
}

// Hub gerencia conexões WebSocket para o serviço de sensores.
// Ele não conhece detalhes de NATS/HTTP; apenas recebe SensorEvent e
// os replica para todos os clientes.
type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]struct{}

	upgrader websocket.Upgrader
}

// NewHub cria um novo Hub.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]struct{}),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// ServeWS faz o upgrade HTTP → WebSocket e registra o cliente.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS][sensors] Upgrade error: %v", err)
		return
	}

	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()

	log.Printf("[WS][sensors] Client connected (%d total)", h.clientCount())

	// Goroutine para detectar desconexão
	go func() {
		defer func() {
			h.mu.Lock()
			delete(h.clients, conn)
			h.mu.Unlock()
			_ = conn.Close()
			log.Printf("[WS][sensors] Client disconnected (%d remaining)", h.clientCount())
		}()

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}

// Broadcast envia um evento de sensor para todos os clientes conectados.
func (h *Hub) Broadcast(event SensorEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("[WS][sensors] Failed to marshal event: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("[WS][sensors] Failed to send to client: %v", err)
		}
	}
}

func (h *Hub) clientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

