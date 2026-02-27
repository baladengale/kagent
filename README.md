<div align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/kagent-dev/kagent/main/img/icon-dark.svg" alt="kagent" width="400">
    <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/kagent-dev/kagent/main/img/icon-light.svg" alt="kagent" width="400">
    <img alt="kagent" src="https://raw.githubusercontent.com/kagent-dev/kagent/main/img/icon-light.svg">
  </picture>
  <div>
    <a href="https://github.com/kagent-dev/kagent/releases">
      <img src="https://img.shields.io/github/v/release/kagent-dev/kagent?style=flat&label=Latest%20version" alt="Release">
    </a>
    <a href="https://github.com/kagent-dev/kagent/actions/workflows/ci.yaml">
      <img src="https://github.com/kagent-dev/kagent/actions/workflows/ci.yaml/badge.svg" alt="Build Status" height="20">
    </a>
      <a href="https://opensource.org/licenses/Apache-2.0">
      <img src="https://img.shields.io/badge/License-Apache2.0-brightgreen.svg?style=flat" alt="License: Apache 2.0">
    </a>
    <a href="https://github.com/kagent-dev/kagent">
      <img src="https://img.shields.io/github/stars/kagent-dev/kagent.svg?style=flat&logo=github&label=Stars" alt="Stars">
    </a>
     <a href="https://discord.gg/Fu3k65f2k3">
      <img src="https://img.shields.io/discord/1346225185166065826?style=flat&label=Join%20Discord&color=6D28D9" alt="Discord">
    </a>
    <a href="https://deepwiki.com/kagent-dev/kagent"><img src="https://deepwiki.com/badge.svg" alt="Ask DeepWiki"></a>
    <a href='https://codespaces.new/kagent-dev/kagent'>
      <img src='https://github.com/codespaces/badge.svg' alt='Open in Github Codespaces' style='max-width: 100%;' height="20">
    </a>
    <a href="https://www.bestpractices.dev/projects/10723"><img src="https://www.bestpractices.dev/projects/10723/badge" alt="OpenSSF Best Practices"></a>
  </div>
</div>

---

## 🚀 Platform Overview — kagent: AI Agent Platform for Kubernetes

> **Internal Announcement** — We are excited to introduce **kagent** as our organization's standard platform for building, deploying, and managing AI agents at scale on Kubernetes.

### What is kagent?

**kagent** is a **Cloud Native, Kubernetes-native AI agent platform** that enables teams to build intelligent, autonomous AI agents and deploy them as first-class Kubernetes workloads. It bridges the power of large language models (LLMs) with the operational reliability of Kubernetes — making AI agents observable, scalable, and manageable using the same tools and workflows your teams already use.

kagent is a **CNCF (Cloud Native Computing Foundation)** sandbox project, built on open standards and designed for enterprise production use.

---

### 🎯 Key Capabilities

| Capability | Description |
|---|---|
| **Agent Lifecycle Management** | Define, deploy, update, and delete AI agents declaratively using Kubernetes CRDs (`Agent` resource). Full GitOps-compatible lifecycle. |
| **Multi-LLM Provider Support** | Connect agents to OpenAI, Azure OpenAI, Anthropic Claude, Google Vertex AI, Ollama, or any OpenAI-compatible gateway (LiteLLM, AI Gateway). |
| **MCP Tool Integration** | Agents connect to any **Model Context Protocol (MCP)** server to access tools. Built-in MCP servers cover Kubernetes, Istio, Helm, Argo CD, Prometheus, Grafana, Cilium, and more. |
| **Remote MCP Servers** | Securely connect agents to external/remote MCP endpoints with token-based authentication and automatic secret-backed header management. |
| **Multi-Agent Orchestration** | Agents can invoke other agents as sub-agents, enabling complex multi-step and multi-agent workflows (A2A protocol). |
| **Long-Term Memory** | Agents can persist and recall information across sessions using configurable memory backends (e.g. Supabase). |
| **Declarative Configuration** | All platform resources — agents, model configs, tool servers — are Kubernetes custom resources, fully declarative and YAML-driven. |
| **Web UI** | A browser-based UI for creating, managing, and chatting with agents without writing code. |
| **CLI** | A powerful `kagent` CLI for managing agents and tools from the terminal, available for Linux, macOS, and Windows. |
| **Full Observability** | OpenTelemetry tracing built-in. Integrates with Prometheus, Grafana, Jaeger and any OTLP-compatible backend. |
| **Skills & OCI Packaging** | Package reusable agent skills as OCI container images or Git repositories, and attach them to agents on demand. |
| **PostgreSQL & SQLite Support** | Flexible database backend — use embedded SQLite for development or PostgreSQL for production-grade persistence. |
| **API Key Passthrough** | Support for federated identity — forward Bearer tokens from incoming requests directly to LLM providers, eliminating separate secret management. |
| **Security** | OpenSSF Best Practices certified. GitHub Advanced Security enabled. Apache 2.0 licensed. |

