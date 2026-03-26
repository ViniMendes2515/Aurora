# Project State: Aurora — schedule-service

**Last updated:** 2026-03-25
**Current phase:** Não iniciado — pronto para planejamento

## Project Reference

See: .planning/PROJECT.md (updated 2026-03-25)

**Core value:** Agendamentos confiáveis que executam as ações certas no horário certo, mesmo após restart do serviço.
**Current focus:** Inicialização completa — próximo passo: Fase 1

## Current Status

| Phase | Name | Status |
|-------|------|--------|
| 1 | Domínio + Schema | Not started |
| 2 | Camada de Aplicação | Not started |
| 3 | Infraestrutura — Repositório + API REST | Not started |
| 4 | Scheduler Core + Integração | Not started |

## Accumulated Context

### Roadmap Evolution

- Projeto inicializado em 2026-03-25
- Roadmap criado com 4 fases cobrindo 28 requisitos v1
- schedule-service será o 7º microsserviço do sistema Aurora

### Key Decisions

- **gocron/v2** escolhido sobre robfig/cron — suporte nativo a OneTimeJob e agendamentos recorrentes
- **NATS pattern**: seguir `sensors-service` (não `rules-service` — tem bug de reconexão)
- **Timezone**: campo na migração desde o início (Phase 1), mesmo que só ativado em v2
- **Atomicidade one-shot**: `completed_at` setado na mesma transação que o dispatch
- **PostgreSQL advisory lock** antes de iniciar o loop gocron

### Research Findings

- Research completa em `.planning/research/` (STACK, FEATURES, ARCHITECTURE, PITFALLS, SUMMARY)
- Pitfall crítico: timezone (Docker roda UTC, usuários BR = UTC-3)
- Pitfall crítico: goroutine leak — usar worker pool limitado com context timeout
- Pitfall crítico: execuções perdidas no restart — recovery check obrigatório no startup

## Open Questions

- `silence_alarm` HTTP endpoint URL precisa ser verificado no security-service antes de implementar HTTPActionClient
- gocron/v2 `OneTimeJob` com `next_run_at` persistence — verificar API antes de codificar Phase 4
- PostgreSQL advisory lock key value precisa ser escolhido e documentado
