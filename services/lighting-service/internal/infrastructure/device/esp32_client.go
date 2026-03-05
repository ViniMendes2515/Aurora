package device

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"aurora/services/lighting-service/internal/domain"
)

// ESP32Client implementa domain.DeviceClient para o ESP32
type ESP32Client struct {
	mu       sync.RWMutex
	devices  map[string]*domain.Device // deviceID → Device
	fallback string                    // IP fallback (env ESP32_IP)
	apiKey   string
}

// NewESP32Client cria um novo cliente ESP32
func NewESP32Client(fallbackIP, apiKey string) *ESP32Client {
	return &ESP32Client{
		devices:  make(map[string]*domain.Device),
		fallback: fallbackIP,
		apiKey:   apiKey,
	}
}

// Register registra ou atualiza um dispositivo
func (c *ESP32Client) Register(device *domain.Device) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.devices[device.ID] = device
	log.Printf("[ESP32Client] Dispositivo registrado: %s @ %s:%d", device.ID, device.IPAddress, device.Port)
	return nil
}

// FindByID retorna um dispositivo pelo ID
func (c *ESP32Client) FindByID(id string) (*domain.Device, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if d, ok := c.devices[id]; ok {
		return d, nil
	}
	return nil, fmt.Errorf("device not found: %s", id)
}

// parseDeviceID separa "esp32-main:2" em deviceID="esp32-main" e ledPin="2".
// Se não houver ":", retorna pin vazio (compatibilidade com registros antigos).
func parseDeviceID(raw string) (deviceID, ledPin string) {
	parts := strings.SplitN(raw, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return raw, ""
}

func (c *ESP32Client) baseURL(rawDeviceID string) string {
	deviceID, _ := parseDeviceID(rawDeviceID)
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

func (c *ESP32Client) get(url string) error {
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

// ledPath retorna o path correto para o LED, ex: "/led/2/on".
// Se não houver pin definido, usa "/led/1/on" como fallback.
func ledPath(rawDeviceID, action string) string {
	_, pin := parseDeviceID(rawDeviceID)
	if pin == "" {
		pin = "1"
	}
	return fmt.Sprintf("/led/%s/%s", pin, action)
}

// TurnOn envia comando de ligar para o ESP32
func (c *ESP32Client) TurnOn(deviceID string) error {
	url := c.baseURL(deviceID) + ledPath(deviceID, "on")
	log.Printf("[ESP32Client] TurnOn → %s", url)
	return c.get(url)
}

// TurnOff envia comando de apagar para o ESP32
func (c *ESP32Client) TurnOff(deviceID string) error {
	url := c.baseURL(deviceID) + ledPath(deviceID, "off")
	log.Printf("[ESP32Client] TurnOff → %s", url)
	return c.get(url)
}

// GetStatus consulta o estado atual do LED no ESP32
func (c *ESP32Client) GetStatus(deviceID string) (domain.LightState, error) {
	url := c.baseURL(deviceID) + ledPath(deviceID, "status")
	client := &http.Client{Timeout: 3 * time.Second}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return domain.LightStateOff, err
	}
	req.Header.Set("X-Device-Key", c.apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return domain.LightStateOff, domain.ErrDeviceUnreachable
	}
	defer resp.Body.Close()

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return domain.LightStateOff, err
	}

	if body.Status == "on" {
		return domain.LightStateOn, nil
	}
	return domain.LightStateOff, nil
}
