package domain

// EventPublisher define o contrato para publicação de eventos
// Esta interface pertence ao domínio e será implementada na infraestrutura
type EventPublisher interface {
	// PublishMotionEvent publica um evento de movimento detectado
	PublishMotionEvent(event *MotionDetectedEvent) error

	// PublishLightEvent publica um evento de luminosidade
	PublishLightEvent(event *LightLevelEvent) error
}
