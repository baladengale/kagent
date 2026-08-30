# Spec-Generation Prompt — Strong Plan Mode + Strict To-Do Ledger for the kagent k8s-agent

> **How to use this document.** Feed everything below the line as the opening instruction to a
> fresh agent session working in `/Users/bala/workspace/kagent` (with the deployment repo at
> `/Users/bala/workspace/kind` available). The generated spec goes to
> `docs/plans/plan-mode-todo-spec.md` for human review. **No implementation in that session.**
> This prompt itself is a draft for review — it has not been executed yet.

---

## 0. Role and Mission

You are a senior kagent platform engineer. Your mission is to **produce a rigorous, evidence-backed
specification** for adding "strong plan mode" to the `k8s-agent` running on the python ADK runtime —
an explicit plan-approval gate plus a strict, code-enforced To-Do ledger that the agent must follow
and keep updated throughout execution. You will NOT implement; you will verify feasibility with live
probes, write the spec, and stop.

Answer these questions from the requester explicitly, in a dedicated section, with evidence:

1. In the overall kagent framework, how can we implement strong plan mode and generate a strict
   To-Do list, follow it throughout execution, and keep it updated — how feasible is this in a
   kagent agent today?
2. Do we need to update the ADK or the python controller code? Or can we instead inject a
   monkey-patch python file **via ConfigMap** into the existing upstream release — no image rebuild,
   no fork, no (or minimal) upstream changes?

Your working assumption, to be confirmed or refuted by probes: **zero upstream code changes** —
the monkey patch ships as ConfigMap-mounted python loaded at interpreter startup via
`PYTHONPATH`/`sitecustomize`, and the pod wiring is expressed entirely through fields the deployed
Agent CRD already supports.

## 1. What "strong plan mode" means here (acceptance definition)

The target behavior, modeled on Claude Code / ZCode plan-and-track discipline:

- **Plan first.** For non-trivial requests the agent produces an explicit plan (goal, constraints,
  ordered steps) and submits it for user approval before any mutating action.
- **Approval is a hard gate.** While a plan is pending approval (or when the agent is in plan mode),
  mutating tools are denied by code — not by prompt etiquette.
- **Strict To-Do ledger.** The approved plan is materialized as todo items in durable session state.
  The agent may only make progress through the ledger: exactly one item in progress at a time,
  dependencies honored, ledger updated as work proceeds.
- **The ledger survives context compaction.** k8s-agent compacts aggressively (interval-based,
  see §2.4); a plan held only in conversation history evaporates. The todo ledger must be re-injected
  into the model's context every turn from session state.
- **Auditable progression.** Plan approval and optionally phase/checkpoint boundaries use the
  existing HITL confirmation mechanism so the transcript shows what was approved and when.
- **Toggleable per agent, zero-cost rollback.** Removing the ConfigMap/CR overrides returns the
  agent to pristine upstream behavior within one reconciliation.

## 2. Ground truth (verified 2026-08-30 — do not re-derive; ⚠ items you must re-verify)

### 2.1 Deployment topology (kind repo: `/Users/bala/workspace/kind`)

- The cluster is **kind-infra**; kagent is deployed by `scripts/80-kagent.sh` in one of two modes:
  `deploy` (mirror **upstream release images** `ghcr.io/kagent-dev/kagent` tag `0.10.0-rc3` into the
  local registry, install upstream OCI charts) and `build-deploy` (build the local checkout). The
  monkey-patch approach must work in **`deploy` mode against upstream images**.
- Customization layer = `kagent/values.yaml` (wrapper values) + `manifests/*.yaml` + numbered
  scripts (`80-kagent.sh`, `85-substrate.sh`, ...). Model routing goes through the shared
  agentgateway; the cheap tier (`agw-cheap-model-config`) is the compaction summarizer.
- **The k8s-agent Agent CR is centrally helm-managed** (verified live: `meta.helm.sh/release-name:
  kagent`, `managed-by: Helm`, subchart label `k8s-agent-0.10.0-rc3`; the Agent CRs for k8s-agent
  and kgateway-agent are in the release manifest). The kind wrapper's `kagent/values.yaml`
  `k8s-agent:` block (compaction, memory, ...) is the central customization path — it flows through
  the deployed chart's `charts/k8s-agent` subchart into the Agent CR, which the kagent controller
  then renders into the Deployment.
