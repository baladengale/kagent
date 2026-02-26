# Token Refresh Without Agent Restart

## Problem

Remote MCP Servers (`RemoteMCPServer` CRD) connect to external endpoints (e.g. ServiceNow, Jira) that use short-lived token-based authentication.  These tokens can expire as frequently as every 30 minutes.

A Kubernetes CRD or controller can rotate the token by updating the underlying `Secret`, but previously the new token was **not picked up** by running agent pods without a manual restart.  Restarting agents during a live workflow causes:

* In-flight prompts / tool calls to fail.
* Loss of ongoing workflow state (especially in parent agents with long multi-step tasks).
* Cascade failures in agents that reference the same `RemoteMCPServer` as a tool.

---

## Architecture Overview

```
Secret (token) ──► RemoteMCPServer CR ──► Agent CR ──► ConfigMap ──► Agent Pod
                         ▲                     ▲              ▲
                    controller-               agent-       volume-
                    watches                  watches        mount
                    Secret                   RMCPS          update
```

The fix is applied at two levels: the **controller** (Go) and the **agent runtime** (Python).

---

## Solution

### 1. Controller Level – Reactive Secret Watching

The `RemoteMCPServerController` now **watches `Secret` resources** and enqueues any `RemoteMCPServer` that references a changed `Secret` through its `spec.headersFrom` field.

**Flow:**

1. External process rotates token → updates `Secret` in Kubernetes.
2. `RemoteMCPServerController` receives a `Secret` update event.
3. Controller finds every `RemoteMCPServer` whose `spec.headersFrom` contains a `valueFrom.type: Secret` entry pointing to the changed `Secret`.
4. Each matching `RemoteMCPServer` is re-reconciled immediately (instead of waiting up to 60 s).
5. Reconciliation resolves the new token, discovers tools, updates the database, and updates the `RemoteMCPServer` status.
6. The `AgentController` **already watches `RemoteMCPServer`** changes and enqueues all dependent `Agent` objects.
7. Agent reconciliation regenerates the agent's `ConfigMap` with the fresh token.

Key file: `go/internal/controller/remote_mcp_server_controller.go`

```go
// SetupWithManager now includes a Secret watch:
Watches(
    &corev1.Secret{},
    handler.EnqueueRequestsFromMapFunc(r.findRemoteMCPServersUsingSecret),
    builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
)
```

The helper `remoteMCPServerReferencesSecret` checks whether the `RemoteMCPServer` is in the same namespace as the `Secret` and has a `headersFrom` entry that references it by name.

### 2. Python Agent Runtime – Config Hot-Reload

When using the `static` CLI command (the default for Kubernetes-deployed agents), the `root_agent_factory` previously loaded `config.json` **once at process startup** and reused the same `AgentConfig` for every session.

The factory is now changed to **re-read `config.json` on every invocation**.  Because ConfigMaps are projected into the pod as volume files, Kubernetes automatically refreshes the file contents when the ConfigMap is updated (typically within 60 s).  Subsequent agent sessions therefore receive the latest token from the updated ConfigMap without a pod restart.

Key file: `python/packages/kagent-adk/src/kagent/adk/cli.py`

```python
def root_agent_factory() -> BaseAgent:
    # Re-read config on every invocation so credential updates in the
    # mounted ConfigMap (e.g. rotated auth tokens from RemoteMCPServer
    # headersFrom secrets) are picked up by new agent sessions without
    # requiring a pod restart.
    try:
        with open(config_path, "r") as f:
            fresh_config = AgentConfig.model_validate(json.load(f))
    except Exception:
        logger.exception("Failed to reload config, using cached config")
        fresh_config = agent_config
    return fresh_config.to_agent(app_cfg.name, sts_integration)
```

If the file cannot be read (e.g. transient I/O error), the factory falls back to the config loaded at startup.

---

## End-to-End Token Refresh Sequence

```
T+0   Secret updated with new token
T+0   RemoteMCPServer controller receives Secret event
T+1s  RemoteMCPServer reconciled; tools & DB updated; status updated
T+1s  Agent controller receives RemoteMCPServer change
T+2s  Agent reconciled; ConfigMap updated with new token
T≤60s Kubernetes propagates ConfigMap update to pod volume mount
T≤60s New agent session reads fresh config.json → uses new token
```

---

## Ongoing Sessions (In-Flight Workflows)

The fix **does not** interrupt running agent sessions.  Each session creates its own `Agent` instance inside the factory; ongoing sessions hold a reference to the old instance and continue unaffected until they complete.

However, if a session makes an MCP tool call **after the token has already expired** but **before the new token propagates**, that individual tool call may fail with an authentication error.  This is the same failure mode that existed before the fix; the improvement is that the recovery window is now automatic (≤60 s) rather than requiring a manual pod restart.

For use cases where even a single failed tool call is unacceptable, consider:

* Increasing the token TTL to exceed the worst-case propagation time.
* Implementing retry logic in the MCP tool server or the client.
* Using a dedicated service account or JWT exchange flow where the token is resolved at each HTTP request by an in-process credential provider.

---

## Configuration Reference

```yaml
apiVersion: kagent.dev/v1alpha2
kind: RemoteMCPServer
metadata:
  name: servicenow-mcp
  namespace: kagent
spec:
  url: https://servicenow.example.com/mcp
  protocol: STREAMABLE_HTTP
  headersFrom:
    - name: Authorization
      valueFrom:
        type: Secret          # <-- watched by the controller
        name: servicenow-token
        key: token
```

When the `servicenow-token` Secret is updated, the controller automatically detects the change and triggers reconciliation across the entire dependency chain.

---

## RBAC

The `RemoteMCPServerController` now declares the following additional RBAC permissions (compiled into the controller's `ClusterRole` via `kubebuilder` markers):

```
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch
```
