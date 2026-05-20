# Flow Diagrams — theo-forge

Generated: Phase 1, Iteration 1 (2026-05-20)

## Flow 1: Build Workflow to YAML (Primary User Flow)

**Criticality**: CRITICAL — this is the entire raison d'etre of the SDK.

```
Consumer code
  |
  | new(forge.Workflow{...Templates: []Templatable{...}})
  v
Workflow.ToYAML()
  |
  v
Workflow.Build()
  |
  +-- validate()                          <- name/generateName checks
  |
  +-- buildTemplateModels(w.Templates)
  |     for each Templatable:
  |       t.BuildTemplate()               <- dispatched to Container/DAG/Steps/etc.
  |       config.DispatchTemplateHooks(&m) <- apply registered template hooks
  |
  +-- w.buildArguments()                  <- Parameters -> ParameterModel
  |
  +-- w.buildVolumes()                    <- VolumeBuilder -> VolumeModel
  |
  +-- w.buildVolumeClaimTemplates()       <- PVCVolume -> PVCModel
  |
  +-- buildRetryStrategyModel()
  |
  +-- resolveAffinity(w)                  <- DefaultPodAffinity injection
  |     if w.Affinity != nil   -> user wins
  |     if DisableDefaultAffinity -> nil
  |     if len(VolumeClaimTemplates)==0 -> nil
  |     else -> DefaultPodAffinityFor(w) -> ColocateByLabel(...)
  |
  +-- assemble model.WorkflowModel{...}
  |
  +-- config.DispatchWorkflowHooks(&wfModel)  <- apply registered workflow hooks
  |
  v
serialize.WorkflowToYAML(model.WorkflowModel)
  |
  v
sigs.k8s.io/yaml.Marshal(m)
  |
  v
YAML string (Argo Workflow CRD)
```

**Error paths:**
- `validate()` returns error if name > 63 chars or both name and generateName empty
- `BuildTemplate()` on any Templatable propagates errors (empty name, empty manifest, etc.)
- `buildArguments()` errors if any Parameter has empty Name
- `yaml.Marshal()` can error on cyclic or unencodable values (practically impossible with these types)

---

## Flow 2: Client Submit Lifecycle

**Criticality**: HIGH — connects SDK output to a running Argo server.

```
Consumer code
  |
  | client.NewWorkflowsService(host, token, namespace)
  v
WorkflowsService{Host, Token, Namespace, VerifySSL=true, HTTPClient=&http.Client{30s}}

  +-- CreateWorkflow(ctx, Buildable b)
  |     b.Build() -> model.WorkflowModel
  |     b.GetNamespace() -> ns
  |     CreateWorkflowFromModel(ctx, wfModel, ns)
  |
  +-- doRequest(ctx, POST, "/api/v1/workflows/"+ns, body)
  |     json.Marshal(body) -> reqBody
  |     http.NewRequestWithContext(ctx, POST, url, reqBody)
  |     req.Header.Set("Content-Type", "application/json")
  |     req.Header.Set("Authorization", "Bearer "+token)
  |     s.HTTPClient.Do(req) -> resp
  |     io.ReadAll(resp.Body)
  |     if resp.StatusCode >= 400 -> &APIError{StatusCode, Message}
  |     else -> respBody
  |
  +-- json.Unmarshal(respBody, &model.WorkflowModel) -> result
  v
model.WorkflowModel (response from Argo API)
```

**Lifecycle operations (all use PUT):**
- `StopWorkflow` -> PUT `/api/v1/workflows/{ns}/{name}/stop`
- `TerminateWorkflow` -> PUT `/api/v1/workflows/{ns}/{name}/terminate`
- `SuspendWorkflow` -> PUT `/api/v1/workflows/{ns}/{name}/suspend`
- `ResumeWorkflow` -> PUT `/api/v1/workflows/{ns}/{name}/resume`

**NOTE**: `VerifySSL` field exists on WorkflowsService struct but is NEVER applied to the http.Client — it is a dead field.

---

## Flow 3: WorkflowTemplate / CronWorkflow Round-trip

**Criticality**: MEDIUM — covers reusable template and scheduled workflow patterns.

