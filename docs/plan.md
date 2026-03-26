# Plano do TCC — Aurora

## Contexto do que já está pronto
- 6 serviços com DDD bem aplicado (domain, application, infrastructure)
- ESP32 real integrado (PIR, LDR, 4 LEDs, buzzer)
- Frontend Angular funcionando
- Testes de domínio em todos os serviços
- `TriggerSchedule` existe no domínio mas **nunca foi implementado**

---

## Etapa 1 — Segurança ✅
Credenciais removidas do `firmware/aurora_esp32/config.h`. Arquivo `config.example.h` criado com placeholders. `firmware/.gitignore` configurado para ignorar `config.h` nos próximos commits.

---

## Etapa 2 — Value Objects no domínio✅
**Alto impacto acadêmico, baixo esforço (~1h)**

O DDD define que conceitos com regras de negócio próprias devem ser tipos, não strings primitivas.

**O que fazer:**
- `auth-service/internal/domain/`: criar tipo `Email` com validação encapsulada, substituir `string` na entidade `User`
- `rules-service/internal/domain/`: criar tipo `ScheduleTime` (formato `HH:MM`) para o campo que surgirá na Etapa 4

**Por que importa no TCC:** demonstra invariantes de domínio encapsuladas — exatamente o que diferencia DDD de CRUD.

---

## Etapa 3 — Testes da camada de Application
**Alto impacto acadêmico, esforço médio (~3h)**

Atualmente só o domain layer tem testes. A camada de application é testável porque as dependências são interfaces — mas nenhum teste explora isso.

**O que fazer:**
- `auth-service`: `auth_service_test.go` com mock de `UserRepository` — testar `Register` (email duplicado, hash de senha) e `Login` (credenciais inválidas)
- `rules-service`: `rules_engine_test.go` com mock de `RuleRepository` — testar `EvaluateMotionTrigger` (regra desabilitada não executa, auto-off é agendado)

**Por que importa no TCC:** demonstra que a arquitetura em camadas viabiliza testes isolados por camada — argumento central do DDD aplicado.

---

## Etapa 4 — Implementar `TriggerSchedule` no rules-service
**O diferencial do TCC, esforço alto (~1 dia)**

O `TriggerType = "schedule"` já existe no domínio mas nunca foi implementado. Completar o ciclo: criar regras de agendamento que executem ações em horários definidos.

**O que fazer:**
- Domínio: usar `ScheduleTime` (criado na Etapa 2) como tipo do campo de horário
- Migration: `ALTER TABLE rules ADD COLUMN schedule_time VARCHAR(5)` e `ADD COLUMN schedule_days VARCHAR(50)` (ex: `"mon,tue,wed"`)
- Application: goroutine com ticker que avalia regras de agendamento a cada minuto
- Frontend: campo de hora no formulário de criação/edição de regra

---

## Etapa 5 — Documentação técnica
**Necessário para TCC, esforço baixo (~2h)**

- Atualizar este arquivo (`docs/plan.md`) para um documento de arquitetura real: decisões de design, bounded contexts, fluxo de eventos NATS
- Swagger em pelo menos um serviço (`rules-service` ou `auth-service`) usando `swaggo/swag`

---

## Ordem de execução

| # | Etapa | Prioridade | Estimativa | Status |
|---|-------|-----------|------------|--------|
| 1 | Credenciais no config.h | Urgente | ~15min | ✅ feito |
| 2 | Value Objects (Email, ScheduleTime) | Alta | ~1h | ✅ feito |
| 3 | Testes de Application layer | Alta | ~3h | feito |
| 4 | Implementar TriggerSchedule | Alta | ~1 dia | feito |
| 5 | Documentação / Swagger | Média | ~2h | pendente |
| 6 | Notificacao Telegram | Média | ~2h | pendente |
