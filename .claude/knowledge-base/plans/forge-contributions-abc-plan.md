# Plan: Forge contributions A+B+C — eliminar 1 duplicação + adicionar 2 presets úteis

> **Version 1.1** — Implementa 3 contribuições identificadas na auditoria de uso Theo↔forge: (A) Theo descarta `ptr` próprio em favor de `forge.Ptr`; (B) forge ganha `NewRetryOnFailure` factory eliminando duplicação de configuração comum; (C) forge ganha `ResourcesTiny/Small/Medium` presets T-shirt-size que Theo já usa. Resultado: 1 helper a menos manter na Theo + 2 features novas no forge sem breaking change, com Theo migrando para consumi-las após release.
>
> v1.1 incorpora 4 MUST FIX de `/edge-case-plan` (ver `## v1.1 Changelog` ao final): rename para evitar colisão com const `RetryOnFailure`, fetch+rebase antes de migrar Theo, alocação fresh de maps/slices em Resources presets, e baseline de coverage antes da release.

## Context

Auditoria executada em 2026-05-20 (sessão atual) revelou:

1. **Duplicação real**: Theo `helpers.go:18` define `ptr[T any](v T) *T` idêntico a `forge.Ptr[T any]` (build_helpers.go:221). Theo adicionou antes de forge ter o helper genérico (T1.1 do plano v0.5.0). Manter as duas formas é débito técnico — mesma assinatura, mesmo propósito.

2. **Padrão comum não-genérico**: Theo `languages.go:535` define `retryStrategy(limit int, duration, factor, maxDuration string) *forge.RetryStrategy` que monta um RetryStrategy com:
   - `Limit: ptr(limit)`
   - `RetryPolicy: "OnFailure"`
   - `Backoff{Duration, Factor, MaxDuration}`
   - `Expression: 'asInt(lastRetry.exitCode) != 137'` (skip-OOM)

   Esse pattern é usado em 5+ sites na Theo (delivery, syft, cosign, buildkit, harbor-mirror). Qualquer outro consumidor do forge com workflows de build vai re-implementar a mesma factory. Forge não oferece atalho.

3. **T-shirt size resources não-genéricas**: Theo `languages.go:551-570` define `resourcesTiny()` (50m/32Mi), `resourcesSmall()` (100m/128Mi), `resourcesMedium()` (500m/512Mi). Padrão convencional em build pipelines Kubernetes; forge não oferece presets, força cada consumidor a re-declarar.

Evidências documentadas:
- Auditoria 2.9% de uso Theo↔forge (sessão atual)
- Theo `api/internal/argo/helpers.go:18` — duplicação `ptr`
- Theo `api/internal/argo/languages.go:535-570` — factories ausentes no forge
- 19/19 E2E forge contra Argo v4.0.3 PASS (linha-base)
- 37/37 packages Theo PASS após bump forge v0.5.0-pseudo (linha-base)

## Objective

**Done = Theo deleta `ptr`, `retryStrategy`, `resourcesTiny/Small/Medium` e usa exclusivamente `forge.Ptr`, `forge.NewRetryOnFailure`, `forge.Resources{Tiny,Small,Medium}`; todos os testes verdes em ambos repos; backward-compat preservada para outros consumidores do forge.**

Goals específicos e mensuráveis:

1. Forge ganha 4 novos símbolos exportados (`NewRetryOnFailure`, `ResourcesTiny`, `ResourcesSmall`, `ResourcesMedium`) com testes unitários e godocs.
2. Forge `make verify` continua ALL QUALITY GATES PASSED.
3. Theo `helpers.go` perde a função `ptr` (callers usam `forge.Ptr`).
4. Theo `languages.go` perde `retryStrategy`, `resourcesTiny`, `resourcesSmall`, `resourcesMedium` (callers usam factories forge).
5. Theo `go test ./...` continua 37/37 PASS, zero comportamento alterado.
6. PRs separados em cada repo, revisáveis isoladamente.

## ADRs

### ADR-001 — `NewRetryOnFailure` é factory function, não builder

| Campo | Valor |
|-------|-------|
| **ID** | D1 |
| **Decisão** | Forge expõe `func NewRetryOnFailure(limit int, initial, factor, maxDuration string) *RetryStrategy`. Assinatura espelha exatamente o `retryStrategy()` da Theo (mesma ordem de parâmetros). Retorna `*RetryStrategy` populado com `RetryPolicy: "OnFailure"`, `Backoff` montado e expressão skip-OOM (`asInt(lastRetry.exitCode) != 137`). |
| **Rationale** | KISS (cita `~/.claude/CLAUDE.md §10` e `.claude/rules/clean-code.md §Function size`). Factory de 1 chamada é mais simples que builder com 4 métodos. Theo já validou o shape em produção (5 sites consumidores). Skip-OOM expression é convenção razoável; consumidores que querem outro behavior podem ainda usar `&RetryStrategy{...}` literal. |
| **Alternativas** | (a) Builder fluente `forge.Retry().OnFailure(limit).Backoff(initial, factor, max).SkipOOM().Build()` — overengineering para 1 caso de uso; YAGNI (rejeitada). (b) Variádicas com options pattern `NewRetryOnFailure(limit, opts ...RetryOpt)` — adiciona indireção sem ganho para 4 parâmetros estáveis (rejeitada). (c) **Chosen**: factory direta. |
| **Consequences** | Habilita: Theo (e outros consumidores) eliminam 14 linhas de função+ducplicação. Constrange: a expressão skip-OOM vira parte do "default" — consumidores com semântica diferente precisam usar struct literal (documentado). |

### ADR-002 — Resources presets são funções `()` retornando `*ResourceRequirements`, não constantes

