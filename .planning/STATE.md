---
gsd_state_version: 1.0
milestone: v1.9.1
milestone_name: milestone
current_phase: 01
status: unknown
stopped_at: Completed 01-02-PLAN.md
last_updated: "2026-03-26T11:37:04.785Z"
progress:
  total_phases: 1
  completed_phases: 0
  total_plans: 4
  completed_plans: 2
---

# Project State: Aurora — schedule-service

**Last updated:** 2026-03-26
**Current phase:** 01
**Stopped at:** Completed 01-02-PLAN.md

## Project Reference

See: .planning/PROJECT.md (updated 2026-03-25)

**Core value:** Agendamentos confiáveis que executam as ações certas no horário certo, mesmo após restart do serviço.
**Current focus:** Phase 01 — dominio-schema (plan 3/4 complete)

## Current Status

| Phase | Name | Status |
|-------|------|--------|
| 1 | Domínio + Schema | In Progress (3/4 plans done) |
| 2 | Camada de Aplicação | Not started |
| 3 | Infraestrutura — Repositório + API REST | Not started |
| 4 | Scheduler Core + Integração | Not started |

## Accumulated Context

### Roadmap Evolution

- Projeto inicializado em 2026-03-25
- Roadmap criado com 4 fases cobrindo 28 requisitos v1
- schedule-service será o 7º microsserviço do sistema Aurora
- 01-03 concluído em 2026-03-26: migração PostgreSQL criada com todos os campos v1

### Key Decisions

- **gocron/v2** escolhido sobre robfig/cron — suporte nativo a OneTimeJob e agendamentos recorrentes
- **NATS pattern**: seguir `sensors-service` (não `rules-service` — tem bug de reconexão)
- **Timezone**: campo na migração desde o início (Phase 1), mesmo que só ativado em v2
- **Atomicidade one-shot**: `completed_at` setado na mesma transação que o dispatch
- **PostgreSQL advisory lock** antes de iniciar o loop gocron
- **Todos os 18 campos da tabela schedules baked in desde o início** — evita ALTER TABLE retroativos
- **ON DELETE CASCADE** em schedule_executions — limpeza automática ao deletar agendamento

### Research Findings

- Research completa em `.planning/research/` (STACK, FEATURES, ARCHITECTURE, PITFALLS, SUMMARY)
- Pitfall crítico: timezone (Docker roda UTC, usuários BR = UTC-3)
- Pitfall crítico: goroutine leak — usar worker pool limitado com context timeout
- Pitfall crítico: execuções perdidas no restart — recovery check obrigatório no startup

## Open Questions

- `silence_alarm` HTTP endpoint URL precisa ser verificado no security-service antes de implementar HTTPActionClient
- gocron/v2 `OneTimeJob` com `next_run_at` persistence — verificar API antes de codificar Phase 4
- PostgreSQL advisory lock key value precisa ser escolhido e documentado
