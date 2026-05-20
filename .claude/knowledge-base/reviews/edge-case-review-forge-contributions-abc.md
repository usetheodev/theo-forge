# Edge Case Review — forge-contributions-abc-plan

Data: 2026-05-20
Plano analisado: `.claude/knowledge-base/plans/forge-contributions-abc-plan.md`
Tasks analisadas: 12 (T1.1, T1.2, T1.3, T2.1, T3.1, T3.2, T3.3, T3.4, T3.5, T4.1, T4.2, T4.3)
Edge cases encontrados: 11 (MUST FIX: 4, SHOULD TEST: 4, DOCUMENT: 3)

---

## MUST FIX

### EC-1: Colisão de nome — `RetryOnFailure` já existe como constante exportada

- **Task afetada:** T1.1 (e cascateia em T1.3, T3.3, todo o release)
- **Família:** Format / Naming
- **Cenário:** `types.go:112` já exporta a constante `RetryOnFailure = model.RetryOnFailure` (valor `RetryPolicy`). Existem ≥3 call sites internos (`workflow_test.go:1056`, `container_test.go:118`, `examples_test.go:1598`) usando `RetryPolicy: RetryOnFailure`. O plano adiciona `func RetryOnFailure(limit int, ...)` no MESMO package — **conflito de nome no espaço do package**: Go não permite uma const e uma func compartilharem identificador.
- **Impacto:** Compilação quebra imediatamente em T1.1 (`RetryOnFailure redeclared in this block`). Bloqueia TODA a Phase 1, 2, 3. Não é um edge — é falha de compilação determinística.
- **Fix sugerido:** Renomear a factory. Opções: `NewRetryOnFailure(...)` (consistente com `NewConfig`), `RetryOnFailureStrategy(...)`, ou `RetryWithBackoff(...)`. Recomendo `NewRetryOnFailure` (idiomatic Go constructor prefix). Atualizar ADR-001 + T1.1/T1.3/T3.3 + CHANGELOG entries + godoc cross-refs.

### EC-2: Sequência forge-release → Theo-bump não tem fallback se review do PR forge demora

- **Task afetada:** T2.1 → T3.1 (Phase 2 gate)
- **Família:** State / Timing
- **Cenário:** ADR-003 proíbe pseudo-version. Se PR forge ficar 2+ dias em review (humano), a worktree `chore/use-forge-helpers` em `/home/paulo/theo-script-pr` fica parada com `helpers.go` modificado mentalmente mas sem código pronto, e o outro agente em `develop` pode mergear conflitos no diretório `api/internal/argo/`.
- **Impacto:** Worktree Theo desatualizada quando Phase 3 começar, gerando merge conflicts em `helpers.go` e `languages.go` (os arquivos mais editados).
- **Fix sugerido:** Adicionar ao plano antes de T3.1: `git fetch origin && git rebase origin/develop` (ou `git merge origin/develop`) na branch worktree imediatamente antes de iniciar T3.1. Documentar como passo 0 de T3.1.

### EC-3: Aliasing das Resources presets — godoc afirma "safe to mutate" mas Slices/Maps internos compartilham

- **Task afetada:** T1.2
- **Família:** State
- **Cenário:** `*ResourceRequirements` é `*model.ResourceRequirements`. Esse model contém `ResourceList` que é tipicamente `map[ResourceName]Quantity` ou estrutura aninhada. Mesmo retornando ponteiro fresco do struct externo, se o **literal interno** usar a mesma `ResourceList{...}` referência ou se algum campo for slice/map, mutação por consumidor A vaza para consumidor B chamando depois.
- **Impacto:** ADR-002 promete aliasing-safety mas só do struct top-level. Bug latente: consumidor faz `r := forge.ResourcesTiny(); r.Limits["cpu"] = ...` e o próximo `ResourcesTiny()` vê valor mutado.
- **Fix sugerido:** Em T1.2 task 1, garantir que cada factory aloca novos maps/listas em cada chamada (`Requests: ResourceList{model.ResourceCPU: ...}` dentro do corpo da função, NÃO referenciando um `var` package-level). Adicionar teste explícito: `TestResourcesTiny_MutationIsolation` — chama duas vezes, muta um, valida que outro não viu.

### EC-4: `make verify-coverage` pode falhar mesmo com 11 testes novos se threshold for por package OU per-file

