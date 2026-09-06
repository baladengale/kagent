# Running kagent with Agent Substrate on a local kind cluster

The v1alpha3 architecture (Harness/AgentTemplate CRDs, PostgreSQL-backed
AgentInstances, A2A gateway) runs agents **only** on Agent Substrate: the
controller dials the substrate ate-api at startup and exits if it is
unreachable. A local cluster therefore needs the Substrate platform
(`ate-system`) installed and bootstrapped **before** kagent.

The canonical bootstrap sequence is the E2E job in
[.github/workflows/ci.yaml](../.github/workflows/ci.yaml) ("Install Agent
Substrate" + "Install Kagent"). This document is the same recipe adapted for
a long-lived local kind cluster, including the steps CI gets for free from a
fresh cluster and the failure modes hit in practice. If you use the
[kind-infra](https://github.com/baladengale/kind-infra) repo,
`make substrate-create` automates all of it — see its
`docs/substrate-kagent.md`; this page documents the underlying commands.

## 1. Cluster prerequisites: podcert APIs

Substrate identities are built on `podCertificate` projected volumes and
`ClusterTrustBundle` projections — **beta APIs, off by default on kubernetes
>= 1.34**. Both the API server and the kubelet need:

- `--runtime-config=certificates.k8s.io/v1beta1=true` (API server)
- feature gates `ClusterTrustBundle`, `ClusterTrustBundleProjection`,
  `PodCertificateRequest` on **both** components

Notes that cost us a debugging session:

- `--runtime-config` alone is not enough — `PodCertificateRequest` storage is
  additionally gated (`PodCertificateRequest storage is disabled because the
  PodCertificateRequest feature gate is disabled` in the apiserver log).
- The kubelet's gate is also named `PodCertificateRequest`; there is no
  `PodCertificate` gate and unknown names panic the kubelet.
- `ClusterTrustBundleProjection` depends on `ClusterTrustBundle`.
- Without `ClusterTrustBundleProjection` on the **API server**, pod specs are
  accepted but the projected volume sources are silently **stripped to `{}`**
  at admission — pods then mount empty volumes and every substrate component
  crash-loops on a missing `trust-bundle.pem`.

**Fresh clusters:** [scripts/kind/kind-config.yaml](../scripts/kind/kind-config.yaml)
already declares the `featureGates` + `runtimeConfig` —
`make create-kind-cluster` gets everything at creation time.

**Existing clusters** (live patch, no recreation): edit the static manifest
and kubelet config inside the node container, then let both restart:

```bash
NODE=kagent-control-plane   # <cluster-name>-control-plane

# API server
docker exec $NODE sh -c '
  f=/etc/kubernetes/manifests/kube-apiserver.yaml
  grep -q certificates.k8s.io/v1beta1 "$f" && exit 0
  sed -i "s|^    - --runtime-config=$|    - --runtime-config=certificates.k8s.io/v1beta1=true|" "$f"
  sed -i "/^    - --runtime-config=certificates/i\\
    - --feature-gates=ClusterTrustBundle=true,ClusterTrustBundleProjection=true,PodCertificateRequest=true" "$f"
'
# kubelet (kind nodes run systemd)
docker exec $NODE sh -c '
  f=/var/lib/kubelet/config.yaml
  grep -q "^featureGates:" "$f" && exit 0
  printf "featureGates:\n  ClusterTrustBundle: true\n  ClusterTrustBundleProjection: true\n  PodCertificateRequest: true\n" >> "$f"
  systemctl restart kubelet
'
kubectl get --raw /apis/certificates.k8s.io/v1beta1 >/dev/null && echo "podcert APIs enabled"
```

The patches live in the node container's writable layer — they survive
restarts of the container but not its recreation.

## 2. Install and bootstrap the Substrate platform

Same sequence as CI (chart version = the substrate release, e.g. `0.0.25`):

```bash
SUBSTRATE_VERSION=0.0.25
helm upgrade --install substrate-crds \
  oci://ghcr.io/kagent-dev/substrate/helm/substrate-crds \
  --version "$SUBSTRATE_VERSION" --namespace ate-system --create-namespace

helm upgrade --install substrate \
  oci://ghcr.io/kagent-dev/substrate/helm/substrate \
  --version "$SUBSTRATE_VERSION" --namespace ate-system \
  --set-string 'atelet.extraArgs[0]=--localhost-registry-replacement=kind-registry:5000'
```

Then the one-time PKI bootstrap with `kubectl-ate` (grab the binary for your
platform from the substrate release page) and a converging upgrade:

```bash
curl -fsSL -o kubectl-ate "https://github.com/kagent-dev/substrate/releases/download/v${SUBSTRATE_VERSION}/kubectl-ate-darwin-arm64" # or linux-amd64
chmod +x kubectl-ate
./kubectl-ate --context kind-kagent admin make-ca-pool --ca-id=1 --name=service-dns-ca-pool --secret-namespace=podcertificate-controller-system
./kubectl-ate --context kind-kagent admin make-ca-pool --ca-id=1 --name=pod-identity-ca-pool --secret-namespace=podcertificate-controller-system
./kubectl-ate --context kind-kagent admin make-jwt-pool --key-id=1 --name=actor-id-jwt-pool --secret-namespace=ate-system
./kubectl-ate --context kind-kagent admin make-ca-pool --ca-id=1 --name=actor-id-ca-pool --secret-namespace=ate-system
actor_id_ca_root="$(kubectl get secret actor-id-ca-pool -n ate-system -o jsonpath='{.data.pool}' | base64 --decode | jq -r '.CAs[0].RootCertificateDER' | base64 --decode | openssl x509 -inform der -outform pem)"
kubectl create secret generic actor-id-ca-certs -n ate-system --from-literal=ca.crt="${actor_id_ca_root}"
kubectl create configmap ate-api-authentication -n ate-system \
  --from-literal=authentication.yaml=$'actorIdentityJWTProvider: kubernetes\njwtProviders:\n- name: kubernetes\n  issuer: https://kubernetes.default.svc\n  audiences: [api.ate-system.svc]\n  certificateAuthorityFile: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt\n  discoveryTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token\n'
helm upgrade substrate oci://ghcr.io/kagent-dev/substrate/helm/substrate \
  --version "$SUBSTRATE_VERSION" --namespace ate-system --reuse-values --wait --timeout 5m
```

If substrate was installed **before** the podcert APIs were enabled, its
workload templates are poisoned (stripped volume sources) — delete the
release first and install fresh; a rollout restart does not fix stored
templates. See [step 1 notes](#1-cluster-prerequisites-podcert-apis).

## 3. Build kagent and install it wired to the platform

```bash
make build   # controller, ui, golang-adk, claude-harness, codex-harness (+ kagent-adk)
```

Then install with the substrate values (upstream CI uses exactly these):

```bash
make helm-install-provider KAGENT_HELM_EXTRA_ARGS='--set controller.substrate.enabled=true \
  --set controller.substrate.ateApiEndpoint=dns:///api.ate-system.svc:443 \
  --set controller.substrate.atenetRouterURL=http://atenet-router.ate-system.svc:80 \
  --set controller.substrate.defaultWorkerPool.name=kagent-default \
  --set substrateWorkerPool.create=true \
  --set substrateWorkerPool.replicas=2 \
  --set-string substrateWorkerPool.workerImage=ghcr.io/kagent-dev/substrate/ateom-gvisor:v0.0.25'
```

Gotchas:

- Pass overrides as **make arguments** (`make target VAR=...`), not
  environment prefixes (`VAR=... make target`) — the root `Makefile` does
  `-include .env`, and a makefile assignment beats an environment-origin
  variable, so a prefixed `KAGENT_HELM_EXTRA_ARGS` is silently ignored.
- The chart's `substrateWorkerPool.workerImage` is **required** when
  `substrateWorkerPool.create=true` (the template `fail`s without it). With a
  local registry + substrate's atelet rewrite, use the
  `localhost:5001/...` mirror ref instead of the ghcr one.
- The `substrate` chart dependency in `helm/kagent/Chart.yaml` (the subchart
  path) is a different install mode from the standalone `ate-system` install
  above; upstream CI uses the standalone path, and so does this recipe.

## 4. Verify

```bash
kubectl -n kagent get pods                    # controller Running, kagent-default worker pods Running
kubectl get workerpool -n kagent              # kagent-default 2/2 ready
kubectl -n kagent logs deploy/kagent-controller | grep -i error   # expect nothing
kubectl get --raw /apis/certificates.k8s.io/v1beta1 >/dev/null && echo ok
```

Agents now run as AgentInstances on Substrate actors: creating an agent in
the UI compiles an ate-api ActorTemplate, schedules an Actor on the
`kagent-default` WorkerPool, and the actor suspends (snapshot) between A2A
task boundaries, resuming on the next request.

## 5. Upgrading a cluster from a pre-substrate kagent

Older clusters (v1alpha2-era) need three one-time cleanups, all dev-cluster
safe but destructive to legacy data:

1. **Legacy CRDs**: v1alpha2 objects in `agents`, `modelconfigs`,
   `modelproviderconfigs`, `remotemcpservers`, ... block the CRD upgrade
   (`status.storedVersions` still names `v1alpha2`). Back up, delete the
   objects and stale CRDs, reinstall, recreate wanted ModelConfigs as
   `v1alpha3`. If helm then fails computing the old release's diff
   (`resource mapping not found for kind "Agent"`), delete the release's
   labeled resources + `sh.helm.release.v1.kagent.*` secrets and install
   fresh.
2. **Database**: the goose migrator refuses golang-migrate tables
   (`unsupported migration table`) — delete the `kagent-postgresql` PVC and
   restart postgres for a fresh database.
3. **Startup race**: right after install the controller can crash once on
   `connection refused` to postgres — it self-heals on retry.
