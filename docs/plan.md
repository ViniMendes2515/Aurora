Atualizar o README principal do repositório
Corrigir: todos os serviços são Go (não Rust)
Documentar portas 8080 a 8085 e presença do API Gateway
Incluir diagrama ou descrição dos Bounded Contexts e comunicação (NATS, HTTP)
Mencionar uso do Gin nos serviços e nginx como gateway

(Opcional, se der tempo) Adicionar um Value Object no domínio

Por exemplo no auth-service: tipo Email com validação, usado na entidade User em vez de string
Ou no rules-service: tipo RuleID/TriggerID para mostrar encapsulamento de invariantes

(Opcional) Documentação Swagger/OpenAPI

Para pelo menos um serviço (ex: rules-service ou auth-service), adicionar anotações Swaggo e endpoint /swagger/index.html com Gin