- **Task afetada:** T1.1, T1.2, T4.1
- **Família:** Boundary
- **Cenário:** `.claude/rules/testing.md §Coverage targets` define 90% para root `forge` e 80% para sub-packages. Se `make verify-coverage` aplicar threshold per-package E o `parameter.go` tiver linhas não-testadas pré-existentes (ex.: o switch de Backoff.Factor com cases *int/int/*float64/float64), os 4 testes novos de `RetryOnFailure` (que constrói um Backoff fixo, hits 1 case) podem **piorar** ratio agregado de `parameter.go`. Idem `types.go` que hoje tem 0% direto (cobertura via integração).
- **Impacto:** `make verify` falha em coverage gate após adicionar código novo + parcialmente testado. Bloqueia T2.1.
- **Fix sugerido:** Adicionar passo a T1.1/T1.2: rodar `make verify-coverage` ANTES de T2.1 com `-func` flag para inspecionar coverage per-function. Se queda, adicionar tests para os cases não cobertos do `Backoff.Factor` switch (existe TODO declarado em parameter.go linha 117-118). Documentar como pré-condição de T2.1.

---

## SHOULD TEST

### EC-5: `forge.Ptr` para tipos não-básicos pode ter aliasing inesperado vs `ptr` original

- **Task afetada:** T3.2
- **Teste sugerido:** `TestPtrMigration_StructTypes` — em Theo, identificar todos os call sites de `ptr` com tipos não-básicos (struct, slice, map) e validar que `forge.Ptr` produz mesmo comportamento. Para basics (`int`, `string`, `bool`) é trivial-equivalente. Para `ptr(SomeStruct{...})` ambos retornam pointer fresh; OK. Mas se existe `ptr(sliceVar)` capturando uma variável local, o pointer aponta para o local da call — `forge.Ptr` faz o mesmo (Go semantics). Apenas confirmar via diff de YAML golden em 1 workflow representativo.

### EC-6: `grep "\bptr\b"` em T3.2 pode pegar falsos positivos (`Ptr` em outros contextos)

- **Task afetada:** T3.2 (task 1 e 6)
- **Teste sugerido:** Antes da substituição cega, rodar `grep -nE "\bptr\(" api/internal/argo/ | grep -v "_test.go" | tee /tmp/ptr-sites.txt` e revisar manualmente; comparar contagem com `grep -nE "\bptr\(" /tmp/ptr-sites.txt | wc -l`. Confirmar que NÃO há identificadores de variável locais nomeados `ptr` (ex.: `ptr := &foo`). O Deep Dive já menciona o risco — adicionar como gate explícito em Acceptance Criteria.

### EC-7: Skip-OOM expression hardcoded surpreende consumidores que tratam exit 137 como falha legítima

- **Task afetada:** T1.1 (ADR-001)
- **Teste sugerido:** `TestRetryOnFailure_DocsHighlightSkipOOM` — assertion no godoc text (via `go doc` parse) que CONTÉM a string "exit 137" e "skip" / "OOM". Garante que comportamento opinativo está documentado, não escondido. Complementa o teste de struct literal continuar funcionando (já planejado em T1.1 DoD).

### EC-8: Coverage para `client/` (31%), `expr/` (0%) etc. — adicionar 4 funções em `parameter.go`/`types.go` muda a média total

- **Task afetada:** T4.1 (`make verify`)
- **Teste sugerido:** Antes de Phase 2, rodar `make verify-coverage` e capturar baseline; depois de Phase 1, rodar de novo e diffar. Se total caiu abaixo de 80% (`.testcoverage.yml threshold`), T2.1 não pode prosseguir. Adicionar este check como gate explícito em DoD de T1.3.

---

## DOCUMENT

### EC-9: `genproto` conflict é mencionado em T3.1 Deep Dive mas remediação fica vaga

- **Risco aceito:** O Deep Dive diz "já resolvido no PR atual (commit `b23bf231`). Se ainda não foi merged em develop, este PR depende dele OU re-aplica a mesma técnica." Aceitável — é dependência conhecida e o plano cita o fix existente. Se Phase 3 começar e genproto reaparecer, basta cherry-pick do commit citado. Não bloqueia.

### EC-10: ADR-003 sequência rígida não cobre cenário "tag v0.5.1 puxada com bug crítico"

- **Risco aceito:** Improvável (Phase 1 tem gates fortes: race, lint, coverage, 19 E2E). Se acontecer, o remédio é tag v0.5.2 com correção. Não justifica adicionar processo no plano.

### EC-11: PR Theo aberto contra `develop` aguarda maintainer merge "fora do escopo"

- **Risco aceito:** Global DoD já marca como fora de escopo. Aceitável. Apenas garantir que o PR description deixa claro que o forge v0.5.1 é pré-requisito (link cruzado).

---

## Resumo

| Task | Edges | MUST FIX | SHOULD TEST | DOCUMENT |
|------|-------|----------|-------------|----------|
| T1.1 | 3 | 1 (EC-1) | 1 (EC-7) | 0 |
| T1.2 | 1 | 1 (EC-3) | 0 | 0 |
| T1.3 | 1 | 0 | 1 (EC-8) | 0 |
| T2.1 | 1 | 0 | 0 | 1 (EC-10) |
| T3.1 | 2 | 1 (EC-2) | 0 | 1 (EC-9) |
| T3.2 | 2 | 0 | 2 (EC-5, EC-6) | 0 |
| T3.3 | 0 | 0 | 0 | 0 |
| T3.4 | 0 | 0 | 0 | 0 |
| T3.5 | 1 | 0 | 0 | 1 (EC-11) |
| T4.1 | 1 | 1 (EC-4) | 0 | 0 |
| T4.2 | 0 | 0 | 0 | 0 |
| T4.3 | 0 | 0 | 0 | 0 |

**Veredicto: PLANO PRECISA DE AJUSTE**

Motivo principal: **EC-1** é showstopper — colisão `func RetryOnFailure` ↔ `const RetryOnFailure` quebra compilação determinística em T1.1. Os outros 3 MUST FIX (EC-2 rebase pre-Phase3, EC-3 aliasing real em maps/listas internos, EC-4 coverage threshold) são correções incrementais simples mas precisam estar no plano antes da execução para evitar retrabalho.

Após aplicar os 4 MUST FIX (renomear factory + adicionar rebase step + reforçar aliasing test + adicionar coverage baseline gate), o plano fica sólido e os 4 SHOULD TEST podem ser adicionados no curso da implementação sem alterar a arquitetura.