```
BUILD path:
  WorkflowTemplate{Name, Namespace, Entrypoint, Templates}
    |
    v
  WorkflowTemplate.ToYAML()
    -> Build()
      -> buildTemplateModels() [same as Flow 1]
      -> serialize.WorkflowTemplateToYAML(model.WorkflowTemplateModel)
    |
    v
  YAML with Kind: WorkflowTemplate

PARSE path:
  serialize.WorkflowTemplateFromYAML(yamlStr)
    -> yaml.Unmarshal([]byte(yamlStr), &model.WorkflowTemplateModel)
    -> model.WorkflowTemplateModel

CronWorkflow follows identical pattern:
  Kind: CronWorkflow
  serialize.CronWorkflowToYAML / CronWorkflowFromYAML
  Additional fields: Schedule, Timezone, ConcurrencyPolicy, etc.
```

---

## Flow 4: DAG Dependency Resolution

**Criticality**: HIGH — DAGs are the primary workflow structure for complex pipelines.

```
Consumer code:
  dag := &forge.DAG{Name: "pipeline"}
  A := &forge.Task{Name: "A", Template: "echo"}
  B := &forge.Task{Name: "B", Template: "echo"}
  C := &forge.Task{Name: "C", Template: "echo"}

  A.Then(B)   // B.Depends = "A"
  A.Then(C)   // C.Depends = "A"
  B.Then(C)   // C.Depends = "A && B"  <- string concatenation, no parens

  dag.AddTask(A)  // nodeNames["A"] = true
  dag.AddTask(B)  // nodeNames["B"] = true
  dag.AddTask(C)  // nodeNames["C"] = true

  dag.BuildTemplate()
    for each task: t.BuildDAGTask() -> model.DAGTaskModel{Depends: "A && B"}
    -> model.TemplateModel{DAG: &model.DAGModel{Tasks: [...]}}

Variant: Task.OnSuccess(other), OnFailure(other), OnError(other)
  sets Depends = "other-name.Succeeded" / ".Failed" / ".Errored"
  NOTE: These REPLACE Depends (not append) — only one condition per call
```

**WARNING identified**: `Task.Then()` appends to `Depends` with ` && ` but does NOT add parentheses. For complex expressions mixing `Then()` and `Or()`, operator precedence may produce unexpected results. Example: `A.Then(B); A.Or(C).Then(D)` will not group correctly.

---

## Flow 5: Global Config Hook Dispatch

**Criticality**: MEDIUM — cross-cutting concern that applies to all Build() calls.

```
Setup (one-time, typically in main/init):
  cfg := forge.GetGlobalConfig()   // returns config.globalConfig singleton
  cfg.RegisterTemplateHook(func(tpl *model.TemplateModel) {
      // mutate tpl — add default resources, inject labels, etc.
  })
  cfg.RegisterWorkflowHook(func(wf *model.WorkflowModel) {
      // mutate wf — add org labels, annotations, etc.
  })

Per Build() call:
  Workflow.Build()
    |
    +-- buildTemplateModels(templates)
    |     for each tpl:
    |       t.BuildTemplate() -> model.TemplateModel
    |       cfg.DispatchTemplateHooks(&tpl)  <- hooks mutate tpl in-place
    |
    +-- assemble wfModel
    +-- cfg.DispatchWorkflowHooks(&wfModel)  <- hooks mutate wfModel in-place
    v
  modified model.WorkflowModel

Isolation (for tests / concurrent builds):
  cfg := forge.NewConfig()   // independent instance, same defaults
  // changes to cfg do NOT affect global singleton
  // NOTE: the root package's buildTemplateModels always uses the *global* singleton
  //       via the package-level `globalConfig` var — NewConfig() is NOT wired in
```

**WARNING identified**: `NewConfig()` creates an isolated GlobalConfig, but the root package's `buildTemplateModels()` in helpers.go uses the package-level `globalConfig` variable (which is `config.GetGlobal()`). Therefore, hooks registered on an isolated config created via `forge.NewConfig()` are NOT applied during `Build()`. This is a documentation/API contract issue.
