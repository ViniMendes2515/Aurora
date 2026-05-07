package domain

// EventPublisher define o contrato para publicação de domain events de segurança.
// A interface fica no domínio e a implementação fica na infraestrutura (inversão de dependência).
type EventPublisher interface {
	PublishAlarmTriggered(event *AlarmTriggeredEvent) error
}
