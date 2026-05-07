package domain

// EventPublisher define o contrato para publicação de domain events de iluminação.
// A interface fica no domínio e a implementação fica na infraestrutura (inversão de dependência).
type EventPublisher interface {
	PublishLightChanged(event *LightStateChangedEvent) error
}