---

### 🏗️ Platform Architecture

kagent consists of **4 core components** running in your Kubernetes cluster:

```
┌─────────────────────────────────────────────────────────┐
│                     kagent Platform                      │
│                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────────┐  │
│  │  Controller  │  │    Engine    │  │      UI       │  │
│  │ (Go/K8s CRD) │  │ (Python ADK) │  │  (Next.js)    │  │
│  └──────────────┘  └──────────────┘  └───────────────┘  │
│         │                 │                  │           │
│         └─────────────────┴──────────────────┘          │
│                           │                              │
│                    ┌──────────────┐                      │
│                    │     CLI      │                      │
│                    │  (Go/Multi)  │                      │
│                    └──────────────┘                      │
└─────────────────────────────────────────────────────────┘
```

- **Controller** — Kubernetes controller watching `Agent`, `ModelConfig`, `ToolServer`, and `RemoteMCPServer` CRDs. Reconciles desired state continuously.
- **Engine** — Python-based runtime (ADK) that executes agent logic, manages LLM calls, and handles MCP tool invocations.
- **UI** — Web dashboard for no-code agent creation, management, and conversational interaction.
- **CLI** — Terminal-based management tool for power users and CI/CD pipelines.

---

### 🔧 Agent Lifecycle

```
Define (YAML CRD)  →  Deploy (kubectl / Helm / GitOps)  →  Run (Engine)
        ↓                                                        ↓
  ModelConfig                                          MCP Tool Calls
  ToolServers                                          LLM Inference
  RemoteMCPServers                                     Sub-Agent Invocation
        ↓                                                        ↓
  Update (reconciled automatically)              Observe (OTel / Grafana)
```

---

### 🔌 MCP Server Ecosystem

kagent ships with a rich set of **built-in MCP tool servers** for cloud-native operations:

| Domain | Tools Available |
|---|---|
| **Kubernetes** | Pods, Deployments, Services, ConfigMaps, Namespaces, Events, RBAC |
| **Istio** | Traffic management, Service mesh observability, mTLS |
| **Helm** | Chart install/upgrade/rollback, repo management |
| **Argo CD** | Application sync, rollouts, health status |
| **Prometheus** | Metric queries, alerting rules |
| **Grafana** | Dashboard management, annotations |
| **Cilium** | Network policies, endpoint visibility |
| **Custom / Remote** | Connect any MCP-compatible server via `RemoteMCPServer` CRD |

---

### 📦 Installation

kagent is deployed via **Helm** into any Kubernetes cluster:

```bash
helm install kagent oci://ghcr.io/kagent-dev/kagent/helm/kagent -n kagent --create-namespace
```

Supports Kind, GKE, EKS, AKS, and any CNCF-conformant Kubernetes distribution.

---