| Campo | Valor |
|-------|-------|
| **ID** | D2 |
| **Decisão** | Forge expõe três factories sem parâmetros: `ResourcesTiny()`, `ResourcesSmall()`, `ResourcesMedium()`. Cada uma retorna `*ResourceRequirements` recém-alocado com valores T-shirt-size pré-definidos (Tiny=50m/32Mi-100m/64Mi; Small=100m/128Mi-500m/256Mi; Medium=500m/512Mi-2000m/2Gi). |
| **Rationale** | `~/.claude/CLAUDE.md §10` (KISS) + `.claude/rules/architecture.md §SOLID OCP` — funções permitem evolução interna sem quebrar API. Constants `var TinyResources = &ResourceRequirements{...}` seriam imutáveis em teoria mas Go permite mutação via ponteiro (consumer modifica acidentalmente, contamina próximos chamadores). Factory garante nova alocação por chamada. |
| **Alternativas** | (a) Constants exportadas — risco de aliasing mutável (rejeitada por DRY/safety). (b) `Resources(preset string)` com enum — typo-prone (rejeitada). (c) Single `Resources(preset ResourceSize)` typed enum — adiciona um type para 3 valores; YAGNI (rejeitada). (d) **Chosen**: 3 factory functions independentes. |
| **Consequences** | Habilita: presets reutilizáveis sem risco de mutação cruzada. Constrange: 3 funções em vez de 1 — mas trivial e auto-documentado. Sizes podem ser ajustados internamente sem breaking change. |

### ADR-003 — Theo trocará helpers APÓS forge release tag estável

| Campo | Valor |
|-------|-------|
| **ID** | D3 |
| **Decisão** | Sequência mandatória: (1) PR forge merge → main + tag `v0.5.1`; (2) PR Theo bump go.mod para `v0.5.1` E migra helpers no MESMO PR. PR Theo NÃO usa pseudo-version. |
| **Rationale** | `.claude/rules/dependencies.md §When adding a new dependency is allowed` — pseudo-versions são frágeis (não-reproduzíveis em CI sem rebaixamento). Tag estável evita "trabalha na minha máquina". Sequência também garante reverter Theo é trivial (revert do bump + dos call sites). |
| **Alternativas** | (a) Theo bumpa direto via pseudo-version pós-merge — mais rápido mas dívida do tag pendente (rejeitada). (b) Single mega-PR com forge + Theo em uma branch fictícia — viola separação de repos (rejeitada). (c) **Chosen**: 2 PRs, ordem rígida. |
| **Consequences** | Habilita: rastreabilidade limpa via SemVer. Constrange: 5-10 min entre os 2 PRs (aceitável). |

### ADR-004 — Backward compat: structs literais continuam suportados

| Campo | Valor |
|-------|-------|
| **ID** | D4 |
| **Decisão** | As 4 funções novas (`NewRetryOnFailure`, 3× `Resources*`) são **additive** — não removem nem deprecam APIs existentes. Outros consumidores podem continuar usando `&RetryStrategy{...}` e `&ResourceRequirements{...}` literais. |
| **Rationale** | `~/.claude/CLAUDE.md §10` (KISS) e prática SemVer: minor release (v0.5.0 → v0.5.1) NÃO pode quebrar APIs. Forge é library pública — outros consumidores futuros podem ter requisitos de retry/resources diferentes da Theo. |
| **Alternativas** | (a) Deprecar struct literal `RetryStrategy{...}` forçando todo mundo via factory — over-prescriptive (rejeitada). (b) **Chosen**: aditivo, retrocompatível 100%. |
| **Consequences** | Habilita: zero risco de regressão em consumidores não-Theo. Constrange: existe agora "two ways to do it" — mitigado via godoc em `RetryStrategy` apontando para `NewRetryOnFailure` quando aplicável. |

### ADR-005 — Testes obrigatórios: factory + golden YAML

| Campo | Valor |
|-------|-------|
| **ID** | D5 |
| **Decisão** | Cada nova função no forge ganha: (1) teste unitário table-driven validando struct retornada; (2) golden test marshalando o `*ResourceRequirements`/`*RetryStrategy` para YAML e comparando com fixture pré-aprovada. Garante que o output wire-format é o esperado. |
| **Rationale** | `.claude/rules/testing.md §Regression-first + Table-driven` exige TDD. Golden test prova que factory produz YAML idêntico ao struct literal — facilita migração da Theo (assertion ponta a ponta). |
| **Alternativas** | (a) Apenas testes de struct — não pega regressão de tag/serialização (rejeitada). (b) **Chosen**: struct + golden YAML. |
| **Consequences** | Habilita: confiança alta na migração. Constrange: ~80 LoC de testes adicionais (aceitável). |

## Dependency Graph

```
Phase 1 (forge code) ─────▶ Phase 2 (forge release v0.5.1) ─────▶ Phase 3 (Theo migration)
       │                              │                                    │
       ▼                              ▼                                    ▼
   T1.1 NewRetryOnFailure          T2.1 tag v0.5.1                T3.1 bump go.mod
   T1.2 Resources presets       T2.2 GitHub release notes      T3.2 migrate `ptr`→Ptr
   T1.3 Tests + golden          (manual: maintainer)           T3.3 migrate retryStrategy
                                                               T3.4 migrate resources*
                                                                    │
                                                                    ▼
                                                            Phase 4 (Dogfood QA)
                                                            T4.1 make verify forge
                                                            T4.2 go test Theo
                                                            T4.3 forge E2E v4.0.3
```

Phase 1 tasks T1.1, T1.2, T1.3 podem rodar em paralelo (arquivos independentes). Phases 2, 3, 4 são sequencialmente dependentes (release antes de consumir; migrar antes de validar).

---

## Phase 1: Forge — adicionar 4 factories

**Objective:** Adicionar `NewRetryOnFailure`, `ResourcesTiny`, `ResourcesSmall`, `ResourcesMedium` ao forge, com testes unitários + golden YAML.

### T1.1 — `forge.NewRetryOnFailure` factory

#### Objective
Adicionar função pública `NewRetryOnFailure(limit int, initial, factor, maxDuration string) *RetryStrategy` em arquivo dedicado, com expressão skip-OOM embutida.

