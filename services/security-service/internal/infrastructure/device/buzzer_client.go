package device

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"aurora/services/security-service/internal/domain"
)

// BuzzerClient implementa domain.BuzzerClient para o ESP32
type BuzzerClient struct {
	mu       sync.RWMutex
	devices  map[string]*domain.AlarmDevice
	fallback string
	apiKey   string
}

// NewBuzzerClient cria um novo cliente de buzzer
func NewBuzzerClient(fallbackIP, apiKey string) *BuzzerClient {
	return &BuzzerClient{
		devices:  make(map[string]*domain.AlarmDevice),
		fallback: fallbackIP,
		apiKey:   apiKey,
	}
}

// Register registra um dispositivo
func (c *BuzzerClient) Register(device *domain.AlarmDevice) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.devices[device.ID] = device
}

func (c *BuzzerClient) baseURL(deviceID string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if d, ok := c.devices[deviceID]; ok {
		return fmt.Sprintf("http://%s:%d", d.IPAddress, d.Port)
	}
	if c.fallback != "" {
		return fmt.Sprintf("http://%s", c.fallback)
	}
	return "http://localhost"
}

func (c *BuzzerClient) get(url string) error {
	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Device-Key", c.apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("device unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("device returned %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// TurnOn ativa o buzzer
func (c *BuzzerClient) TurnOn(deviceID string) error {
	url := c.baseURL(deviceID) + "/buzzer/on"
	log.Printf("[BuzzerClient] TurnOn → %s", url)
	return c.get(url)
}

// TurnOff desativa o buzzer
func (c *BuzzerClient) TurnOff(deviceID string) error {
	url := c.baseURL(deviceID) + "/buzzer/off"
	log.Printf("[BuzzerClient] TurnOff → %s", url)
	return c.get(url)
}
