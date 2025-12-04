package messaging

import (
	"log"

	"github.com/nats-io/nats.go"
)

// NATSConnection gerencia a conexão com o NATS
type NATSConnection struct {
	conn *nats.Conn
}

// NewNATSConnection cria uma nova conexão com o NATS
func NewNATSConnection(url string) (*NATSConnection, error) {
	conn, err := nats.Connect(url,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(10),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Printf("NATS disconnected: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("NATS reconnected to %s", nc.ConnectedUrl())
		}),
	)
	if err != nil {
		return nil, err
	}

	log.Printf("Connected to NATS at %s", url)
	return &NATSConnection{conn: conn}, nil
}

// GetConnection retorna a conexão NATS
func (c *NATSConnection) GetConnection() *nats.Conn {
	return c.conn
}

// Close fecha a conexão
func (c *NATSConnection) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