- **Gap to close in the spec:** the deployed 0.10.0-rc3 subchart's `agent.deploymentSpec` helper
  renders only imagePullSecrets/podSecurityContext/securityContext/nodeSelector/resources — it does
  **not** pass `env`, `volumes`, or `volumeMounts` through to the Agent CR. So wrapper values alone
  cannot add the patch ConfigMap mount + `PYTHONPATH`. The spec must pick and fully specify a
  central, drift-free wiring, in this preference order: (1) **helm post-renderer** on the
  `helm upgrade` in `scripts/80-kagent.sh` (kustomize strategic-merge patch adding the Agent CR
  volumes/volumeMounts/env; single helm ownership, survives upgrades, zero drift);
  (2) chart-native `extraObjects` passthrough (`.Values.extraObjects`, each entry `tpl`-ed) for the
  ConfigMap itself, combined with (1) or (3) for the CR fields; (3) a post-apply "ensure overrides"
  step in `80-kagent.sh` (works, but re-applies per deploy and drifts if helm runs outside the
  script); (4) vendoring the k8s-agent subchart into the kind repo with passthrough added (last
  resort — a copy to own).
- Do not treat the local checkout's `helm/` as the source of truth for the deployed chart: main
  branch removed/reworked the legacy agent subcharts (v1alpha3 migration, e.g. "remove legacy
  controller runtime #2595") while 0.10.0-rc3 still ships them. Inspect the deployed chart
  (`helm pull oci://ghcr.io/kagent-dev/kagent/helm/kagent --version 0.10.0-rc3`) when a packaging
  question depends on its templates.

### 2.2 Live cluster state

- `deployment.apps/k8s-agent` (ns `kagent`), single container `kagent`, image upstream `app`.
- Container mounts: Secret `k8s-agent` → `/config` (holds `config.json` = the serialized AgentConfig;
  the runtime materializes it via `_config_materialize.py` from `KAGENT_CONFIG_JSON` env under
  substrate, or mounts the secret directly in this deployment) and projected SA token
  `/var/run/secrets/tokens`. Runtime env currently: `ANTHROPIC_API_KEY`, `OTEL_LOGGING_ENABLED`,
  `OTEL_TRACING_ENABLED`, `KAGENT_NAMESPACE`, `KAGENT_NAME`, `KAGENT_URL`. No `PYTHONPATH` set.
- No Kyverno/Gatekeeper or admission webhooks in the cluster — do not design around policy engines.
- Served CRD `agents.kagent.dev` (v1alpha1, v1alpha2). **v1alpha2 `spec.declarative` fields:**
  `a2aConfig, context, deployment, executeCodeBlocks, memory, modelConfig, promptTemplate, runtime,
  shareTools, stream, systemMessage, systemMessageFrom, tools`.
  **`spec.declarative.deployment` supports: `env`, `envFrom`, `volumes`, `volumeMounts`,
  `extraContainers`, plus standard pod scheduling fields.** This is the load-bearing fact for the
  no-rebuild injection path (⚠ re-verify with the probe in §6.1 before writing the spec).

### 2.3 Python runtime and code anchors (kagent repo, packages/kagent-adk)

- Image (`python/Dockerfile`): `python:3.13-slim`, uv venv at `/.kagent/.venv`, runs as uid 65532,
  `WORKDIR /app`, writable `/config`, entrypoint
  `["/.kagent/.venv/bin/kagent-adk", "run", "--host", "0.0.0.0", "--port", "8080"]`
  (console script → `kagent.adk.cli:run_cli`). A plain python process → **`sitecustomize` is
  auto-imported at startup from any dir on `PYTHONPATH`** (⚠ verify no `-S`/isolated-mode surprises).
- google-adk **2.7.1** (pinned `>=2.6.2,<3` in `python/uv.lock`).
- `src/kagent/adk/types.py:403` — `AgentConfig.to_agent()`: builds the ADK agent from the deployed
  `config.json`. At `types.py:536-548` it already wires a `before_tool_callback`
  (`make_approval_callback(...)` from `_approval.py`) when tools declare `require_approval` — the
  composition precedent the patch must respect.
- `src/kagent/adk/_approval.py` — `make_approval_callback`: before_tool_callback that calls
  `tool_context.request_confirmation(...)` → ADK-native HITL pause/resume.
- `src/kagent/adk/tools/ask_user_tool.py:94` — built-in ask_user tool using
  `tool_context.request_confirmation`; the HITL plumbing lives in `_hitl.py` + the A2A HITL
  extension (pause/resume events).
- `src/kagent/adk/tools/` currently has: `ask_user_tool.py`, `bash_tool.py`, `file_tools.py`,
  `memory_tools.py`, `prefetch_memory_tool.py`, `share_tools.py`, `skill_tool.py`,
  `skills_plugin.py`, `skills_toolset.py` — **no plan/todo tools exist upstream**.
- Session state is durable: postgres-backed session service (`_session_service.py`,
  `session_db_url`). `tool_context.state` writes persist across invocations.
- Config load path: `_config_materialize.py` writes `KAGENT_CONFIG_JSON` → `/config/config.json`
  (env placeholders expanded); the CLI then loads it into `AgentConfig`. There is exactly one
  choke point where the patch can hook agent construction: **`AgentConfig.to_agent`** (⚠ trace the
  `run_cli` call path to confirm to_agent is the single construction site).

### 2.4 Compaction semantics (why the ledger must live in session state)

- k8s-agent runs `runtime: python` with context compaction; the kind wrapper values set
  `compactionInterval: 5` with summarizer `agw-cheap-model-config` (⚠ read the live value from
  Secret `k8s-agent`, key `config.json`, `context_config.compaction` — the effective interval has
  changed across deploys; historically 15).
- google-adk 2.7.1: compaction is **post-invocation**, interval counts distinct invocation_ids.
  Known bug candidate: kagent appends a synthetic `header_update` event per A2A request, so each
  real turn counts as 2 — the effective compaction cadence is ~2× faster than configured.
- Consequence: any plan/todos held only in conversation history are summarized away within a few
  turns. The design must re-inject the todo snapshot into the instruction every turn
  (ADK instruction templating `{state_key}` or a `before_agent_callback`), which makes the ledger
  compaction-proof by construction.

## 3. Hard constraints

1. **No upstream source changes and no image rebuild.** Nothing in `python/packages/`, nothing in
   `go/`, no fork, no `build-deploy` dependency. The only python we write ships via ConfigMap.
   (If a probe proves this impossible, document precisely which primitive is missing and propose
   the *smallest* upstream change as a clearly-marked fallback — do not silently widen scope.)
2. **No controller drift games.** Do not design around `kubectl patch deployment` reverts or
   admission hacks; everything must be expressible as Agent CR/ConfigMap first-class state that
   the controller renders and re-renders idempotently.
3. **Fail-open startup, fail-closed enforcement.** A bug in the patch must never prevent the agent
   from starting (log loudly, degrade to upstream behavior) — but once active, the plan gate must
   deny in code, deterministically.
4. **Respect repo conventions** (AGENTS.md / STYLE.md): semantic vs pragmatic function separation,
   data models that make wrong states unrepresentable, no speculative abstractions. Conventional
   Commits with `-s` **only if** the reviewer later asks for commits — this session writes files,
   never commits.
5. **No parallel session/task models.** Reuse upstream A2A semantics and the existing HITL
   (`request_confirmation`) plumbing. Do not invent a second todo store outside session state.
6. The k8s-agent's existing behavior (its tools, model, compaction, gateway routing) must be
   unchanged when the feature is toggled off.

## 4. Design hypothesis (turn into a spec; challenge anything that fails a probe)

### 4.1 Three layers

| Layer | What | Where it lives |
| --- | --- | --- |
| Prompt/behavior | Plan-first discipline, ledger update etiquette, when to request checkpoint approval | `spec.declarative.systemMessage` addition (Agent CR) + plan-tool docstrings |
| Tools/ledger | `submit_plan`, `write_todos`, `read_todos` (+ optional `set_checkpoint`) mutating session state | Monkey-patched-in FunctionTools |
| Enforcement | before_tool_callback gate + per-turn todo re-injection + plan approval via HITL | Monkey-patched callbacks wrapped around `AgentConfig.to_agent` output |

### 4.2 Todo ledger data model (session state; make wrong states unrepresentable)

State keys under a namespaced prefix (e.g. `plan_mode`, `plan`, `todos`) in `tool_context.state`.
Sketch — the spec owns the final shape:

```
todos: {
  version: 1,
  goal: str,
  items: [ { id, content, status: pending|in_progress|completed|cancelled,
             blocked_by: [id...], created_at, updated_at } ],
  active_item: id | null
}
plan_status: draft | awaiting_approval | approved | rejected
```

Code-enforced invariants in `write_todos` (the single mutation path):

- At most one `in_progress` item at a time.
- An item may become `in_progress` only when every `blocked_by` dependency is `completed`.
- `completed` items are immutable except an explicit, logged reopen transition.
- Transition to `completed` requires a result/summary note (auditable).
- Every write is a whole-ledger validated replace (atomic), not field patches.

State the honest limit in the spec: code enforces the **shape** and the gating, not the truth of
"the step is actually done"; checkpoint approvals via `request_confirmation` before destructive or
phase boundaries buy auditable progression.

### 4.3 Tool set (added by the patch, mirroring upstream built-in tool style)

- `submit_plan(goal, steps, constraints?)` → sets `plan_status=awaiting_approval`, calls
  `tool_context.request_confirmation(...)` (exact precedent: `ask_user_tool.py`). Rejection loops
  back to draft with the user's feedback.
- `write_todos(items)` → validated whole-ledger replace (invariants above). Also the only way to
  flip `plan_status` to `approved` on the confirmation-resume path.
- `read_todos()` → returns the ledger snapshot (for the model and for debugging).
- Optional `set_checkpoint(label)` → `request_confirmation` at phase boundaries.

### 4.4 Enforcement callbacks

- `before_tool_callback` gate: when `plan_status != approved` (or todos empty), deny every tool
  except the allowlist (`read_todos`, `read_*`-style read-only tools, `ask_user`, plan tools).
  Denials return a structured, model-readable error telling it exactly which rule it violated and
  how to proceed. ⚠ **Composition hazard:** `to_agent` may already install the approval callback
  (`types.py:537`); determine whether google-adk 2.7.1 accepts callback lists or single callables
  and compose accordingly (chain both, plan gate + approval callback).
- Per-turn re-injection: `before_agent_callback` (or instruction templating) prepends a compact
  ledger snapshot + plan status line so it survives compaction. Keep the injection small and
  stable-sized (cap item text length) to avoid polluting context.
- Config flags via env (read by the patch): e.g. `KAGENT_PLAN_MODE=on|off`,
  `KAGENT_PLAN_MUTATION_ALLOWLIST=...`, `KAGENT_PLAN_PATCH_VERSION=<n>` (bump to force pod rollout
  after ConfigMap change — python reads ConfigMap files only at process start).

### 4.5 Injection mechanism decision matrix (spec must verify and pick)

| Option | Mechanism | Status |
| --- | --- | --- |
| **A (primary)** | ConfigMap (`sitecustomize.py` + `kagent_plan_patch/` package) mounted via Agent CR `spec.declarative.deployment.{volumes, volumeMounts, env:[PYTHONPATH=/opt/kagent-plan]}`; python auto-imports `sitecustomize` at startup; the patch wraps `AgentConfig.to_agent` | CRD-native on served v1alpha2 — verify with probe §6.1/6.2 |
| B (fallback) | Tiny derived image: `FROM <upstream app image>` + `COPY` the same python + `ENV PYTHONPATH`; point the Agent CR image at it via local registry. Still zero upstream source changes, but adds an image | Only if A fails a probe |
| C (rejected — document why) | Deployment kubectl-patch (controller drift/revert), skills-init channel abuse (semantic mismatch, auth/OCI friction), admission webhook (new component), upstream fork (violates constraint 1) | Explain each rejection in one paragraph |

### 4.6 Monkey-patch architecture

- `sitecustomize.py` (ConfigMap) imports `kagent_plan_patch/` (same dir), which:
  1. Guards idempotency and wraps everything in try/except → startup must never fail open (constraint 3).
  2. Wraps `kagent.adk.types.AgentConfig.to_agent` (call original first, then: append plan tools,
     compose the gating `before_tool_callback`, add the `before_agent_callback` re-injector).
  3. Reads env flags; no-op when `KAGENT_PLAN_MODE != "on"`.
- Keep patch code free of imports that may not exist across the `>=2.6.2,<3` adk range where
  possible; import defensively and degrade explicitly.
- Log a distinctive startup marker (`kagent_plan_patch loaded, version=<n>, mode=<...>`) for
  verification.

## 5. Required spec document structure

Write `docs/plans/plan-mode-todo-spec.md` with exactly these sections:

1. **Summary & goals / non-goals** — one paragraph each; non-goals must include "no upstream
   changes", "no UI changes required (HITL round-trip must already work — verify)".
2. **Answers to the requester's questions** (§0) with probe evidence citations.
3. **Verified environment & anchors** — re-state §2 with your re-verification results (⚠ items),
   each with the command you ran and the observed output (trimmed).
4. **Functional spec** — user-visible behavior: happy path (plan → approve → execute → complete),
   rejection loop, gate denial messages, checkpoint approvals, toggle on/off.
5. **Todo ledger spec** — exact state schema, invariant table (invariant → enforcing function →
   error message), transition diagram.
6. **Tool spec** — for each tool: signature, docstring text (models read these), state effects,
   failure modes.
7. **Callback & gating spec** — decision table (plan_status × tool category → allow/deny),
   composition with the existing approval callback, re-injection format (exact template string).
8. **Prompt additions** — the exact `systemMessage` appendix text (Agent CR) and tool docstrings;
   follow the discipline: plan first for non-trivial work, one in-progress item, update before
   switching, finish-or-hand-back, no mutation while unapproved.
9. **Injection & packaging spec** — ConfigMap manifest (full YAML), the exact Agent CR override
   (volumes, volumeMounts, env incl. `KAGENT_PLAN_PATCH_VERSION`), and the wiring mechanism chosen
   from §2.1's preference order (expected: helm post-renderer in `scripts/80-kagent.sh` +
   `extraObjects` for the ConfigMap), with the full post-renderer kustomization and any script
   diff; include the apply/rollout procedure (helm upgrade → rollout wait → log-marker check) and
   proof it survives a plain re-run of `make kagent-deploy` without drift.
