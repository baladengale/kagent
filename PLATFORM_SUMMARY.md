# kagent — Executive Platform Summary

> *Prepared for internal organisational announcement · February 2026*

---

## 🚀 Executive Summary

**kagent** is our organisation's adopted open-source, **Kubernetes-native AI Agent Platform** — a [CNCF Sandbox project](https://cncf.io) that enables engineering teams to build, deploy, and operate intelligent AI agents as first-class Kubernetes workloads. Launched in January 2025, kagent has matured rapidly into a production-ready platform with over **640 merged pull requests** in just over 13 months of active development.

In one sentence: **kagent brings AI agents into the Kubernetes control plane — making them as manageable, observable, and reliable as any other cloud-native workload.**

Whether your team needs an autonomous Kubernetes operator, a self-healing infrastructure agent, an AI-assisted deployment pipeline, or a multi-step intelligent workflow engine — kagent provides the platform to build, run, and govern it.

---

## 🏗️ What is kagent?

kagent is a **declarative, GitOps-compatible AI agent platform** built on top of Kubernetes. Agents are defined as Kubernetes Custom Resources (CRDs), meaning they can be version-controlled, templated, and managed with the same tooling your teams already use for infrastructure — Helm, Argo CD, kubectl, and standard CI/CD pipelines.

The platform abstracts the complexity of LLM integration, tool orchestration, multi-agent coordination, memory management, and observability — so your engineers focus on building agent *behaviour*, not plumbing.

---

## 🧩 Platform Architecture — 4 Core Components

| Component | Technology | Role |
|---|---|---|
| **Controller** | Go / Kubernetes CRDs | Watches Agent, ModelConfig, ToolServer, RemoteMCPServer CRDs. Reconciles desired state continuously. |
| **Engine (kagent-adk)** | Python / Google ADK | Executes agent logic, manages LLM calls, handles MCP tool invocations |
| **UI** | Next.js / React | Browser-based dashboard for no-code agent creation, management, and interactive chat |
| **CLI** | Go (multi-platform) | Terminal management for agents and tools across Linux, macOS, Windows |

**Deploy everything with a single command:**

```bash
helm install kagent oci://ghcr.io/kagent-dev/kagent/helm/kagent \
  -n kagent --create-namespace
```

---

## ⚡ Key Platform Capabilities

### 1. Agent Lifecycle Management
Agents are **Kubernetes Custom Resources**. Create, update, scale, and retire agents using standard Kubernetes tooling. Agents carry a system prompt, LLM configuration, tool assignments, and optional skill packages — all defined declaratively in YAML.

### 2. Model Context Protocol (MCP) — Tool Ecosystem
kagent adopts **MCP (Model Context Protocol)** as its native tool integration standard. Agents connect to any MCP server to gain capabilities. kagent ships with **built-in MCP tool servers** covering:

| Domain | Tools Available |
|---|---|
| **Kubernetes** | Pods, Deployments, Services, ConfigMaps, RBAC, Events, Namespaces |
| **Istio** | Traffic management, mTLS, service mesh observability |
| **Helm** | Chart install/upgrade/rollback, repository management |
| **Argo CD** | Application sync, rollouts, health status |
| **Prometheus** | Metric queries, alerting rules |
| **Grafana** | Dashboard management with service account token auth |
| **Cilium** | Network policies, endpoint visibility |
| **AWS EKS** | Kubernetes resource management via AWS |
| **Custom / Remote** | Connect any MCP endpoint via `RemoteMCPServer` CRD |

### 3. Agent-to-Agent (A2A) Orchestration
Agents can **invoke other agents as sub-agents**, enabling complex, hierarchical, multi-step workflows. A parent orchestrator agent can decompose a task and delegate to specialised child agents — all within a single Kubernetes cluster or across namespaces. MCP session isolation ensures each A2A invocation is stateless and safe.

### 4. Multi-LLM Provider Support
kagent is **LLM-agnostic**. Teams can choose the best model for each agent use case:

- OpenAI (GPT-4o, o3, o1)
- Azure OpenAI
- Anthropic Claude (direct + via AWS Bedrock)
- Google Vertex AI / Gemini
- AWS Bedrock (Meta Llama, Amazon Titan)
- Ollama (local/self-hosted models)
- LiteLLM / AI Gateway (proxy to any provider)
- Custom model endpoints

