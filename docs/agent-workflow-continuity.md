# Agent Workflow Continuity

## Overview

This document describes how kagent preserves in-flight agent workflows during
infrastructure events such as token refresh cycles and `RemoteMCPServer` configuration
updates.

---

## What Happens to In-Flight Workflows During Token Refresh

### Before the Fix

Previously, any change to a `RemoteMCPServer` (including routine status updates from
tool discovery or token refresh) would trigger re-reconciliation of all dependent
`Agent` objects. This could cause rolling updates to the agent `Deployment`, which
**terminated in-progress sessions** without warning.

### After the Fix

The `AgentController` now uses `predicate.GenerationChangedPredicate{}` on its
`RemoteMCPServer` watch. The Kubernetes resource generation counter increments only
when the **spec** of a resource changes — status updates do not increment it.

**Result:** Token refresh cycles only update:
1. The Kubernetes `Secret` (managed externally)
2. The `RemoteMCPServer` status fields (`tokenSecretHash`, `lastTokenRefreshTime`)
3. The database `ToolServer` record (headers/token updated for future tool calls)

None of these changes bump the `RemoteMCPServer` generation, so the `AgentController`
does **not** re-reconcile, the `Deployment` is **not** updated, and no pod restart
occurs.

---

## Multi-Agent Dependency Graph and Token Propagation

Kagent supports agent-as-tool (A2A) configurations where a parent agent invokes a
child agent as a tool. Both agents may share the same `RemoteMCPServer`.

```
Parent Agent
    │
    ├─── Tool: RemoteMCPServer A  (e.g. ServiceNow MCP)
    │
    └─── Tool: Child Agent
              │
              └─── Tool: RemoteMCPServer A  (same server)
```

When `RemoteMCPServer A`'s token is refreshed:

1. The `RemoteMCPServerController` detects the Secret change.
2. The database `ToolServer` record for `RemoteMCPServer A` is updated with the new
   token.
3. **Neither** the Parent Agent nor the Child Agent pods are restarted.
4. Subsequent tool calls from either agent use the updated token from the database.

This means long-running workflows that span multiple agents and multiple tool calls
continue without interruption.

---

## Graceful Degradation Strategies

### Token Expiry Before Refresh

If a token expires before the rotation controller updates the Secret:

- Tool calls to the `RemoteMCPServer` will fail with authentication errors.
- The `RemoteMCPServer` status condition will show `Accepted: False` with the error
  message.
- The agent will receive tool call errors and can handle them via its system prompt
  instructions (e.g., retry, report to user).
- Once the Secret is updated, the `RemoteMCPServerController` will re-reconcile and
  restore `Accepted: True`.

### Temporary Network Failures

The 60-second periodic requeue in `RemoteMCPServerController` ensures that transient
failures are retried automatically. The last successfully discovered tool list is
preserved in the database and used as a fallback while connectivity is degraded.

---

## Configuring Agents for Long-Running Workflow Resilience

### System Prompt Guidance

Add instructions in the agent's system prompt to handle transient tool failures:

```
If a tool call fails with an authentication error, wait 30 seconds and retry once
before reporting the failure to the user. The authentication token may be in the
process of being refreshed.
```

### Monitoring Token Freshness

Check when the token was last refreshed:

```bash
kubectl get remotemcpserver servicenow-mcp -n kagent \
  -o jsonpath='{.status.lastTokenRefreshTime}'
```

Alert when the refresh time is older than your token TTL:

```bash
# Example: alert if token hasn't been refreshed in 35 minutes (TTL is 30 min)
kubectl get remotemcpserver servicenow-mcp -n kagent \
  -o jsonpath='{.status.conditions[?(@.type=="Accepted")].status}'
```

### Observing Token Hash Changes

To verify a token rotation completed successfully, compare the `tokenSecretHash`
before and after updating the Secret:

```bash
# Before rotation
kubectl get remotemcpserver servicenow-mcp -n kagent \
  -o jsonpath='{.status.tokenSecretHash}'

# Update the Secret with a new token...

# After rotation (should show a different hash within ~5 seconds)
kubectl get remotemcpserver servicenow-mcp -n kagent \
  -o jsonpath='{.status.tokenSecretHash}'
```

---

## Summary of Changes

| Component | Change | Effect |
|-----------|--------|--------|
| `RemoteMCPServerController` | Added `corev1.Secret` watch | Token changes detected without pod restart |
| `AgentController` | Added `GenerationChangedPredicate` to RMCPS watch | Status-only RMCPS changes no longer restart agent pods |
| `RemoteMCPServerStatus` | Added `TokenSecretHash`, `LastTokenRefreshTime` | Observability for token rotation events |
| `reconciler.go` | Compute headers hash, update status on change | Tracks when token was last refreshed |
