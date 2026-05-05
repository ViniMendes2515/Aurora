package messaging

import (
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

type lightStateEvent struct {
	LightID   string    `json:"light_id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	State     string    `json:"state"` // "on" | "off"
	ChangedAt time.Time `json:"changed_at"`
}

// NATSLightPublisher publica eventos de mudanca de estado de luz no NATS
type NATSLightPublisher struct {
	conn *nats.Conn
}

func NewNATSLightPublisher(natsURL string) (*NATSLightPublisher, error) {
	var nc *nats.Conn
	var err error
	for i := range 10 {
		nc, err = nats.Connect(natsURL)
		if err == nil {
			break
		}
		log.Printf("[NATS] Aguardando NATS... tentativa %d/10", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, err
	}
	return &NATSLightPublisher{conn: nc}, nil
}

func (p *NATSLightPublisher) PublishLightChanged(lightID, userID, name, location, state string) {
	event := lightStateEvent{
		LightID:   lightID,
		UserID:    userID,
		Name:      name,
		Location:  location,
		State:     state,
		ChangedAt: time.Now(),
	}
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("[NATS] Erro ao serializar lighting.state.changed: %v", err)
		return
	}
	if err := p.conn.Publish("lighting.state.changed", data); err != nil {
		log.Printf("[NATS] Erro ao publicar lighting.state.changed: %v", err)
	}
}

func (p *NATSLightPublisher) Close() {
	p.conn.Close()
}