### 5. Skills — Reusable Agent Behaviours
**Skills** are packaged, reusable agent capabilities distributed as **OCI container images** or **Git repositories**. A skill can contain custom Python tools, domain logic, or specialised knowledge. Skills are mounted into agents at deploy time, enabling:
- Version-controlled, auditable agent capabilities
- Shared skill libraries across teams
- Near-instant loading via a lightweight (~30MB) init image

### 6. Long-Term Agent Memory
Agents can **persist and recall information across sessions** using the `MemoryStore` CRD. This transforms agents from stateless assistants into persistent intelligent workers that accumulate context, user preferences, and operational knowledge over time. Backed by Supabase (with extensible backend support).

### 7. Federated Identity & API Key Passthrough
kagent supports **API Key Passthrough** — forwarding your organisation's SSO/OIDC Bearer tokens directly to LLM providers. This eliminates per-agent secret management and integrates natively with your existing IAM infrastructure. Custom headers can also be forwarded into MCP tool calls for fine-grained access control.

### 8. Full Observability — OpenTelemetry Native
Every agent run is fully traced end-to-end using **OpenTelemetry (OTEL)**. Plug into any OTLP-compatible backend — Grafana, Jaeger, Datadog, or your organisation's existing observability stack. All 70+ platform environment variables are self-documenting via `kagent env`.

---

## 🔒 Enterprise & Security Posture

| Capability | Status |
|---|---|
| OpenSSF Best Practices certified | ✅ |
| GitHub Advanced Security on all PRs | ✅ |
| Apache 2.0 Open Source License | ✅ |
| CNCF Sandbox Project | ✅ |
| Pod security contexts & runAsUser | ✅ |
| Custom Kubernetes Service Account per agent | ✅ |
| imagePullSecrets for private registries | ✅ |
| HTTP/HTTPS proxy support (air-gapped networks) | ✅ |
| Continuous CVE patching (Go, Python, Node.js) | ✅ |
| HA Controller with leader election | ✅ |
| PostgreSQL production backend | ✅ |
| OIDC Authentication | 🔄 In progress |

---

## 📈 Platform Growth — By the Numbers

| Metric | Value |
|---|---|
| **Project launched** | January 2025 |
| **This summary prepared** | February 2026 |
| **Total merged PRs** | 640+ |
| **Sustained velocity** | ~50 PRs / month |
| **Primary languages** | Go, Python, TypeScript |
| **LLM providers supported** | 8+ |
| **Built-in MCP tool domains** | 8+ |
| **Deployment method** | Single `helm install` |
| **CNCF status** | Sandbox project |

---

## 🗓️ Innovation Timeline — Major Milestones

| Period | Milestone |
|---|---|
| **Jan – Mar 2025** | Core Kubernetes controller, Agent CRD, MCP ToolServer CRD, Web UI, CLI, Multi-LLM support, Helm chart |
| **Apr – Sep 2025** | Agent-to-Agent (A2A) protocol, A2A session isolation, LiteLLM gateway support |
| **Oct – Dec 2025** | Agent Skills (OCI packaging), OpenAI Agents SDK support, BYO Agent framework, Cross-namespace tools, HA leader election, Security hardening |
| **Jan 2026** | AWS Bedrock support, API Key Passthrough / federated identity, Dynamic LLM provider discovery, PostgreSQL backend, OpenTelemetry standardisation |
| **Feb 2026** | Long-term agent memory, Git-based skill fetching, Lightweight skills init image (~30MB), Go ADK declarative runtime, Customisable LLM call timeouts |

---

## 🌐 Resources

| Resource | Link |
|---|---|
| **Documentation** | [kagent.dev](https://kagent.dev) |
| **Quick Start** | [kagent.dev/docs/kagent/getting-started/quickstart](https://kagent.dev/docs/kagent/getting-started/quickstart) |
| **Upstream GitHub** | [github.com/kagent-dev/kagent](https://github.com/kagent-dev/kagent) |
| **Discord Community** | [discord.gg/Fu3k65f2k3](https://discord.gg/Fu3k65f2k3) |
| **CNCF Slack** | [#kagent on cloud-native.slack.com](https://cloud-native.slack.com/archives/C08ETST0076) |
| **Project Roadmap** | [GitHub Kanban Board](https://github.com/orgs/kagent-dev/projects/3) |

---

*kagent is a [Cloud Native Computing Foundation](https://cncf.io) Sandbox project. Licensed under [Apache 2.0](/LICENSE).*