10. **Monkey-patch code spec** — module layout (`sitecustomize.py`, `kagent_plan_patch/__init__.py`,
    `patch.py`, `plan_tools.py`, `gating.py`, `ledger.py`, `inject.py`), function-by-function
    responsibilities (semantic vs pragmatic per STYLE.md), and the exact wrap point at
    `AgentConfig.to_agent`. Include the code-level contract for each, not necessarily full code.
11. **Observability & debugging** — startup marker, structured denial logs, how to inspect the
    ledger live (postgres: `select data from event where session_id='<id>' order by created_at`;
    session-state tables), and a `kubectl exec` one-liner to dump effective flags.
12. **Test plan** — unit tests for ledger invariants + gating decisions (pure functions, run on
    host against the ConfigMap python only); E2E script via the A2A endpoint (or `kagent` CLI)
    asserting: plan submission pauses for confirmation; mutation denied pre-approval; approval
    unblocks; ledger survives ≥1 compaction boundary; toggle-off restores upstream behavior.
    List each test with setup/steps/assertions.
13. **Rollout & rollback** — apply order, verification gates, rollback = remove CR overrides +
    delete ConfigMap (≤1 minute, no state cleanup needed beyond ledger keys).
14. **Risks & open questions** — including: does the kagent UI render
    `adk_request_confirmation`/resume properly for this agent (probe it via the ask_user tool);
    `header_update` invocation inflation accelerating compaction (ledger is immune by design —
    say so); ConfigMap size limit (1 MiB) vs patch size; python version pin drift; what happens
    on agent replacement (session state persists in postgres — confirm the ledger outlives the pod).

