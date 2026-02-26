# Token Refresh Architecture

## Problem Description

`RemoteMCPServer` resources use `headersFrom` (backed by a Kubernetes `Secret`) to carry
authentication tokens (e.g. OAuth2 bearer tokens that expire every 30 minutes).
A separate process may update the Secret with a fresh token, but without active
controller logic to detect the change, the running agent pod continues to use the
stale token until the pod is manually restarted.

Two interrelated issues existed:

1. **Token changes not picked up** — the `RemoteMCPServerController` had no watch on
   `corev1.Secret`, so a Secret update never triggered re-reconciliation of the
   `RemoteMCPServer`.
2. **Unnecessary pod restarts** — the `AgentController` watched `RemoteMCPServer` with no
   predicate, so every status-only update (e.g. tool discovery refresh) caused all
   dependent Agent pods to be re-reconciled and potentially restarted, interrupting
   in-flight workflows.

---

## Architecture Overview

```
┌──────────────────────┐    Secret change    ┌────────────────────────────┐
│  Kubernetes Secret   │ ─────────────────▶  │ RemoteMCPServerController  │
│  (auth token)        │                     │  • Secret watch             │
└──────────────────────┘                     │  • Re-reconcile RMCPS       │
                                             │  • Update DB ToolServer     │
                                             │  • Track TokenSecretHash    │
                                             └─────────────┬──────────────┘
                                                           │ status update
                                                           ▼
                                             ┌────────────────────────────┐
                                             │   RemoteMCPServer Status   │
                                             │  • TokenSecretHash          │
                                             │  • LastTokenRefreshTime     │
                                             └─────────────┬──────────────┘
                                                           │
                                             GenerationChangedPredicate
                                             (spec-only changes propagate)
                                                           │
                                                           ▼
                                             ┌────────────────────────────┐
                                             │     AgentController        │
                                             │  • Only re-reconciles on   │
                                             │    RMCPS spec changes      │
                                             │  • NO pod restart on       │
                                             │    token refresh           │
                                             └────────────────────────────┘
```

---

## Three-Layer Solution

### Layer 1 — Controller Watch (`remote_mcp_server_controller.go`)

The `RemoteMCPServerController` now watches `corev1.Secret` objects in addition to
`RemoteMCPServer` objects. When a Secret changes:

1. `findRemoteMCPServersReferencingSecret` lists all `RemoteMCPServer` objects and checks
   whether any `headersFrom[].valueFrom.secretKeyRef` references the changed Secret.
2. Matching `RemoteMCPServer` objects are enqueued for reconciliation.
3. The reconciler re-resolves the headers, creates a fresh MCP transport client with the
   new token, and refreshes the `ToolServer` record in the database.

**RBAC annotation added:**
```go
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
```

### Layer 2 — DB Update (`reconciler.go`)

In `ReconcileKagentRemoteMCPServer`, after successfully upserting the tool server:

1. The resolved header values are hashed with SHA-256.
2. The hash is compared to `RemoteMCPServerStatus.TokenSecretHash`.
3. If changed, `TokenSecretHash` and `LastTokenRefreshTime` are updated in the status.

This provides an audit trail for token refresh events and enables external tooling to
observe when a token was last refreshed.

### Layer 3 — Predicate Guard (`agent_controller.go`)

The `AgentController`'s watch on `RemoteMCPServer` now uses
`predicate.GenerationChangedPredicate{}`. This means:

- **Spec changes** (which bump `metadata.generation`) → agent re-reconciles → Deployment
  may be updated.
- **Status-only changes** (token refresh, condition updates, tool discovery) → generation
  unchanged → agent is **not** re-reconciled → **no pod restart**.

This is the key change that prevents running workflows from being disrupted during
routine token refresh cycles.

---

## New Status Fields (`RemoteMCPServerStatus`)

```go
// LastTokenRefreshTime records when the auth token was last successfully refreshed.
// +optional
LastTokenRefreshTime *metav1.Time `json:"lastTokenRefreshTime,omitempty"`

// TokenSecretHash records the hash of the last successfully resolved auth token secret.
// +optional
TokenSecretHash string `json:"tokenSecretHash,omitempty"`
```

These fields allow operators and monitoring tools to:
- Verify that token rotation is happening on schedule.
- Alert when `LastTokenRefreshTime` is older than the expected rotation interval.
- Detect stale tokens before they cause authentication failures.

---

## Configuration

| Concern | Mechanism |
|---------|-----------|
| Trigger token refresh | Update the Kubernetes Secret referenced in `headersFrom` |
| Periodic DB refresh interval | `RemoteMCPServerController` requeues every 60 s (existing behavior) |
| Observe last refresh | Check `status.lastTokenRefreshTime` on the `RemoteMCPServer` object |
| Verify token changed | Compare `status.tokenSecretHash` before and after Secret update |

---

## Troubleshooting

### Token not refreshing after Secret update

1. Check that the Secret is in the same namespace as the `RemoteMCPServer`.
2. Verify the `headersFrom[].valueFrom.type` is `Secret` (not `ConfigMap`).
3. Check controller logs for `RemoteMCPServerController` for reconcile errors.
4. Inspect `status.conditions` on the `RemoteMCPServer` — a `ReconcileFailed` condition
   indicates the controller could not resolve or use the new token.

### Agent pod restarting on every token refresh

Ensure the cluster is running the updated controller that includes the
`predicate.GenerationChangedPredicate{}` fix in `AgentController.SetupWithManager`.
The agent controller should only re-reconcile when the `RemoteMCPServer` spec changes,
not on status updates.

---

## Examples

### ServiceNow OAuth2 Token Setup

```yaml
# Secret managed by an external token rotation controller
apiVersion: v1
kind: Secret
metadata:
  name: servicenow-oauth-token
  namespace: kagent
type: Opaque
stringData:
  token: "Bearer eyJhbGci..."
---
apiVersion: kagent.dev/v1alpha2
kind: RemoteMCPServer
metadata:
  name: servicenow-mcp
  namespace: kagent
spec:
  url: https://myinstance.service-now.com/api/mcp
  protocol: STREAMABLE_HTTP
  headersFrom:
    - name: Authorization
      valueFrom:
        type: Secret
        name: servicenow-oauth-token
        key: token
```

When the `servicenow-oauth-token` Secret is updated with a new token, the
`RemoteMCPServerController` detects the change, re-resolves the headers, updates the
database `ToolServer` record with the fresh token, and updates `status.tokenSecretHash`
and `status.lastTokenRefreshTime`. No agent pod restart occurs.

### Generic Bearer Token

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: api-bearer-token
  namespace: kagent
type: Opaque
stringData:
  value: "Bearer sk-..."
---
apiVersion: kagent.dev/v1alpha2
kind: RemoteMCPServer
metadata:
  name: my-api-mcp
  namespace: kagent
spec:
  url: https://api.example.com/mcp
  headersFrom:
    - name: Authorization
      valueFrom:
        type: Secret
        name: api-bearer-token
        key: value
```
