package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"aurora/services/schedule-service/internal/domain"
)

// HTTPActionClient despacha ações de agendamento via HTTP para os serviços de dispositivos
type HTTPActionClient struct {
	lightingURL  string
	securityURL  string
	deviceAPIKey string
	client       *http.Client
}

// NewHTTPActionClient cria um novo HTTPActionClient com timeout de 3 segundos
func NewHTTPActionClient(lightingURL, securityURL, deviceAPIKey string) *HTTPActionClient {
	return &HTTPActionClient{
		lightingURL:  lightingURL,
		securityURL:  securityURL,
		deviceAPIKey: deviceAPIKey,
		client:       &http.Client{Timeout: 3 * time.Second},
	}
}

// Dispatch executa a ação do agendamento realizando uma requisição HTTP ao serviço correspondente
func (c *HTTPActionClient) Dispatch(ctx context.Context, schedule *domain.Schedule) error {
	deviceID := schedule.Action.DeviceID

	var url string
	switch schedule.Action.Type {
	case domain.ActionTurnOnLight:
		url = fmt.Sprintf("%s/lights/%s/on", c.lightingURL, deviceID)
	case domain.ActionTurnOffLight:
		url = fmt.Sprintf("%s/lights/%s/off", c.lightingURL, deviceID)
	case domain.ActionTriggerAlarm:
		url = fmt.Sprintf("%s/alarms/trigger", c.securityURL)
	case domain.ActionSilenceAlarm:
		url = fmt.Sprintf("%s/alarms/%s/silence", c.securityURL, deviceID)
	default:
		return domain.ErrInvalidAction
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, http.NoBody)
	if err != nil {
		return err
	}

	req.Header.Set("X-Device-Key", c.deviceAPIKey)

	log.Printf("HTTPActionClient: dispatching action %s to %s", schedule.Action.Type, url)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("action %s returned HTTP %d", schedule.Action.Type, resp.StatusCode)
	}

	return nil
}