## 6. Verification probes (run before/while writing the spec; cite results in §3 of the spec)

1. **CRD round-trip (blocks Option A):** create a scratch ConfigMap + add
   `spec.declarative.deployment.volumes/volumeMounts/env` to k8s-agent (or a scratch agent),
   wait for rollout, and confirm `kubectl get deploy k8s-agent -o jsonpath=...` shows the mount +
   env rendered by the controller — i.e. the fields survive translation and are not stripped.
   Revert afterwards.
2. **sitecustomize hook:** with the probe mount, ship a `sitecustomize.py` that just
   `print("KAGENT_PLAN_PATCH_PROBE ...", file=sys.stderr)`; confirm the marker in pod logs
   (`kubectl logs`), proving interpreter-startup loading works with the
   `/.kagent/.venv/bin/kagent-adk` console-script entrypoint (no `-S`, uid 65532 can read
   ConfigMap mounts — defaultMode 0444).
3. **Hook-point trace:** in-pod (`kubectl exec ... python -c ...`) confirm
   `import kagent.adk.types` works, inspect `AgentConfig.to_agent` signature, and check whether
   google-adk 2.7.1 `LlmAgent` callbacks accept lists or single callables (this decides the
   composition strategy in §4.4).
4. **HITL round-trip:** on the current (unpatched) k8s-agent, invoke ask_user through the same
   channel you'll use for E2E and confirm the pause/resume event flow works end-to-end — this
   validates `submit_plan`'s approval mechanism and tells you whether the UI needs anything.