#### Evidence
- Theo `api/internal/argo/languages.go:535-549` define exatamente esta função internamente
- 5+ call sites na Theo (delivery, syft, cosign keybased, cosign keyless, harbor-mirror) consomem
- ADR-001 (este plano)

#### Files to edit

```
parameter.go — adicionar função NewRetryOnFailure logo após RetryStrategy.Build (linha ~123)
parameter_test.go — adicionar TestNewRetryOnFailure_* table-driven (RED antes do GREEN)
```

**Por que `parameter.go` e não arquivo novo:** `RetryStrategy` já mora em `parameter.go` (junto com `Parameter` + `RetryStrategy.Build()`). Adicionar a factory na mesma file preserva coesão. Arquivo atual tem ~140 LoC, +30 não chega perto do budget de 500 LoC (`.claude/rules/size-allowlist.txt`).

#### Deep file dependency analysis

- `parameter.go`: contém `Parameter`, `RetryStrategy`, `Backoff`. Adicionar uma function `NewRetryOnFailure` aqui é coesa. Downstream: `Build()` continua intacta; consumidores via struct literal continuam funcionando (ADR-004). Nenhum import novo necessário.
- `parameter_test.go`: novo bloco de testes; não afeta os existentes.

#### Deep Dives

- **Assinatura**: `func NewRetryOnFailure(limit int, initial, factor, maxDuration string) *RetryStrategy`. Parâmetros ordenados como em Theo (limit, initial-duration, factor, max-duration) para compatibilidade de leitura. Retorna `*RetryStrategy` (não value) porque o campo consumer (`Workflow.RetryStrategy`/`Container.RetryStrategy`) é `*RetryStrategy`.
- **Default policy**: `RetryPolicy: "OnFailure"` hardcoded — refletido no nome da factory.
- **Skip-OOM expression**: `asInt(lastRetry.exitCode) != 137` — exit 137 é SIGKILL via OOM. Argo expression DSL. Aceito como default razoável; documentar em godoc.
- **Edge cases**:
  - `limit < 0` → função aceita; caller responsabilidade (Argo rejeita server-side se inválido).
  - Strings vazias para backoff → emite `Backoff{Duration: "", ...}` que vira `{}` no YAML (Argo aplica defaults). Aceitável.

#### Tasks

1. **EC-4 — Baseline coverage**: antes de tocar o arquivo, rodar `make verify-coverage` e anotar coverage atual de `parameter.go` (esperado ~85.7%). A nova função MUST não reduzir esse número.
2. Em `parameter.go`, adicionar a função `NewRetryOnFailure` com godoc explicando parâmetros + skip-OOM expression.
3. Em `parameter_test.go`, adicionar `TestNewRetryOnFailure_PopulatesAllFields` (RED).
4. Em `parameter_test.go`, adicionar `TestNewRetryOnFailure_SkipOOMExpression` validando a expression literal (RED).
5. Em `parameter_test.go`, adicionar `TestNewRetryOnFailure_BuildProducesValidModel` chamando `.Build()` no resultado (RED).
6. **EC-4 — Cobrir cases não-cobertos de `Backoff.Factor`** (TODO `code-p4-backoff-factor-partial-normalization` no review.db): adicionar `TestBackoffFactor_Float64Normalization` e `TestBackoffFactor_PointerFloat64`.
7. Rodar `go test -count=1 ./...` confirmando RED → GREEN.
8. Rodar `make verify` confirmando lint, vet, race, coverage **≥ baseline anotada no passo 1**.

#### TDD

