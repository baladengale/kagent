# agentgateway LLM failover (dola-seed-2.0-lite → glm-4.7)

Primary/fallback routing through your existing `kind-infra` agentgateway in the
local kind cluster, exposed as an OpenAI-compatible endpoint for ZCode.

## What this gives you

- **Active model:** `dola-seed-2.0-lite`
- **Fallback model:** `glm-4.7` (Z.AI, `api.z.ai`)
- **Failover:** if the active model returns 5xx or 429 (rate-limited), agentgateway
  evicts it and routes to glm-4.7 until it recovers.
- **Inbound endpoint:** OpenAI Chat Completions, `http://localhost:8080/v1/chat/completions`

## About the "Anthropic endpoint" goal (important)

You asked for a **default Anthropic endpoint** to use from ZCode / Claude Code.
agentgateway **cannot** provide that for these two models, because:

- Both `dola-seed-2.0-lite` and `glm-4.7` are **OpenAI-compatible** backends
  (`/v1/chat/completions`).
- agentgateway only accepts **Anthropic `/v1/messages` inbound** when the *backend*
  is the first-class **Anthropic** provider. It does **not** translate
  Anthropic Messages → OpenAI Completions. An OpenAI backend rejects `/v1/messages`.

So the realistic options are:

| Client | Works with this endpoint? | How |
| --- | --- | --- |
| **ZCode** | Yes | ZCode speaks OpenAI-compatible natively. Point it at `http://localhost:8080/v1` as an OpenAI-compatible provider. |
| **Claude Code** | Not directly | Claude Code expects an Anthropic `/v1/messages` endpoint. Put a small translation shim in front (e.g. `claude-code-router` / LiteLLM) that presents `/v1/messages` and forwards to `http://localhost:8080/v1/chat/completions`. |

The manifests below therefore expose an **OpenAI-compatible** endpoint. That fully
satisfies the failover requirement and ZCode access.

## Prerequisites: verify the `kind-infra` gateway

```bash
kubectl get gateway kind-infra -n agentgateway-system
# listeners: http (8080), https (8443); allowedRoutes.namespaces.from: All
```

## 1. Add your API keys

```bash
export DOLA_API_KEY=<your dola / Volcengine ARK key>
export ZAI_API_KEY=<your Z.AI key>
bash 00-create-secrets.sh
```

> The primary provider host assumes **Volcengine ARK** (`ark.cn-beijing.volces.com`,
> path `/api/v3/chat/completions`). If `dola-seed-2.0-lite` is served elsewhere,
> edit `host` / `port` / `path` / `model` in `10-agentgateway-backend.yaml`.
> ARK may also need an endpoint ID (e.g. `ep-xxxxxxxx`) as the `model` value.

## 2. Apply the manifests

```bash
kubectl apply -f 10-agentgateway-backend.yaml
kubectl apply -f 20-agentgateway-policy.yaml
kubectl apply -f 30-httproute.yaml
```

## 3. Check they programmed

```bash
kubectl get agentgatewaybackend,agentgatewaypolicy,httproute -n agentgateway-system
kubectl describe httproute llm-failover -n agentgateway-system | grep -A3 Conditions
# expect: ResolvedRefs=True, Programmed/Accepted=True
```

## 4. Reach it locally

The `kind-infra` LoadBalancer NodePorts are accessible on localhost in kind:

```bash
# OpenAI-compatible base URL for clients:
#   http://localhost:8080/v1      (via port-forward, below)  OR
#   http://localhost:31279/v1     (http NodePort directly)

kubectl -n agentgateway-system port-forward svc/kind-infra 8080:8080
```

Quick test (note `model` can be anything; the backend pins the upstream model):

```bash
curl http://localhost:8080/v1/chat/completions \
  -H 'content-type: application/json' \
  -d '{
    "model": "any",
    "messages": [{"role":"user","content":"say hi in one word"}]
  }' | jq
```

## 5. Point ZCode at it

Configure ZCode with an OpenAI-compatible provider:

```
base_url: http://localhost:8080/v1
api_key: any-nonempty-string        # gateway has no client auth in this setup
model: glm-4.7                     # or dola-seed-2.0-lite; backend pins upstream model anyway
```

If you want client auth on the gateway too, add an auth policy to the
HTTPRoute/AgentgatewayBackend (out of scope for this minimal setup).

## 6. Test failover

Temporarily force every active response to be unhealthy so traffic falls through
to glm-4.7:

```bash
kubectl patch agentgatewaypolicy llm-failover-health -n agentgateway-system --type=merge -p \
  '{"spec":{"backend":{"health":{"unhealthyCondition":"true"}}}}'
```

Send a request — it should still succeed (served by the fallback). Restore the
real condition when done:

```bash
kubectl patch agentgatewaypolicy llm-failover-health -n agentgateway-system --type=merge -p \
  '{"spec":{"backend":{"health":{"unhealthyCondition":"response.code >= 500 || response.code == 429"}}}}'
```

## Cleanup

```bash
kubectl delete -f 30-httproute.yaml
kubectl delete -f 20-agentgateway-policy.yaml
kubectl delete -f 10-agentgateway-backend.yaml
kubectl -n agentgateway-system delete secret dola-ark-secret zai-secret
```
