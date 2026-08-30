#!/usr/bin/env bash
# Create the provider API-key secrets referenced by the AgentgatewayBackend.
# Both providers use Bearer auth; agentgateway sends the secret's `Authorization`
# value as `Authorization: Bearer <value>` to the upstream.
set -euo pipefail

NS=agentgateway-system

: "${DOLA_API_KEY:?export DOLA_API_KEY (Volcengine ARK key, or your dola provider key)}"
: "${ZAI_API_KEY:?export ZAI_API_KEY (Z.AI / Zhipu key for glm-4.7)}"

kubectl -n "$NS" create secret generic dola-ark-secret \
  --from-literal=Authorization="$DOLA_API_KEY" -o yaml --dry-run=client | kubectl apply -f -

kubectl -n "$NS" create secret generic zai-secret \
  --from-literal=Authorization="$ZAI_API_KEY" -o yaml --dry-run=client | kubectl apply -f -

echo "Secrets ready in $NS: dola-ark-secret, zai-secret"