```
RED:     TestNewRetryOnFailure_PopulatesAllFields — asserts Limit, RetryPolicy=="OnFailure",
         Backoff.{Duration,Factor,MaxDuration}, Expression no exato shape.
RED:     TestNewRetryOnFailure_SkipOOMExpression — exact string match em "asInt(lastRetry.exitCode) != 137".
RED:     TestNewRetryOnFailure_BuildProducesValidModel — chama .Build() e valida RetryStrategyModel
         tem todos os campos serializados para YAML/JSON.
RED:     TestNewRetryOnFailure_TableDriven — 3 casos: limit=0, limit=2, limit=10; cada um valida Limit=ptr(N).
GREEN:   Implementar a função em parameter.go com a assinatura ADR-001.
REFACTOR: Confirmar que godoc tem exemplo executável (`// Example: forge.NewRetryOnFailure(2, "10s", "2", "120s")`).
VERIFY:  go test -count=1 -race ./... && make verify
```

#### Acceptance Criteria

- [ ] `forge.NewRetryOnFailure` existe, exportada, com godoc completo.
- [ ] Godoc cita exit 137 + skip-OOM rationale.
- [ ] 4 testes (3 unit + 1 build-roundtrip) passam.
- [ ] Coverage do `parameter.go` mantém ≥80% (.testcoverage.yml threshold).
- [ ] Pass: `make verify-lint` zero issues.
- [ ] Pass: `make verify-coverage` total ≥80%.
- [ ] Pass: `make verify-fmt`, `make verify-vet`, `make verify-race`.
- [ ] File `parameter.go` ≤500 LoC (`.claude/rules/size-allowlist.txt`).

#### DoD (Definition of Done)

- [ ] Tarefas 1–6 concluídas.
- [ ] Testes verdes.
- [ ] `make verify` ALL QUALITY GATES PASSED.
- [ ] Godoc renderiza limpo (`go doc github.com/usetheodev/theo-forge`).
- [ ] Backward compat: tests existentes do `RetryStrategy{...}` struct literal continuam passando.

---

### T1.2 — `forge.ResourcesTiny/Small/Medium` presets

#### Objective

Adicionar 3 factories: `ResourcesTiny()`, `ResourcesSmall()`, `ResourcesMedium()`, cada uma retornando `*ResourceRequirements` recém-alocado.

#### Evidence

- Theo `api/internal/argo/languages.go:551-570` define exatamente estes 3 presets internamente.
- Consumidos em ~6 sites na Theo.
- ADR-002 (este plano).

#### Files to edit

```
types.go — adicionar 3 funções factory + 3 constants documentando os valores (linha ~80)
types_test.go — novo arquivo: testes table-driven dos 3 presets + golden YAML
```

**Por que `types.go`:** já contém `ResourceRequirements`, `ResourceList`, `ImagePullPolicy`. Coesão. Tamanho atual ~100 LoC, +50 não estoura budget.

#### Deep file dependency analysis

- `types.go`: type aliases + helpers de tipos comuns. Adicionar factories de Resources aqui é coesa. Downstream: nenhum — tipos existentes intactos.
- `types_test.go` (NEW): testes isolados; sem dependência cruzada.

#### Deep Dives

- **Valores T-shirt** (espelham Theo exatamente):
  - Tiny: requests CPU 50m / Memory 32Mi; limits CPU 100m / Memory 64Mi.
  - Small: requests 100m / 128Mi; limits 500m / 256Mi.
  - Medium: requests 500m / 512Mi; limits 2000m / 2Gi.
- **Aliasing safety (EC-3)**: cada chamada retorna nova instância — INCLUSIVE para campos internos como `ResourceList` (que tem maps no futuro se evoluir). Implementação MUST usar struct literal completo dentro da factory, sem variáveis-pacote-level (`var tinyResources = ...`) que viriam aliased entre callers. Garantido por `TestResourcesTiny_MutationIsolation`: cria 2 instâncias, muta a primeira, verifica a segunda inalterada. Documentado em godoc — "Each call returns a freshly-allocated struct including any nested maps/slices; safe to mutate by caller without affecting other callers."
- **Sem `Large/XLarge`**: YAGNI. Adicionar quando houver caso de uso concreto.

#### Tasks

1. Em `types.go`, adicionar `ResourcesTiny`, `ResourcesSmall`, `ResourcesMedium` com godocs explicando os valores.
2. Em `types_test.go` (NEW), adicionar `TestResources*_AllocatesFresh` validando que duas chamadas retornam ponteiros distintos (RED).
3. Em `types_test.go`, adicionar `TestResources*_RequestLessThanLimit` validando invariantes (RED).
4. Em `types_test.go`, adicionar `TestResourcesPresets_GoldenYAML` (table-driven) marshalando cada preset e comparando com fixtures (RED).
5. Criar `testdata/resources-tiny.golden.yaml`, `testdata/resources-small.golden.yaml`, `testdata/resources-medium.golden.yaml` (use flag `-update-golden`).
6. Rodar `go test -count=1 -race ./...` → GREEN.
7. Rodar `make verify`.

#### TDD

```
RED:     TestResourcesTiny_AllocatesFresh — duas chamadas, asserts ponteiros !=.
RED:     TestResourcesSmall_AllocatesFresh — idem.
RED:     TestResourcesMedium_AllocatesFresh — idem.
RED:     TestResourcesTiny_MutationIsolation (EC-3) — cria duas instâncias do MESMO preset,
         muta CPU/Memory na primeira, valida segunda inalterada (nested maps/slices também).