5. **Compaction reality check:** read the live compaction config from Secret `k8s-agent`
   (`config.json`) and note the effective interval; compute how many turns the plan has to survive
   (accounting for the `header_update` double-count).

## 7. Questions the spec must answer explicitly

- Which exact python entry state do we hook (`sitecustomize` vs wrapping `run_cli`), and why?
- Is `to_agent` the single agent-construction choke point on the `run` path? (cite the call chain)
- How do the two `before_tool_callback`s compose in adk 2.7.1 — list vs manual chain?
- What is the minimal Agent CR diff, and does anything in the Go translator normalize/strip
  ConfigMap volumes or `PYTHONPATH` env (check `adk_api_translator.go` / `deployments.go` handling
  of `SharedDeploymentSpec`)?
- How exactly do the ConfigMap + CR overrides get in centrally (post-renderer vs extraObjects vs
  post-apply step), and does the post-rendered output stay stable across repeated
  `make kagent-deploy` runs (idempotency proof)?
- What breaks when ConfigMap content changes but the pod doesn't restart — and why does
  `KAGENT_PLAN_PATCH_VERSION` fix it?
- If any probe invalidates Option A: what is the smallest upstream change that would restore a
  no-rebuild path (e.g. translator honoring a new field), stated as fallback only?

## 8. Deliverables

1. `docs/plans/plan-mode-todo-spec.md` — the spec (§5 structure), written for review, evidence-cited.
2. Any probe manifests/scripts you used, left under `docs/plans/plan-mode-probes/` (clearly named,
   reversible; delete cluster scratch objects and say so).
3. A closing summary message: verdict on "no upstream changes" feasibility, the chosen injection
   option with probe evidence, and the 3 riskiest assumptions for the reviewer to check first.

**Do not** implement the feature, modify `python/packages/`, `go/`, the Agent CR beyond probes
(reverted), or commit anything.

## 9. Review criteria (how this spec will be judged)

- Every load-bearing claim is backed by a probe output or a `file:line` citation.
- The "no upstream changes" answer is definitive, with the fallback precisely scoped if needed.
- The invariant table + gating decision table are complete enough to implement without further
  design decisions.
- Rollback is genuinely ≤1 minute and returns to pristine upstream behavior.
- The spec is implementable by a fresh agent session without re-deriving any of this exploration.
