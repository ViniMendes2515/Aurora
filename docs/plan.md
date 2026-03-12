(Opcional, se der tempo) Adicionar um Value Object no domínio

Por exemplo no auth-service: tipo Email com validação, usado na entidade User em vez de string
Ou no rules-service: tipo RuleID/TriggerID para mostrar encapsulamento de invariantes

(Opcional) Documentação Swagger/OpenAPI

Para pelo menos um serviço (ex: rules-service ou auth-service), adicionar anotações Swaggo e endpoint /swagger/index.html com Gin