> 📖 **Full documentation:** [kagent.dev](https://kagent.dev) | 🐙 **Upstream project:** [github.com/kagent-dev/kagent](https://github.com/kagent-dev/kagent)

---

**kagent** is a Kubernetes native framework for building AI agents. Kubernetes is the most popular orchestration platform for running workloads, and **kagent** makes it easy to build, deploy and manage AI agents in Kubernetes. The **kagent** framework is designed to be easy to understand and use, and to provide a flexible and powerful way to build and manage AI agents.

<div align="center">
  <img src="img/hero.png" alt="Autogen Framework" width="500">
</div>

---

<!-- markdownlint-disable MD033 -->
<table align="center">
  <tr>
    <td>
      <a href="#getting-started"><b><i>Getting Started</i></b></a>
    </td>
    <td>
      <a href="#technical-details"><b><i>Technical Details</i></b></a>
    </td>
    <td>
      <a href="#get-involved"><b><i>Get Involved</i></b></a>
    </td>
    <td>
      <a href="#reference"><b><i>Reference</i></b></a>
    </td>
  </tr>
</table>
<!-- markdownlint-disable MD033 -->

---

## Getting Started

- [Quick Start](https://kagent.dev/docs/kagent/getting-started/quickstart)
- [Installation guide](https://kagent.dev/docs/kagent/introduction/installation)

## Technical Details

### Core Concepts

- **Agents**: Agents are the main building block of kagent. They are a system prompt, a set of tools and agents, and an LLM configuration represented with a Kubernetes custom resource called "Agent". 
- **LLM Providers**: Kagent supports multiple LLM providers, including [OpenAI](https://kagent.dev/docs/kagent/supported-providers/openai), [Azure OpenAI](https://kagent.dev/docs/kagent/supported-providers/azure-openai), [Anthropic](https://kagent.dev/docs/kagent/supported-providers/anthropic), [Google Vertex AI](https://kagent.dev/docs/kagent/supported-providers/google-vertexai), [Ollama](https://kagent.dev/docs/kagent/supported-providers/ollama) and any other [custom providers and models](https://kagent.dev/docs/kagent/supported-providers/custom-models) accessible via AI gateways. Providers are represented by the ModelConfig resource.
- **MCP Tools**: Agents can connect to any MCP server that provides tools. Kagent comes with an MCP server with tools for Kubernetes, Istio, Helm, Argo, Prometheus, Grafana, Cilium, and others. All tools are Kubernetes custom resources (ToolServers) and can be used by multiple agents.
- **Observability**: Kagent supports [OpenTelemetry tracing](https://kagent.dev/docs/kagent/getting-started/tracing), which allows you to monitor what's happening with your agents and tools.

### Core Principles

- **Kubernetes Native**: Kagent is designed to be easy to understand and use, and to provide a flexible and powerful way to build and manage AI agents.
- **Extensible**: Kagent is designed to be extensible, so you can add your own agents and tools.
- **Flexible**: Kagent is designed to be flexible, to suit any AI agent use case.
- **Observable**: Kagent is designed to be observable, so you can monitor the agents and tools using all common monitoring frameworks.
- **Declarative**: Kagent is designed to be declarative, so you can define the agents and tools in a YAML file.
- **Testable**: Kagent is designed to be tested and debugged easily. This is especially important for AI agent applications.

### Architecture

The kagent framework is designed to be easy to understand and use, and to provide a flexible and powerful way to build and manage AI agents.

<div align="center">
  <img src="img/arch.png" alt="kagent" width="500">
</div>

Kagent has 4 core components:

- **Controller**: The controller is a Kubernetes controller that watches the kagent custom resources and creates the necessary resources to run the agents.
- **UI**: The UI is a web UI that allows you to manage the agents and tools.
- **Engine**: The engine runs your agents using [ADK](https://google.github.io/adk-docs/).
- **CLI**: The CLI is a command-line tool that allows you to manage the agents and tools.

## Get Involved

_We welcome contributions! Contributors are expected to [respect the kagent Code of Conduct](https://github.com/kagent-dev/community/blob/main/CODE-OF-CONDUCT.md)_

There are many ways to get involved:

- 🐛 [Report bugs and issues](https://github.com/kagent-dev/kagent/issues/)
- 💡 [Suggest new features](https://github.com/kagent-dev/kagent/issues/)
- 📖 [Improve documentation](https://github.com/kagent-dev/website/)
- 🔧 [Submit pull requests](/CONTRIBUTING.md)
- ⭐ Star the repository
- 💬 [Help others in Discord](https://discord.gg/Fu3k65f2k3)
- 💬 [Join the kagent community meetings](https://calendar.google.com/calendar/u/0?cid=Y183OTI0OTdhNGU1N2NiNzVhNzE0Mjg0NWFkMzVkNTVmMTkxYTAwOWVhN2ZiN2E3ZTc5NDA5Yjk5NGJhOTRhMmVhQGdyb3VwLmNhbGVuZGFyLmdvb2dsZS5jb20)
- 🤝 [Share tips in the CNCF #kagent slack channel](https://cloud-native.slack.com/archives/C08ETST0076)
- 🔒 [Report security concerns](SECURITY.md)

### Roadmap

`kagent` is currently in active development. You can check out the full roadmap in the project Kanban board [here](https://github.com/orgs/kagent-dev/projects/3).

### Local development

For instructions on how to run everything locally, see the [DEVELOPMENT.md](DEVELOPMENT.md) file.

### Contributors

Thanks to all contributors who are helping to make kagent better.

<a href="https://github.com/kagent-dev/kagent/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=kagent-dev/kagent" />
</a>

### Star History

<a href="https://www.star-history.com/#kagent-dev/kagent&Date">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=kagent-dev/kagent&type=Date&theme=dark" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=kagent-dev/kagent&type=Date" />
   <img alt="Star history of kagent-dev/kagent over time" src="https://api.star-history.com/svg?repos=kagent-dev/kagent&type=Date" />
 </picture>
</a>

## Reference

### License

This project is licensed under the [Apache 2.0 License.](/LICENSE)

---

<div align="center">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/cncf/artwork/refs/heads/main/other/cncf/horizontal/color-whitetext/cncf-color-whitetext.svg">
      <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/cncf/artwork/refs/heads/main/other/cncf/horizontal/color/cncf-color.svg">
      <img width="300" alt="Cloud Native Computing Foundation logo" src="https://raw.githubusercontent.com/cncf/artwork/refs/heads/main/other/cncf/horizontal/color-whitetext/cncf-color-whitetext.svg">
    </picture>
    <p>kagent is a <a href="https://cncf.io">Cloud Native Computing Foundation</a> project.</p>
</div>