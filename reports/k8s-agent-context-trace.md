# k8s-agent context-management trace — live run, 2026-08-30

Session: `df1fae8a-bbc5-42a4-b187-5a87ecc7d948` (Postgres `event` table, kagent DB)
Runtime: **python** (`kagent-dev/kagent/app:0.10.0-rc3`) — the Go runtime silently ignores compaction.
Main model: `anthropic/ark-code-latest` via agentgateway (`agw-model-config`).
Summarizer: `anthropic/seed-2-0-mini-260428` via agentgateway (`agw-cheap-model-config`).
Installed google-adk: **2.7.1**.

## 0. What "context" is here

One user invocation = a run of ADK Events:
`user msg → model event (text + function_call) → tool event (function_response) → … → model event (final text)`.
Every LLM call re-sends: system instruction + 20 tool schemas + ALL prior event contents.
Tool calls do not call the LLM; each model turn = exactly one LLM call.

## 1. Full trace of the 20-step health check (22 tool calls, 23 LLM calls)

=========================================================================================================
FIXED PREFIX EVERY CALL: system ≈ 1 tok + 20 tool schemas ≈ 2600 tok  →  ≈ 2601 tok
=========================================================================================================

USER INVOCATION: 1358 ch (~340 tok)

LLM CALL #1 (event 2)
  reported tokens: prompt=417  cached=None  output=59  thoughts=None
  est. ACTUAL prompt ≈ 3250 tok  = 1 sys + 2600 tools + 649 conv
  response: 119 ch text + 1 tool call(s): k8s_get_cluster_configuration
    text: I'll work through the health check steps strictly in order. Starting with step 1.  ## Step 1: Get Cl
    -> k8s_get_cluster_configuration({})
  ⮑ TOOL RESULT k8s_get_cluster_configuration: 139 ch (~35 tok) isError=False
      {
    "kind": "Config",
    "apiVersion": "v1",
    "clusters": null,
    "users": null,
    "contex

LLM CALL #2 (event 4)
  reported tokens: prompt=149  cached=None  output=43  thoughts=None
  est. ACTUAL prompt ≈ 3628 tok  = 1 sys + 2600 tools + 1027 conv
  response: 38 ch text + 1 tool call(s): k8s_get_available_api_resources
    text: ## Step 2: Get Available API Resources
    -> k8s_get_available_api_resources({})
  ⮑ TOOL RESULT k8s_get_available_api_resources: 9815 ch (~2454 tok) isError=False
      NAME                                SHORTNAMES   APIVERSION                        NAMESPACED   KIND

LLM CALL #3 (event 6)
  reported tokens: prompt=1714  cached=None  output=95  thoughts=None
  est. ACTUAL prompt ≈ 6458 tok  = 1 sys + 2600 tools + 3857 conv
  response: 48 ch text + 1 tool call(s): k8s_get_resources
    text: ## Step 3: Get Resources for pods in kube-system
    -> k8s_get_resources({"resource_type": "pod", "namespace": "kube-system", "output": "wide"})
  ⮑ TOOL RESULT k8s_get_resources: 1350 ch (~338 tok) isError=False
      NAME                                           READY   STATUS    RESTARTS       AGE     IP          

LLM CALL #4 (event 8)
  reported tokens: prompt=448  cached=None  output=95  thoughts=None
  est. ACTUAL prompt ≈ 7151 tok  = 1 sys + 2600 tools + 4550 conv
  response: 60 ch text + 1 tool call(s): k8s_get_resources
    text: ## Step 4: Get Resources for deployments in kagent namespace
    -> k8s_get_resources({"resource_type": "deployment", "namespace": "kagent", "output": "wide"})
  ⮑ TOOL RESULT k8s_get_resources: 1968 ch (~492 tok) isError=False
      NAME                             READY   UP-TO-DATE   AVAILABLE   AGE     CONTAINERS   IMAGES       

LLM CALL #5 (event 10)
  reported tokens: prompt=573  cached=None  output=94  thoughts=None
  est. ACTUAL prompt ≈ 7998 tok  = 1 sys + 2600 tools + 5397 conv
  response: 57 ch text + 1 tool call(s): k8s_get_resources
    text: ## Step 5: Get Resources for services in kagent namespace
    -> k8s_get_resources({"resource_type": "service", "namespace": "kagent", "output": "wide"})
  ⮑ TOOL RESULT k8s_get_resources: 1770 ch (~442 tok) isError=False
      NAME                                             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)  