RED:     TestResourcesTiny_RequestLessThanLimit — CPU/Memory request <= limit.
RED:     TestResourcesSmall_RequestLessThanLimit — idem.
RED:     TestResourcesMedium_RequestLessThanLimit — idem.
RED:     TestResourcesPresets_GoldenYAML — table-driven, 3 casos, byte-compare contra testdata/*.golden.yaml.
GREEN:   Implementar as 3 factories.
REFACTOR: Verificar que godoc dos 3 presets tem mesma estrutura (consistência).
VERIFY:  go test -count=1 -race ./... && make verify
```

#### Acceptance Criteria

- [ ] `forge.ResourcesTiny/Small/Medium` existem, exportadas, godoc completo.
- [ ] 7 testes passam (3 fresh + 3 invariant + 1 golden table-driven).
- [ ] 3 golden YAML files em `testdata/`.
- [ ] Pass: `make verify-lint` zero issues.
- [ ] Pass: `make verify-coverage`.
- [ ] File `types.go` ≤500 LoC.
- [ ] File `types_test.go` ≤1000 LoC (test budget conforme `.claude/rules/size-allowlist.txt`).

#### DoD

- [ ] Tarefas 1–7 concluídas.
- [ ] Testes verdes.
- [ ] `make verify` ALL QUALITY GATES PASSED.
- [ ] Godoc renderiza limpo.

---

### T1.3 — CHANGELOG + godoc cross-references

#### Objective

Documentar as 4 novas APIs no CHANGELOG `[Unreleased]` e adicionar godoc-links cruzados entre `RetryStrategy` ↔ `NewRetryOnFailure` e `ResourceRequirements` ↔ presets.

#### Evidence

- `~/.claude/CLAUDE.md §6 Changelogs` exige toda mudança em `[Unreleased]` com referência ao PR/issue.
- ADR-004 menciona godoc cross-reference para mitigar "two ways to do it".

#### Files to edit

```
CHANGELOG.md — adicionar bloco Added sob [Unreleased]
parameter.go — godoc de RetryStrategy menciona NewRetryOnFailure
types.go — godoc de ResourceRequirements menciona os 3 presets
```

#### Deep file dependency analysis

- `CHANGELOG.md`: append-only sob `[Unreleased]`. Sem risco.
- `parameter.go`, `types.go`: apenas comentários. Sem mudança comportamental.

#### Tasks

1. Em `CHANGELOG.md` `[Unreleased]` `### Added`: adicionar 4 entries citando ADR-001/ADR-002.
2. Em `parameter.go`, no godoc de `RetryStrategy`, adicionar: `// For the common "retry-on-failure with skip-OOM" pattern, see [NewRetryOnFailure].`
3. Em `types.go`, no godoc de `ResourceRequirements`, adicionar referência aos 3 presets.

#### TDD

```
RED:     N/A (mudança puramente documental).
GREEN:   Editar CHANGELOG e godocs.
VERIFY:  go doc github.com/usetheodev/theo-forge | grep -E "NewRetryOnFailure|ResourcesTiny|ResourcesSmall|ResourcesMedium"
         confirma os 4 símbolos aparecem.
         grep "NewRetryOnFailure" parameter.go confirma cross-ref no godoc.
```

#### Acceptance Criteria

- [ ] CHANGELOG tem 4 linhas novas sob `[Unreleased] ### Added`.
- [ ] Godoc de `RetryStrategy` referencia `NewRetryOnFailure`.
- [ ] Godoc de `ResourceRequirements` referencia os 3 presets.
- [ ] Pass: `make verify-lint` (lint inclui godot que valida comments end-with-period).

#### DoD

- [ ] Tarefas 1–3 concluídas.
- [ ] `make verify` PASSED.
- [ ] PR description (a ser escrito em T2.1) usa as entries do CHANGELOG como base.

---

## Phase 2: Forge release v0.5.1

**Objective:** Cortar tag `v0.5.1` estável + GitHub release com notas; preparar consumidor (Theo) para bump.

### T2.1 — PR no forge: merge para main + tag v0.5.1

#### Objective

Abrir PR no forge consolidando Phase 1, merge após review, tag `v0.5.1`, push do tag, GitHub release com notas extraídas do CHANGELOG.

#### Evidence

- ADR-003 (este plano) — Theo NÃO bumpa antes do tag estável.
- `.claude/rules/dependencies.md` exige pinning estável para reprodutibilidade.

#### Files to edit

```
(nenhum no working tree — operações git/gh exclusivamente)
```

Comandos:

```bash
cd /home/paulo/theo-forge
git checkout -b feat/retry-resources-factories main
# (Phase 1 já committado nesta branch)
git push -u origin feat/retry-resources-factories
gh pr create --base main --title "feat: NewRetryOnFailure factory + Resources T-shirt presets" \
  --body "Closes A+B+C contributions per plan forge-contributions-abc-plan.md"
# após merge:
git checkout main && git pull
git tag -a v0.5.1 -m "v0.5.1: NewRetryOnFailure factory + Resources T-shirt presets"
git push origin v0.5.1
gh release create v0.5.1 --notes-from-tag
```

#### Deep file dependency analysis

N/A — operação git/release.

#### Deep Dives

- **Semver**: v0.5.0 → v0.5.1 (MINOR bump) porque mudança é additive (ADR-004). Patch (v0.5.0 → v0.5.0.1) seria insuficiente porque adiciona API pública.
- **Tag annotated**: `-a v0.5.1 -m "..."` para que `go install @v0.5.1` resolva determinístico.
- **Release notes**: extraídas automaticamente do CHANGELOG `[Unreleased]` no momento da promoção para `[0.5.1] - YYYY-MM-DD`.

#### Tasks

1. Verificar `git status` limpo na branch da Phase 1.
2. Push da branch.
3. `gh pr create` com PR description citando ADR-001, ADR-002, ADR-004.
4. Aguardar review/merge (gate humano — se sou maintainer, autoapprove).
5. Após merge: `git pull main && git tag -a v0.5.1 -m ...`.
6. `git push origin v0.5.1`.
7. `gh release create v0.5.1 --notes-from-tag`.
8. Promover CHANGELOG `[Unreleased]` → `[0.5.1] - YYYY-MM-DD` (commit separado pós-tag).

#### TDD

```
RED:     N/A (release operation).
GREEN:   Tag pushed, release published.
VERIFY:  go list -m github.com/usetheodev/theo-forge@v0.5.1 resolve sem erro.
         gh release view v0.5.1 mostra notas.
```

#### Acceptance Criteria

- [ ] PR aprovado e merged em `main`.
- [ ] Tag `v0.5.1` anotado e pushed.
- [ ] GitHub release publicado com notas.
- [ ] `go list -m github.com/usetheodev/theo-forge@v0.5.1` retorna sem erro.
- [ ] CHANGELOG `[0.5.1]` carimbado com data.

#### DoD

- [ ] Tarefas 1–8 concluídas.
- [ ] Tag visível em `git tag --list`.
- [ ] Release page acessível via gh CLI.

---

## Phase 3: Theo PR — migrar 3 helpers para forge

**Objective:** PR no `/home/paulo/theo` que bumpa forge para v0.5.1 e substitui `ptr`, `retryStrategy`, `resourcesTiny/Small/Medium` por funções do forge. Zero comportamento alterado.

Trabalho realizado em worktree isolado `/home/paulo/theo-script-pr` (já existente) ou novo worktree para evitar conflito com outro agente em `develop`.

### T3.1 — Bump forge para v0.5.1

#### Objective

Atualizar `api/go.mod` para `github.com/usetheodev/theo-forge v0.5.1` e validar build + suite.

#### Evidence

- ADR-003 (consumir tag estável, não pseudo-version).
- Plano forge-contributions-abc-plan.md (este).

#### Files to edit

```
/home/paulo/theo-script-pr/api/go.mod — bump
/home/paulo/theo-script-pr/api/go.sum — atualizado por go get
/home/paulo/theo-script-pr/go.work.sum — atualizado por go work sync
```

(Caminhos no worktree isolado; o outro agente continua em `/home/paulo/theo` sem interferência.)

#### Deep file dependency analysis

- `api/go.mod`: 1 linha alterada (`v0.4.1-0.20260520173502-87c647f0c071` → `v0.5.1`).
- `api/go.sum`: regenerado por `go get`.
- `go.work.sum`: regenerado por `go work sync`.
- Downstream: todos os 10 arquivos `api/internal/argo/*.go` que importam `forge` continuam funcionando (símbolos antigos intactos por ADR-004).

#### Deep Dives

- **Possível conflito `genproto`**: já resolvido no PR atual (`chore/forge-script-migration` commit `b23bf231`). Se ainda não foi merged em develop, este PR depende dele OU re-aplica a mesma técnica.
- **Backward compat**: forge v0.5.1 é additive sobre v0.5.0; todos os símbolos usados pela Theo continuam funcionando.

#### Tasks

1. **EC-2 — Refresh do worktree antes de qualquer mudança**:
   `cd /home/paulo/theo-script-pr && git fetch origin && git rebase origin/develop` (resolve qualquer divergência com o outro agente trabalhando em `develop`; se conflito, parar e escalar antes de prosseguir).
2. `cd /home/paulo/theo-script-pr/api`
3. `go get github.com/usetheodev/theo-forge@v0.5.1`
4. `go mod tidy`
5. `cd /home/paulo/theo-script-pr && go work sync`
6. `cd api && go build ./...` (must succeed)
7. `go test -count=1 ./internal/argo/...` (must pass)

#### TDD

```
RED:     N/A (bump deps).
GREEN:   Build + test pass com nova versão.
VERIFY:  grep "theo-forge v0.5.1" api/go.mod
         go test ./... 37/37 PASS.
```

#### Acceptance Criteria

- [ ] `api/go.mod` mostra `github.com/usetheodev/theo-forge v0.5.1` (tag estável, sem pseudo).
- [ ] `go build ./...` PASS para todos os módulos do workspace.
- [ ] `go test -count=1 ./internal/argo/...` PASS.
- [ ] Pass: zero novas warnings de lint.

#### DoD

- [ ] Tarefas 1–6 concluídas.
- [ ] Suite Theo continua 37/37 PASS.

---

### T3.2 — Migrar `ptr` → `forge.Ptr` na Theo

#### Objective

Deletar a função `ptr[T any]` em `api/internal/argo/helpers.go:18` e atualizar todos os call sites para usar `forge.Ptr`.

#### Evidence

- Auditoria sessão atual: `forge.Ptr` exporta `func Ptr[T any](v T) *T` idêntico.
- Theo `helpers.go:18` define função privada equivalente.

#### Files to edit

```
api/internal/argo/helpers.go — DELETAR função ptr (linhas 16-18)
api/internal/argo/workflow_builder.go — call sites: substituir ptr( → forge.Ptr(
api/internal/argo/cosign.go — idem
api/internal/argo/syft.go — idem
api/internal/argo/delivery.go — idem
api/internal/argo/buildkit.go — idem
api/internal/argo/languages.go — idem
api/internal/argo/dockerfile_lint.go — idem
api/internal/argo/harbor_mirror.go — idem
```

#### Deep file dependency analysis

- `helpers.go`: -3 linhas (função deletada). Demais helpers intactos.
- Outros arquivos: substituição literal `ptr(` → `forge.Ptr(`. Cada arquivo já importa `forge` (verificado em sessão anterior).
- Downstream: zero — `forge.Ptr` tem mesma assinatura.

#### Deep Dives

- **Conflito de nomes**: `ptr` é palavra curta — poderia colidir com variável local. Verificar via `grep -nE "\bptr\b\s*[^(]"` antes de fazer substituição cega.
- **Generics signature**: Go 1.18+ — `forge.Ptr[T any](v T) *T`. O tipo é inferido pelo argumento. Não preciso forçar `forge.Ptr[int](42)`; `forge.Ptr(42)` é suficiente.

#### Tasks

1. `grep -rn "\bptr(" api/internal/argo/ | grep -v "_test.go"` — inventário de call sites.
2. Para cada arquivo: substituir `ptr(` por `forge.Ptr(` (sed-safe pois `ptr(` é assinatura única).
3. Deletar a função `ptr` em `helpers.go` (linhas 16-18).
4. `go build ./...` (must compile).
5. `go test -count=1 ./internal/argo/...` (must pass).
6. Confirmar `grep -n "\bptr\b" api/internal/argo/*.go | grep -v "forge\.Ptr"` retorna vazio.

#### TDD

```
RED:     N/A (refactor; tests existentes JÁ cobrem o comportamento — usar como regression net).
GREEN:   Suite continua passando após migração.
VERIFY:  go test -count=1 -race ./internal/argo/...
         grep -n "func ptr\[" api/internal/argo/ retorna vazio (função removida).
         grep "forge.Ptr" api/internal/argo/*.go | wc -l > 0 (uso novo).
```

#### Acceptance Criteria

- [ ] Função `ptr` em `helpers.go` deletada.
- [ ] Todos os call sites usam `forge.Ptr`.
- [ ] `grep "\bptr(" api/internal/argo/ | grep -v "_test.go"` retorna 0 matches (exceto false positives nomeados como `ptr` mas que não são a função).
- [ ] Suite Theo 37/37 PASS.
- [ ] `go vet ./...` clean.

#### DoD

- [ ] Tarefas 1–6 concluídas.
- [ ] Diff revisável: -3 linhas helpers.go, +N linhas em call sites (substituição literal).

---

### T3.3 — Migrar `retryStrategy` → `forge.NewRetryOnFailure`

#### Objective

Deletar `retryStrategy(limit, duration, factor, maxDuration string) *forge.RetryStrategy` em `languages.go:535` e atualizar call sites para `forge.NewRetryOnFailure`.

#### Evidence

- ADR-001 (este plano) — `forge.NewRetryOnFailure` tem assinatura idêntica.
- 5 call sites na Theo.

#### Files to edit

```
api/internal/argo/languages.go — DELETAR função retryStrategy (linhas 535-549)
api/internal/argo/delivery.go — call site: retryStrategy(... → forge.NewRetryOnFailure(...
api/internal/argo/syft.go — idem
api/internal/argo/cosign.go — 2 call sites
api/internal/argo/buildkit.go — idem
api/internal/argo/harbor_mirror.go — idem
```

#### Deep file dependency analysis

- `languages.go`: -15 linhas (função+godoc deletada). Demais helpers intactos.
- Call sites: substituição literal `retryStrategy(` → `forge.NewRetryOnFailure(`. Parâmetros na mesma ordem (validado em ADR-001).

#### Deep Dives

- **Expressão skip-OOM hardcoded**: Theo `retryStrategy` embute `asInt(lastRetry.exitCode) != 137`. `forge.NewRetryOnFailure` faz idem (ADR-001). Migração mantém semântica.
- **Side effects**: nenhum — função pura, retorna struct fresh.

#### Tasks

1. `grep -rn "retryStrategy(" api/internal/argo/ | grep -v "_test.go"` — inventário (5 sites esperados).
2. Para cada arquivo, substituir `retryStrategy(` por `forge.NewRetryOnFailure(`.
3. Deletar a função `retryStrategy` + godoc em `languages.go:535-549`.
4. `go build ./...`.
5. `go test -count=1 ./internal/argo/...`.
6. Confirmar `grep "retryStrategy" api/internal/argo/*.go` zero matches.

#### TDD

```
RED:     N/A (refactor).
GREEN:   Suite passa.
VERIFY:  go test -count=1 -race ./internal/argo/...
         grep "func retryStrategy" api/internal/argo/ vazio.
         grep "forge.NewRetryOnFailure" api/internal/argo/ ≥5 matches.
```

#### Acceptance Criteria

- [ ] Função `retryStrategy` deletada.
- [ ] Todos os call sites usam `forge.NewRetryOnFailure`.
- [ ] Suite Theo PASS.
- [ ] YAML emitido inalterado (validar via test que compara golden Workflow).

#### DoD

- [ ] Tarefas 1–6 concluídas.
- [ ] Argo-template tests passam sem alteração (regressão net).

---

### T3.4 — Migrar `resourcesTiny/Small/Medium` → `forge.Resources*`

#### Objective

Deletar as 3 funções `resourcesTiny/Small/Medium` em `languages.go:551-570` e atualizar call sites.

#### Evidence

- ADR-002 (este plano) — assinatura idêntica.
- 6+ call sites na Theo.

#### Files to edit

```
api/internal/argo/languages.go — DELETAR 3 funções (linhas 551-570)
api/internal/argo/delivery.go — Resources usage
api/internal/argo/syft.go — idem
api/internal/argo/cosign.go — idem
api/internal/argo/buildkit.go — idem
api/internal/argo/dockerfile_lint.go — idem
api/internal/argo/harbor_mirror.go — idem
```

#### Deep file dependency analysis

- `languages.go`: -20 linhas (3 funções deletadas).
- Call sites: substituição literal `resourcesTiny()` → `forge.ResourcesTiny()`, idem Small e Medium.

#### Tasks

1. `grep -rn "resources\(Tiny\|Small\|Medium\)" api/internal/argo/ | grep -v "_test.go"` — inventário.
2. Substituições literais.
3. Deletar 3 funções em `languages.go`.
4. `go build && go test ./internal/argo/...`.
5. Confirmar zero matches do nome antigo.

#### TDD

```
RED:     N/A.
GREEN:   Suite PASS.
VERIFY:  go test -count=1 -race ./internal/argo/...
         grep "func resourcesTiny\|func resourcesSmall\|func resourcesMedium" vazio.
```

#### Acceptance Criteria

- [ ] 3 funções deletadas em `languages.go`.
- [ ] Call sites migrados para `forge.Resources*`.
- [ ] Suite Theo PASS.

#### DoD

- [ ] Tarefas 1–5 concluídas.

---

### T3.5 — Commit + push + abrir PR Theo

#### Objective

Criar 2 commits (bump + migração), push branch, abrir PR `develop` ← `chore/use-forge-helpers`.

#### Evidence

- ADR-003 — sequência: forge ship → Theo bump+migrate.

#### Files to edit

N/A — operações git.

#### Tasks

1. Verificar `git status` no worktree isolado.
2. `git add api/go.mod api/go.sum go.work.sum` + commit `chore(api): bump theo-forge to v0.5.1`.
3. `git add api/internal/argo/` + commit `refactor(argo): use forge.Ptr/NewRetryOnFailure/Resources presets (eliminate 4 duplicate helpers)`.
4. `git push -u origin chore/use-forge-helpers`.
5. `gh pr create --base develop --title "..." --body "..."` (referenciando este plano + ADR-001/002/003).

#### TDD

N/A.

#### Acceptance Criteria

- [ ] 2 commits separados na branch.
- [ ] PR aberto contra `develop`.
- [ ] Diff stat: ~10 LoC em `go.mod/sum`, ~30 LoC em call sites, -38 LoC em deleted helpers.

#### DoD

- [ ] Tarefas 1–5 concluídas.
- [ ] PR URL retornado.

---

## Phase 4: Dogfood QA

**Objective:** Validar que A+B+C funciona em ambos repos como uma usuária real veria.

### T4.1 — `make verify` no forge

#### Objective

Confirmar que o forge `main` pós-v0.5.1 passa ALL QUALITY GATES.

#### Tasks

1. `cd /home/paulo/theo-forge && git checkout main && git pull`
2. `make verify`
3. Confirmar saída `==> ✅ verify: ALL QUALITY GATES PASSED`.

#### Acceptance Criteria

- [ ] Output literal `ALL QUALITY GATES PASSED`.
- [ ] Coverage ≥80% total (atual 91.2%).
- [ ] Zero lint, zero CVEs, race detector clean.

#### DoD

- [ ] `make verify` exit code 0.

---

### T4.2 — Suite full Theo

#### Objective

Confirmar que Theo `chore/use-forge-helpers` branch passa `go test ./...` 37/37.

#### Tasks

1. `cd /home/paulo/theo-script-pr/api && go test -count=1 ./...`
2. Contar `ok` lines (esperado 37) e `FAIL` (esperado 0).

#### Acceptance Criteria

- [ ] 37/37 packages PASS.
- [ ] Zero FAIL.

#### DoD

- [ ] Comando exit code 0.

---

### T4.3 — Forge E2E v4.0.3 contra Argo real

#### Objective

Re-rodar os 19 testes E2E do forge contra Argo Workflows v4.0.3 (versão de produção Theo) para garantir que as novas factories não introduziram regressão sutil.

#### Tasks

1. `cd /home/paulo/theo-forge && make e2e-up`
2. `make e2e` (executa 19 tests, ~6 min)
3. `make e2e-down` (cleanup)

#### Acceptance Criteria

- [ ] 19/19 E2E PASS contra Argo v4.0.3.
- [ ] Output contém `ok github.com/usetheodev/theo-forge/e2e <duration>`.

#### DoD

- [ ] Suite E2E exit code 0.
- [ ] Cluster kind destruído.

---

## Coverage Matrix

| # | Requirement / Gap | Task(s) | Resolution |
|---|---|---|---|
| 1 | Duplicação `Theo.ptr` ↔ `forge.Ptr` | T3.2 | Theo deleta `ptr`, usa `forge.Ptr` |
| 2 | Forge não tem retry factory; Theo reinventou | T1.1 + T3.3 | Forge adiciona `NewRetryOnFailure`; Theo migra |
| 3 | Forge não tem resource presets; Theo reinventou 3 | T1.2 + T3.4 | Forge adiciona 3 factories; Theo migra |
| 4 | Documentação cruzada para evitar "two ways to do it" | T1.3 | Godocs apontam factories como atalho para structs |
| 5 | Release estável obrigatório antes do consumo | T2.1 | Tag v0.5.1 + GitHub release |
| 6 | Backward compat para outros consumidores forge | ADR-004 + T1.1/T1.2 | Mudanças aditivas; struct literais intactos |
| 7 | Validação ponta-a-ponta sem regressão | T4.1/T4.2/T4.3 | make verify forge + go test Theo + 19 E2E |
| 8 | Cobertura testes nas novas factories ≥80% | T1.1/T1.2 TDD blocks | 11 testes + 3 golden YAMLs |
| 9 | Zero conflito com outro agente em develop | T3.1 | Worktree separado `/home/paulo/theo-script-pr` + EC-2 fetch+rebase antes de tocar |
| 10 | Rastreabilidade no CHANGELOG | T1.3 + T2.1 | Entries em `[Unreleased]` promovidos a `[0.5.1]` |

**Coverage: 10/10 gaps covered (100%)**

## Global Definition of Done

- [ ] Todas as Phases 1–4 concluídas.
- [ ] Forge `make verify` ALL QUALITY GATES PASSED após v0.5.1.
- [ ] Theo `go test ./...` 37/37 PASS após migração.
- [ ] Forge 19/19 E2E PASS contra Argo v4.0.3.
- [ ] Zero clippy/lint warnings em ambos repos.
- [ ] Backward compat: nenhum consumidor antigo do forge (que use struct literal) quebra — validado em CI do forge.
- [ ] PR forge merged + tag v0.5.1 publicado.
- [ ] PR Theo aberto contra `develop` (await maintainer merge — fora do escopo deste plano).
- [ ] CHANGELOG forge promovido `[Unreleased]` → `[0.5.1] - YYYY-MM-DD`.
- [ ] Plano de migração documentado no PR body de ambos repos referenciando este arquivo.
- [ ] **Dogfood QA PASS** — `/dogfood full` no forge ≥70 health, zero CRITICAL.
- [ ] **Runtime-metric proof** — N/A (sem métricas runtime novas; presets são puros factories).

## Final Phase: Dogfood QA (MANDATORY)

> Roda APÓS Phases 1–4 estarem completas. Plano NÃO é done sem `/dogfood` passing.

### Execution

```bash
cd /home/paulo/theo-forge && /dogfood full
```

(Se `/dogfood` skill não estiver disponível no forge, usar substituto: `make verify && make e2e` como proxy de "dogfood" — ambos exercitam a biblioteca como um usuário real consumindo.)

### Acceptance Criteria

- [ ] Health score ≥ 70/100 (ou `make verify` + `make e2e` ambos exit 0).
- [ ] Zero CRITICAL introduzidos por A+B+C.
- [ ] Zero HIGH em comandos/features modificados (RetryStrategy, ResourceRequirements, parameter.go, types.go).
- [ ] Qualquer issue pre-existente documentado como não-relacionado.

### If Dogfood Fails

1. Identificar issues do plano vs pré-existentes.
2. Fixar CRITICAL e HIGH causados por A+B+C.
3. Re-rodar.
4. Pré-existentes NÃO bloqueiam — só logged.

---

## v1.1 Changelog

Absorvidos 4 MUST FIX de `/edge-case-plan` em `.claude/knowledge-base/reviews/edge-case-review-forge-contributions-abc.md`. Todas as mudanças são pequenas (rename + 1 sub-task adicional cada).

| EC | Tarefa afetada | Família | Mudança |
|----|---------------|---------|---------|
| EC-1 | T1.1, T1.3, T3.3, ADR-001, CHANGELOG, Coverage Matrix | Naming | `RetryOnFailure` (factory) **colidiria** com const `RetryOnFailure` em `types.go:112` (re-export de `model.RetryOnFailure`). Renomeado para **`NewRetryOnFailure`** em todo o plano. Idiomatic Go: `New<Type>` prefix sinaliza factory. |
| EC-2 | T3.1 | Concurrency | Adicionado passo 1: `git fetch && git rebase origin/develop` antes de tocar o worktree Theo. Evita conflito com outro agente em `develop`. |
| EC-3 | T1.2 Deep Dives + TDD | State / Aliasing | Garantia explícita de alocação fresh INCLUSIVE de campos internos (maps/slices futuros); novo teste `TestResourcesTiny_MutationIsolation` que valida 2 instâncias não compartilham estado mutável. |
| EC-4 | T1.1 Tasks + T4.1 | Coverage gate | Baseline coverage snapshot ANTES da edição; gate "≥ baseline" pós-edição. Plus: aproveitar PR para cobrir `Backoff.Factor` float64/`*float64` cases já no `review.db` como TODO (-1 finding ao mesmo tempo). |

**Coverage update**: 10/10 + 4/4 EC = 14/14 itens cobertos (100%).

