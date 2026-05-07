package domain

// EventPublisher define o contrato para publicação de eventos
type EventPublisher interface {
	// PublishMotionEvent publica um evento de movimento detectado
	PublishMotionEvent(event *MotionDetectedEvent) error

	// PublishLightEvent publica um evento de luminosidade
	PublishLightEvent(event *LightLevelEvent) error
}