LLM CALL #6 (event 12)
  reported tokens: prompt=549  cached=None  output=59  thoughts=None
  est. ACTUAL prompt ≈ 8779 tok  = 1 sys + 2600 tools + 6178 conv
  response: 42 ch text + 1 tool call(s): k8s_get_events
    text: ## Step 6: Get Events for namespace kagent
    -> k8s_get_events({"namespace": "kagent"})
  ⮑ TOOL RESULT k8s_get_events: 14047 ch (~3512 tok) isError=False
      {
    "apiVersion": "v1",
    "items": [
        {
            "action": "Binding",
            "api

LLM CALL #7 (event 14)
  reported tokens: prompt=4076  cached=None  output=76  thoughts=None
  est. ACTUAL prompt ≈ 13000 tok  = 1 sys + 2600 tools + 10399 conv
  response: 48 ch text + 1 tool call(s): k8s_get_resources
    text: ## Step 7: Get Resources for nodes (wide output)
    -> k8s_get_resources({"resource_type": "node", "output": "wide"})
  ⮑ TOOL RESULT k8s_get_resources: 327 ch (~82 tok) isError=False
      NAME                   STATUS   ROLES           AGE     VERSION   INTERNAL-IP   EXTERNAL-IP   OS-IMA

LLM CALL #8 (event 16)
  reported tokens: prompt=173  cached=None  output=102  thoughts=None
  est. ACTUAL prompt ≈ 13443 tok  = 1 sys + 2600 tools + 10842 conv
  response: 62 ch text + 1 tool call(s): k8s_describe_resource
    text: ## Step 8: Describe the kagent-controller deployment in kagent
    -> k8s_describe_resource({"resource_type": "deployment", "resource_name": "kagent-controller", "namespace": "kagent"})
  ⮑ TOOL RESULT k8s_describe_resource: 3049 ch (~762 tok) isError=False
      Name:                   kagent-controller
Namespace:              kagent
CreationTimestamp:      Sun

LLM CALL #9 (event 18)
  reported tokens: prompt=906  cached=None  output=135  thoughts=None
  est. ACTUAL prompt ≈ 14598 tok  = 1 sys + 2600 tools + 11997 conv
  response: 121 ch text + 1 tool call(s): k8s_get_resources
    text: ## Step 9: GetPodLogs for the kagent-controller pod (last 40 lines)  Let me first get the pod name f
    -> k8s_get_resources({"resource_type": "pod", "namespace": "kagent", "output": "wide", "resource_name": "kagent-controller"})
  ⮑ TOOL RESULT k8s_get_resources: 78 ch (~20 tok) isError=True
      [Kubernetes] get pod kagent-controller -n kagent -o wide failed: exit status 1

LLM CALL #10 (event 20)
  reported tokens: prompt=52  cached=None  output=92  thoughts=None
  est. ACTUAL prompt ≈ 14966 tok  = 1 sys + 2600 tools + 12365 conv
  response: 47 ch text + 1 tool call(s): k8s_get_resources
    text: Let me get the pod list to find the exact name.
    -> k8s_get_resources({"resource_type": "pod", "namespace": "kagent", "output": "wide"})
  ⮑ TOOL RESULT k8s_get_resources: 1377 ch (~344 tok) isError=False
      NAME                                              READY   STATUS    RESTARTS       AGE     IP       

LLM CALL #11 (event 22)
  reported tokens: prompt=479  cached=None  output=118  thoughts=None
  est. ACTUAL prompt ≈ 15672 tok  = 1 sys + 2600 tools + 13071 conv
  response: 65 ch text + 1 tool call(s): k8s_get_pod_logs
    text: ## Step 9: Get logs for the kagent-controller pod (last 40 lines)
    -> k8s_get_pod_logs({"pod_name": "kagent-controller-64986db4b4-n9q6j", "namespace": "kagent", "tail_lines": 40})
  ⮑ TOOL RESULT k8s_get_pod_logs: 11743 ch (~2936 tok) isError=False
      {"level":"info","ts":"2026-08-30T04:33:46Z","logger":"http.sessions-handler","msg":"Successfully add

LLM CALL #12 (event 24)
  reported tokens: prompt=4845  cached=None  output=115  thoughts=None
  est. ACTUAL prompt ≈ 19343 tok  = 1 sys + 2600 tools + 16742 conv
  response: 58 ch text + 1 tool call(s): k8s_get_pod_logs
    text: ## Step 10: Get logs for the kagent-ui pod (last 20 lines)
    -> k8s_get_pod_logs({"pod_name": "kagent-ui-6df775c67b-7bnqm", "namespace": "kagent", "tail_lines": 20})
  ⮑ TOOL RESULT k8s_get_pod_logs: 1078 ch (~270 tok) isError=False
      2026-08-30 04:32:40,010 DEBG 'nginx' stdout output:
10.244.0.1 - - [30/Aug/2026:04:32:40 +0000] "GET

LLM CALL #13 (event 26)
  reported tokens: prompt=656  cached=None  output=115  thoughts=None
  est. ACTUAL prompt ≈ 19987 tok  = 1 sys + 2600 tools + 17386 conv
  response: 61 ch text + 1 tool call(s): k8s_get_pod_logs
    text: ## Step 11: Get logs for the kagent-tools pod (last 20 lines)
    -> k8s_get_pod_logs({"pod_name": "kagent-tools-666cb9dbcf-lpmfn", "namespace": "kagent", "tail_lines": 20})
  ⮑ TOOL RESULT k8s_get_pod_logs: 4315 ch (~1079 tok) isError=False
      time=2026-08-30T04:33:43.083Z level=INFO msg="command execution successful" trace_id=2ad0734e45e9a6f

LLM CALL #14 (event 28)
  reported tokens: prompt=1737  cached=None  output=103  thoughts=None
  est. ACTUAL prompt ≈ 21451 tok  = 1 sys + 2600 tools + 18850 conv
  response: 65 ch text + 1 tool call(s): k8s_get_resource_yaml
    text: ## Step 12: Get resource YAML for the kagent-ui service in kagent
    -> k8s_get_resource_yaml({"resource_type": "service", "resource_name": "kagent-ui", "namespace": "kagent"})
  ⮑ TOOL RESULT k8s_get_resource_yaml: 980 ch (~245 tok) isError=False
      apiVersion: v1
kind: Service
metadata:
  annotations:
    meta.helm.sh/release-name: kagent
    meta

LLM CALL #15 (event 30)
  reported tokens: prompt=413  cached=None  output=96  thoughts=None
  est. ACTUAL prompt ≈ 22071 tok  = 1 sys + 2600 tools + 19470 conv
  response: 74 ch text + 1 tool call(s): k8s_check_service_connectivity
    text: ## Step 13: Check service connectivity for the kagent-ui service in kagent
    -> k8s_check_service_connectivity({"service_name": "kagent-ui.kagent.svc.cluster.local:8080", "namespace": "kagent"})
  ⮑ TOOL RESULT k8s_check_service_connectivity: 60650 ch (~15162 tok) isError=False
      <!DOCTYPE html><html lang="en" class=""><head><meta charSet="utf-8"/><meta name="viewport" content="

LLM CALL #16 (event 32)
  reported tokens: prompt=19841  cached=None  output=114  thoughts=None
  est. ACTUAL prompt ≈ 39104 tok  = 1 sys + 2600 tools + 36503 conv
  response: 181 ch text + 1 tool call(s): k8s_get_resources
    text: The service connectivity check returned the full HTML page from the kagent-ui service, confirming it
    -> k8s_get_resources({"resource_type": "pod", "namespace": "default", "output": "wide"})
  ⮑ TOOL RESULT k8s_get_resources: 41 ch (~10 tok) isError=False
      No resources found in default namespace.


LLM CALL #17 (event 34)
  reported tokens: prompt=134  cached=None  output=95  thoughts=None
  est. ACTUAL prompt ≈ 39465 tok  = 1 sys + 2600 tools + 36864 conv
  response: 50 ch text + 1 tool call(s): k8s_get_resources
    text: ## Step 15: Get Resources for configmaps in kagent
    -> k8s_get_resources({"resource_type": "configmap", "namespace": "kagent", "output": "wide"})
  ⮑ TOOL RESULT k8s_get_resources: 226 ch (~56 tok) isError=False
      NAME                     DATA   AGE
kagent-builtin-prompts   5      6d13h
kagent-controller        2

LLM CALL #18 (event 36)
  reported tokens: prompt=185  cached=None  output=113  thoughts=None
  est. ACTUAL prompt ≈ 39885 tok  = 1 sys + 2600 tools + 37284 conv
  response: 56 ch text + 1 tool call(s): k8s_describe_resource
    text: ## Step 16: Describe the kagent-postgresql pod in kagent
    -> k8s_describe_resource({"resource_type": "pod", "resource_name": "kagent-postgresql-7665497754-2sld9", "namespace": "kagent"})
  ⮑ TOOL RESULT k8s_describe_resource: 3017 ch (~754 tok) isError=False
      Name:             kagent-postgresql-7665497754-2sld9
Namespace:        kagent
Priority:         0
Se

LLM CALL #19 (event 38)
  reported tokens: prompt=958  cached=None  output=118  thoughts=None
  est. ACTUAL prompt ≈ 41017 tok  = 1 sys + 2600 tools + 38416 conv
  response: 66 ch text + 1 tool call(s): k8s_get_pod_logs
    text: ## Step 17: Get logs for the kagent-postgresql pod (last 15 lines)
    -> k8s_get_pod_logs({"pod_name": "kagent-postgresql-7665497754-2sld9", "namespace": "kagent", "tail_lines": 15})
  ⮑ TOOL RESULT k8s_get_pod_logs: 2967 ch (~742 tok) isError=False
      2026-08-30 03:57:01.836 UTC [21] LOG:  checkpoint complete: wrote 38 buffers (0.2%), wrote 1 SLRU bu

LLM CALL #20 (event 40)
  reported tokens: prompt=1286  cached=None  output=94  thoughts=None
  est. ACTUAL prompt ≈ 42113 tok  = 1 sys + 2600 tools + 39512 conv
  response: 49 ch text + 1 tool call(s): k8s_get_resources
    text: ## Step 18: Get Resources for endpoints in kagent
    -> k8s_get_resources({"resource_type": "endpoints", "namespace": "kagent", "output": "wide"})
  ⮑ TOOL RESULT k8s_get_resources: 823 ch (~206 tok) isError=False
      Warning: v1 Endpoints is deprecated in v1.33+; use discovery.k8s.io/v1 EndpointSlice
NAME           

LLM CALL #21 (event 42)
  reported tokens: prompt=347  cached=None  output=106  thoughts=None
  est. ACTUAL prompt ≈ 42686 tok  = 1 sys + 2600 tools + 40085 conv
  response: 76 ch text + 1 tool call(s): k8s_get_resource_yaml
    text: ## Step 19: Get resource YAML for the kagent-controller deployment in kagent
    -> k8s_get_resource_yaml({"resource_type": "deployment", "resource_name": "kagent-controller", "namespace": "kagent"})
  ⮑ TOOL RESULT k8s_get_resource_yaml: 4409 ch (~1102 tok) isError=False
      apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    deployment.kubernetes.io/revision:

LLM CALL #22 (event 44)
  reported tokens: prompt=1349  cached=None  output=96  thoughts=None
  est. ACTUAL prompt ≈ 44183 tok  = 1 sys + 2600 tools + 41582 conv
  response: 51 ch text + 1 tool call(s): k8s_get_resources
    text: ## Step 20: Get Resources for replicasets in kagent
    -> k8s_get_resources({"resource_type": "replicaset", "namespace": "kagent", "output": "wide"})
  ⮑ TOOL RESULT k8s_get_resources: 4912 ch (~1228 tok) isError=False
      NAME                                        DESIRED   CURRENT   READY   AGE     CONTAINERS   IMAGES 

LLM CALL #23 (event 46)
  reported tokens: prompt=1553  cached=None  output=250  thoughts=None
  est. ACTUAL prompt ≈ 45912 tok  = 1 sys + 2600 tools + 43311 conv
  response: 1134 ch text + 0 tool call(s): text only
    text: All 20 steps are complete. Here is the final health summary.  ---  ## Cluster Health Summary  The cl

=========================================================================================================
CONTEXT GROWTH CURVE (est. actual prompt tokens per LLM call):
  #1:3250 → #2:3628 → #3:6458 → #4:7151 → #5:7998 → #6:8779 → #7:13000 → #8:13443 → #9:14598 → #10:14966 → #11:15672 → #12:19343 → #13:19987 → #14:21451 → #15:22071 → #16:39104 → #17:39465 → #18:39885 → #19:41017 → #20:42113 → #21:42686 → #22:44183 → #23:45912

TOOL OUTPUTS (by size):
  k8s_check_service_connectivity           60650 ch   err=False
  k8s_get_events                           14047 ch   err=False
  k8s_get_pod_logs                         11743 ch   err=False
  k8s_get_available_api_resources           9815 ch   err=False
  k8s_get_resources                         4912 ch   err=False
  k8s_get_resource_yaml                     4409 ch   err=False
  k8s_get_pod_logs                          4315 ch   err=False
  k8s_describe_resource                     3049 ch   err=False
  k8s_describe_resource                     3017 ch   err=False
  k8s_get_pod_logs                          2967 ch   err=False
  k8s_get_resources                         1968 ch   err=False
  k8s_get_resources                         1770 ch   err=False
  k8s_get_resources                         1377 ch   err=False
  k8s_get_resources                         1350 ch   err=False
  k8s_get_pod_logs                          1078 ch   err=False
  k8s_get_resource_yaml                      980 ch   err=False
  k8s_get_resources                          823 ch   err=False
  k8s_get_resources                          327 ch   err=False
  k8s_get_resources                          226 ch   err=False
  k8s_get_cluster_configuration              139 ch   err=False
  k8s_get_resources                           78 ch   err=True
  k8s_get_resources                           41 ch   err=False

RUN TOTALS:
  LLM calls: 23   tool calls: 22
  reported prompt sum: 42840
  reported output sum: 2383
  total tool-output chars: 129081 (~32270 tok)
  cache tokens reported: 0 (field null on every call)


## 2. The compaction event (captured live with compactionInterval=1)

Appended **after** invocation N completes (post-invocation, `runners._run_post_invocation_compaction`).

```json
{
  "author": "user",
  "invocation_id": "1229d2e7-a42d-4930-b2b3-b7a61d959775",
  "actions.compaction.start_timestamp": 1788064418.7389755,
  "actions.compaction.end_timestamp": 1788064896.842009,
  "compacted_content.role": "model",
  "summary_chars": 1283,
  "usage_metadata (this is the SUMMARIZER call)": {
    "cache_tokens_details": null,
    "cached_content_token_count": null,
    "candidates_token_count": 264,
    "candidates_tokens_details": null,
    "prompt_token_count": 1600,
    "prompt_tokens_details": null,
    "thoughts_token_count": null,
    "tool_use_prompt_token_count": null,
    "tool_use_prompt_tokens_details": null,
    "total_token_count": 1864,
    "traffic_type": null
  }
}
```

**Summary written by the cheap model (1,283 chars, replacing ~46k tokens of raw history):**

> Here is a concise summary of the conversation:

## Summary
The user requested a comprehensive 20-step health check of a Kubernetes cluster, executed strictly in sequential order with exactly one tool call per step (no batching). The agent followed the steps precisely, gathering cluster configuration, API resources, pod/deployment/service information, logs, YAML specs, and connectivity checks.

**Key findings:** The cluster is healthy overall. The control plane and all kube-system components are Ready. All 8 deployments in the `kagent` namespace are at 1/1 readiness, with services and endpoints properly mapped. The kagent-ui service is reachable and serving its web interface. PostgreSQL runs normally. Several pods had restarts but are currently stable.

**Follow-up:** The user then asked for the top 3 risks based only on the data already gathered (no new tool calls). The agent identified: (1) container crashes/restarts across multiple pods (controller, UI, postgresql), (2) kgateway-agent running with `UnsupportedFeatures=True` (context compaction not implemented), and (3) high churn from repeated rollouts with 10+ retained ReplicaSets and a temporary readiness failure.

**No unresol

Anatomy of what the NEXT LLM call sees (built by `flows/llm_flows/_content_compaction.py`):
1. System instruction + 20 tool schemas (unchanged, still prefix-cacheable)
2. The summary text, inserted where the compaction event sits
3. All raw events with `timestamp >= start_ts` that are NOT inside a compacted range (verbatim)
4. The new user message

Each further compaction: range starts `overlap_size` (2) invocations before the previous
compaction's end → consecutive summaries share a seam; both summaries stay in the request.

## 3. Verified google-adk 2.7.1 semantics (from source)

- Trigger is **post-invocation only** — nothing compacts mid-invocation, however large it grows.
- `compactionInterval` counts **user invocations**, not tool calls / LLM calls.
- **`eventRetentionSize` is honored ONLY on the token-threshold trigger** (keeps last N raw
  events verbatim). Interval-triggered compaction summarizes the full range — observed: the
  summary covered both invocations end-to-end.
- Validator requires `tokenThreshold`⇔`eventRetentionSize` and `compactionInterval`⇔`overlapSize`
  to be set together; at least one pair must be present.
- Summarizer input: each tool arg/response **truncated to 2,000 chars**; temp 0; fixed template
  ("reiterate the user request, summarize decisions, information, unresolved tasks").

## 4. Usage-reporting caveat on this provider chain (fine-tuning critical)

`usage_metadata.prompt_token_count` from agentgateway→ARK is **NOT cumulative** — it tracks the
newest block(s) (e.g. 149 for a ~3.6k-token request; 19,841 right after the 60k-char HTML tool
result), and `cached_content_token_count` is null on every call. Most consistent explanation:
the upstream reports only non-cached input tokens and the gateway drops cache details.

Consequences:
- **`tokenThreshold` compares against a non-cumulative number on this chain** — 300k will
  effectively never fire; the threshold would need a provider reporting true cumulative prompts.
- The prefix cache evidently IS working (low reported inputs), but its size is invisible.
- Compaction timing on this chain should be driven by `compactionInterval`.

## 5. Tuning verdict for this agent

| Question | Verdict |
| --- | --- |
| Interval 15 vs 20? | **15.** Heavy invocation here = ~46k tok/22 tool calls. 20 heavy invocations ≈ 970k — too close to 1M. 15 keeps worst case ≈ 700k with margin. |
| Compaction vs cache? | Not competitors. Cache makes repetition cheap *within* the current window; compaction bounds the window itself. With caching active, compaction "costs" only the boundary re-cache + summary call; it wins on overflow safety and long-run cost once context ≫ ~100k. |
| `eventRetentionSize: 6` | Only matters if the threshold path ever fires; harmless otherwise. |
| Biggest lever found | `k8s_check_service_connectivity` returned a 60,650-char HTML page (~15k tok) — one call doubled the context. Cap/truncate tool outputs at the tool-server; that beats any compaction tuning. |
| Reliability note | This run also showed a tool error path (`get pod kagent-controller` failed → retry with pod list) — errors are just events in context; fine. |

## 6. How to reproduce / observe

```bash
# events (full ADK JSON incl. usage + compaction):
kubectl exec -n kagent deploy/kagent-postgresql -- psql -U kagent -d kagent \
  -t -A -c "select data from event where session_id='<id>' order by created_at"

# trigger compaction quickly for demos: set compactionInterval: 1, send a follow-up
# with contextId=<session>, then look for the event with actions.compaction.
```


---

# PART 2 — Parameter-by-parameter demo (all knobs shrunk to fire fast)

Demo config: `compactionInterval: 2, overlapSize: 1, eventRetentionSize: 3, tokenThreshold: 1500` (restored to 15/2/6/300000 after).

## Demo 1 — `compactionInterval` + `overlapSize` (session e05cc906, 4 tiny invocations)

Two compaction events captured. Both fired **after every single invocation**, not every 2nd:

| | created | range (start .. end) | summarizer usage | what it replaced |
|---|---|---|---|---|
| Compaction #1 | 05:28:09.9 | 1788067681.8 .. 1788067686.0 | 286 in / 241 out (seed-2-0-mini) | invocation 1 (~1.1k tok) |
| Compaction #2 | 05:28:15.2 | 1788067681.8 .. 1788067692.8 | 514 in / 284 out | invocations 1–2, re-summarizing #1's range |

### FINDING 1 (bug): kagent's `header_update` events inflate ADK's invocation count

Every A2A request makes kagent append a synthetic event with `invocation_id: "header_update"`
(it carries HTTP headers into session state via `state_delta`). google-adk 2.7.1's sliding
window counts **unique `invocation_id`s**, so every real invocation counts as **2**:

```
new_invocation_ids after invocation 1 = [header_update, e-cc49c6e5]  → len 2 ≥ interval 2 → FIRE
```

Evidence: `compactionInterval: 2` compacted after **every** invocation. Consequence for the
production config: `compactionInterval: 15` really compacts every **~7–8** invocations, and
`overlapSize` pulls back to the synthetic invocation instead of a real one (compaction #2's
range starts at the header_update timestamp). Until fixed upstream (kagent-adk should not put
these into the session, or ADK should exclude non-`e-*` ids), halve your intended interval
semantics in your head — or report the issue.

### How `overlapSize` works (mechanism verified from source + observed)

When the next window compaction fires, its range **starts `overlapSize` invocations before the
previous compaction's end**, so consecutive summaries share that content (no cliff between
summary #1 and summary #2). In demo 1 the pullback landed on the synthetic `header_update`
invocation (see finding 1); in a clean session with interval=2/overlap=1 you'd see summary #2
re-cover exactly the last invocation of summary #1. Compaction #2's input excludes compaction
#1's summary event itself, but re-ingests the same raw invocations → the overlap content IS
in both summaries.

## Demo 2 — `tokenThreshold` + `eventRetentionSize` (session fresh, 1 invocation)

Config: threshold 1500, retention 3. Invocation: "show recent events in kagent" →
`k8s_get_events` returned ~14k chars → last LLM call reported **prompt=13,412 ≥ 1500 →
post-invocation threshold compaction fired** (summarizer: 370 in / 115 out, 582-char summary):

```
[0] hdr  [1] user  [2] call k8s_get_events  [3] resp k8s_get_events  [4] text(final)
[5] hdr                                        ← only event kept verbatim
[6] COMPACTION range=[1788068064.1 .. 1788068070.6] usage=370/115
```

### FINDING 2: threshold measures the *newest block*, not cumulative context (on this chain)

It fired at 13,412 while the whole session was only ~6k tokens of content — the reported
prompt count from the agentgateway→ARK chain tracks the newest message(s). Mechanically the
trigger works; semantically on this chain `tokenThreshold` means "compact when the last step
added ≥ N *new* tokens", not "when total context ≥ N". With production threshold 300000 it
will never fire. A cumulative-reporting provider would make it behave as documented.

### FINDING 3: `eventRetentionSize` — honored in memory, fuzzy in the persisted record

Source (v2.7.1 `_safe_token_compaction_split_index`): retention keeps the last N events AND
shifts the split earlier so no `function_response` is retained without its `function_call`
(avoids the Anthropic "tool_use without tool_result" 400 we hit in demo 1's broken session).
Observed: the persisted compaction range swallowed the retained tail (1 event kept instead of
3) because compaction ran against the in-memory session, which contains streaming partial
events that are never persisted. Retention works in-process; verify its effect via the next
LLM request's contents (DEBUG logs), not via the DB range alone.

### Also captured in demo 1 (failure-path gold)

The MCP toolset connection failed once → the model hallucinated a tool name `GetResources`
(not in the schema) → ADK killed the invocation with `Tool 'GetResources' not found` → the
session was left with a dangling `tool_use` → the NEXT invocation died with Anthropic 400
"`tool_use` ids were found without `tool_result` blocks" and the session is permanently
poisoned. Two takeaways: (a) compaction was NOT involved; (b) a failed tool call can poison
an Anthropic-format session — worth an upstream issue (ADK should synthesize an error
function_response instead of aborting).

---

# PART 3 — Complete knob inventory & support matrix (kagent CRD → kagent-adk → google-adk 2.7.1)

| CRD field (`spec.declarative.context.compaction.*`) | Passed through? | Honored by google-adk 2.7.1? | Verified live |
| --- | --- | --- | --- |
| `compactionInterval` | ✅ compiler.go:227 → config.json → EventsCompactionConfig | ✅ (pair: requires `overlapSize`) | ✅ demo 1 |
| `overlapSize` | ✅ | ✅ (pair: requires `compactionInterval`) | ✅ demo 1 (shifted by header_update bug) |
| `tokenThreshold` | ✅ (set via hasattr guard) | ✅ (pair: requires `eventRetentionSize`; preferred over window) | ✅ demo 2 |
| `eventRetentionSize` | ✅ (hasattr guard) | ✅ threshold-path only; keeps call+response pairs together | ✅ demo 2 (with partial-event caveat) |
| `summarizer.modelConfig` | ✅ separate model entry (compiler.go:243–255) | ✅ LlmEventSummarizer on that model | ✅ usage on every compaction event |
| `summarizer.promptTemplate` | ✅ compiler.go:234 → config.json `prompt_template` | ✅ kagent-adk appends a tool-name-preservation warning to your template | not exercised (uses kagent default template) |

That is the **complete** set — the v1alpha2 CRD schema exposes exactly these six; kagent-adk's
`ContextConfig` has nothing else, and `EventsCompactionConfig` (v2.7.1, `extra="forbid"`)
accepts nothing else. Hardcoded (not tunable via CRD): summarizer temperature 0, tool
arg/response truncation at **2,000 chars**, the summarizer's max-output, and the trigger
preference (threshold before window). Non-compaction context levers you control: system
prompt size, tool output truncation at the tool server (your biggest lever — the 60k-char
HTML), and `spec.declarative.stream` (streaming partials affect the retention bookkeeping).

Validator rules (google-adk 2.7.1, raises at startup if violated):
`compactionInterval` ⇔ `overlapSize` must be set together; `tokenThreshold` ⇔
`eventRetentionSize` must be set together; at least one pair required.

---

# PART 4 — Recommended config for large agents (10+ sub-agents, 50+ tools, ~40k-token system prompt)

Assumptions from live data: fixed prefix ≈ 55–65k tok (40k system+skills + tool/sub-agent schemas),
50–150k new tokens per heavy triage turn, 1M window, agentgateway chain (implicit prefix caching
works; reported prompt counts are non-cumulative; header_update bug halves interval/overlap).

```yaml
context:
  compaction:
    compactionInterval: 12     # ~6 real turns today (header_update bug); set 6 when fixed
    overlapSize: 4             # ~2 real turns re-summarized per boundary (synthetic ids eat half)
    tokenThreshold: 150000     # huge-turn safety net: fires when ONE step adds >=150k new tokens
    eventRetentionSize: 8      # last ~2 tool exchanges stay verbatim when threshold fires
    summarizer:
      modelConfig: agw-cheap-model-config
```

Structural guarantees: system prompt / skills / tool schemas are NOT session events — compaction
never touches them; ADK re-sends them on every call (and the prefix stays upstream-cached across
compactions). Keep skills in the system prompt or promptTemplate dataSources, never as early
conversation events or tool outputs — those DO get compacted.

Fleet notes: sub-agents are separate Agent CRs — apply this config to each of them; the parent's
config does not cover them. Parent sessions only store a sub-agent's returned message, so
sub-agent transcripts don't bloat the parent. Trim the parent's tool surface (toolNames) —
schemas are re-sent every call; push specialized tools into sub-agents.
