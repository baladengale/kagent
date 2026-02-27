# Kagent — Executive Product Announcement

**Date:** February 27, 2026
**Product:** kagent — Kubernetes-Native AI Agent Framework
**Current Version:** v0.7.18
**License:** Apache 2.0
**Status:** CNCF (Cloud Native Computing Foundation) Project — Active Development

---

## Product Overview

**kagent** is a Kubernetes-native framework for building, deploying, and managing AI agents at scale. It bridges the gap between the world's most popular container orchestration platform — Kubernetes — and the rapidly evolving landscape of AI agents, providing enterprises with a production-ready, declarative, and observable system for running intelligent automation.

kagent empowers platform engineering teams, DevOps organizations, and AI practitioners to define AI agents as Kubernetes custom resources, connect them to any LLM provider, equip them with powerful tools via the Model Context Protocol (MCP), and orchestrate multi-agent workflows — all within the familiar Kubernetes operational model.

### Why kagent?

- **Kubernetes is the standard.** Organizations already run their critical workloads on Kubernetes. kagent extends this investment to AI agents without introducing new infrastructure.
- **Declarative and GitOps-friendly.** Agents, tools, and model configurations are expressed as YAML manifests, fitting naturally into existing CI/CD and GitOps workflows.
- **Vendor-neutral LLM support.** Switch between OpenAI, Anthropic, Azure OpenAI, Google Vertex AI, Amazon Bedrock, Ollama, or custom providers without rewriting agent logic.
- **Enterprise-grade observability.** Full OpenTelemetry tracing integration lets teams monitor agent behavior end-to-end using their existing monitoring stack.
- **CNCF-backed and open source.** kagent is a CNCF project with an active and growing contributor community.

---

## Platform Architecture

kagent consists of four core components working in concert:

```
┌─────────────────────────────────────────────────────────────┐
│                     kagent Platform                         │
│                                                             │
│  ┌───────────┐  ┌──────────────┐  ┌──────────┐  ┌───────┐  │
│  │Controller │  │  HTTP/API    │  │  Web UI  │  │  CLI  │  │
│  │   (Go)    │──│  Server (Go) │──│(Next.js) │  │ (Go)  │  │
│  └───────────┘  └──────────────┘  └──────────┘  └───────┘  │
│        │               │                                    │
│        ▼               ▼                                    │
│  ┌───────────┐  ┌──────────────┐                            │
│  │ Database  │  │Agent Runtime │                            │
│  │(SQLite/PG)│  │  (Python)    │                            │
│  └───────────┘  └──────────────┘                            │
└─────────────────────────────────────────────────────────────┘
```

| Component | Technology | Role |
|-----------|-----------|------|
| **Controller** | Go / controller-runtime | Kubernetes operator that watches Agent, RemoteMCPServer, and ModelConfig custom resources and reconciles desired state |
| **HTTP/API Server** | Go | REST and gRPC API layer for the UI, CLI, and agent-to-agent communication |
| **Agent Runtime (ADK)** | Python | Executes agent logic using the Agent Development Kit, powered by Google ADK and LiteLLM |
| **Web UI** | TypeScript / Next.js | Interactive dashboard for managing agents, sessions, tools, and model configurations |
| **CLI** | Go / Cobra | Command-line interface for deploying, inspecting, and chatting with agents |
| **Database** | SQLite (dev) / PostgreSQL (prod) | Persistent storage for agent metadata, sessions, and long-term memory |

---

## Core Capabilities

### 1. Declarative Agent Management

Agents are defined as Kubernetes Custom Resources (CRDs) with system prompts, tool bindings, sub-agent references, and LLM configuration — all in standard YAML:

```yaml
apiVersion: kagent.dev/v1alpha2
kind: Agent
metadata:
  name: k8s-troubleshooter
spec:
  type: Declarative
  declarative:
    systemPrompt: "You are an expert Kubernetes troubleshooter..."
    modelConfig: gpt-4o
    tools:
      - type: McpServer
        mcpServer:
          name: kubernetes-tools
          kind: MCPServer
```

### 2. Multi-LLM Provider Support

kagent integrates with all major LLM providers through a unified ModelConfig resource:

- **OpenAI** (GPT-4o, GPT-4, o1, o3)
- **Anthropic** (Claude 3.5 Sonnet, Claude 3 Opus)
- **Azure OpenAI**
- **Google Vertex AI** (Gemini 2.5 Pro, Gemini 2.0 Flash)
- **Amazon Bedrock**
- **Ollama** (local/self-hosted models)
- **Custom providers** via AI gateways (LiteLLM, vLLM, etc.)

Organizations can switch providers per agent without code changes — simply update the ModelConfig reference.

### 3. Model Context Protocol (MCP) Integration

kagent provides first-class support for MCP, the emerging standard for connecting AI agents to tools:

- **RemoteMCPServer** CRD for connecting to any remote MCP-compatible tool server
- **Native KMCP integration** for running MCP servers as Kubernetes workloads
- **Built-in tool servers** for Kubernetes, Istio, Helm, Argo CD, Prometheus, Grafana, and Cilium
- **Automatic tool discovery** — the controller connects to MCP servers, lists available tools, and makes them selectable in the UI

### 4. Agent-to-Agent (A2A) Communication

kagent implements the Agent-to-Agent protocol, enabling multi-agent orchestration:

- Orchestrator agents can delegate tasks to specialized sub-agents
- A2A routing is handled independently from the controller, enabling horizontal scaling
- Dedicated A2A Registrar maintains consistent routing tables across replicas
- SSE streaming support with configurable timeouts for long-running LLM responses

### 5. Long-Term Memory for Agents

A production-ready memory system enables agents to learn and recall information across sessions:

- **Semantic search** using vector embeddings (pgvector for PostgreSQL, Turso/libSQL for SQLite)
- **Auto-save** summarizes and stores key information every 5 user turns
- **Explicit save** via `SaveMemoryTool` for agent-controlled persistence
- **Memory prefetch** injects relevant context at the start of each new session
- **TTL-based pruning** with popularity-based retention — frequently accessed memories are preserved

### 6. Human-in-the-Loop (HITL)

Agents can pause execution and request human approval before taking critical actions, with DataPart support and event-based synchronization for real-time interaction through the UI.

### 7. Guardrails

The kagent ADK includes a guardrails framework that lets organizations define safety boundaries for agent behavior, ensuring compliance and preventing unintended actions.

### 8. Voice Support

All agents now support voice interaction, enabling spoken input and output for more natural human-agent collaboration.

### 9. Built-In Observability

- Full **OpenTelemetry tracing** with standardized environment variable configuration
- Custom span attributes via span processors for detailed diagnostics
- OTEL Context propagation (not just LogRecord) for accurate distributed tracing
- Agent log timestamps for auditing and debugging

### 10. Security and Enterprise Readiness

- **SSL/TLS configuration** support on ModelConfig for encrypted LLM communication
- **OIDC Authentication** proposal for identity-based access control (in design)
- **STS (Security Token Service)** integration for secure, token-based agent identity
- **Pod security contexts** for Helm deployments
- **API key passthrough** for ModelConfig, enabling flexible credential management
- **OpenSSF Best Practices** badge
- Proactive CVE remediation (e.g., cryptography CVE-2026-26007, Go CVE patches)

### 11. Flexible Deployment

- **Helm charts** for production Kubernetes deployment with full customization
- **Agent tolerations and affinities** for fine-grained workload placement
- **Leader election** for horizontal scaling of the controller
- **File-based database credentials** for secret management integration
- **Configurable init containers** for MCP server setup
- Multi-platform binaries: Linux (amd64/arm64), macOS (amd64/arm64), Windows (amd64)

### 12. Developer Experience

- **Web UI** with auto-refreshing agent/MCP server lists, dynamic provider model discovery, skill display, and interactive chat
- **CLI** with `kagent deploy`, `kagent init`, `kagent chat` commands, `.env` file support, and TUI tool call visualization
- **Agent Development Kit (ADK)** in Python for building custom agent logic
- **CrewAI** and **LangGraph** integration packages for popular agent frameworks
- **Skills framework** (`kagent-skills`) for reusable agent capabilities
- **GitHub Codespaces** support for zero-setup development

---

## Recent Release Highlights (December 2025 – February 2026)

Over the past three months, kagent has shipped **19 releases** (v0.7.0 through v0.7.18), delivering significant new capabilities, improved reliability, and enhanced security. Below are the key achievements organized by theme.

### 🚀 Major Features Delivered

| Feature | Release | Impact |
|---------|---------|--------|
| **Long-Term Agent Memory** | v0.7.13 | Agents can now persist and recall information across sessions using vector-based semantic search, enabling learning from past interactions |
| **Agent-to-Agent Communication (A2A)** | v0.7.0 – v0.7.10 | Full A2A protocol support with dedicated routing, horizontal scaling, and MCP session isolation |
| **Guardrails Framework** | v0.7.13 | Safety boundaries for agent behavior to ensure compliance and prevent unintended actions |
| **Voice Support** | v0.7.14 | All agents now support voice interaction for natural human-agent collaboration |
| **Human-in-the-Loop (HITL)** | v0.7.0 | Agents can pause and request human approval with DataPart support and event-based sync |
| **KMCP Integration** | v0.7.0 | First-class support for running MCP tool servers as Kubernetes workloads |
| **Dynamic Provider Model Discovery** | v0.7.17 | UI automatically discovers and lists available models from configured providers |
| **SSL/TLS for Model Communication** | v0.7.5 | Encrypted LLM communication with configurable TLS settings per ModelConfig |
| **STS Client Integration** | v0.7.0 – v0.7.16 | Secure token-based agent identity with full library integration |

### 🏗️ Architecture and Scalability

| Improvement | Release | Details |
|-------------|---------|---------|
| **Decoupled A2A Handler Registration** | v0.7.6 | A2A routing managed independently from controller reconciliation, enabling horizontal scaling |
| **Leader Election** | v0.7.6 | Controller supports leader election when scaled to multiple replicas |
| **Subclassed ADK Executor** | v0.7.15 | Reduced code duplication by subclassing upstream Google ADK executor |
| **Separate Python Venv for Bash Tool** | v0.7.5 | Isolated execution environment for bash-based tools |
| **MCP Cancel Scope Fix** | Latest | Prevented async cancel scope corruption in multi-agent MCP tool calls |

### 🔒 Security Improvements

| Fix | Release | Details |
|-----|---------|---------|
| **CVE-2026-26007** | v0.7.14 | Upgraded cryptography library to address vulnerability |
| **Go CVE Patch** | v0.7.16 | Bumped Go to 1.25.7 for security fix |
| **Pod Security Context** | v0.7.15 | Added pod-level security context support for Helm deployments |
| **Nil Pointer Dereference Fix** | Latest | Prevented controller panic with nil A2A config |
| **Trivy Security Scanning** | v0.7.14 | Updated Trivy action with fail-on-vulnerability enforcement |

### 🎨 UI/UX Enhancements

| Enhancement | Release | Details |
|-------------|---------|---------|
| **Dynamic Provider Model Discovery** | v0.7.17 | Models auto-populated from provider configuration |
| **Auto-Refresh Agent/MCP Lists** | v0.7.18 | UI automatically refreshes after mutations |
| **Skills Display** | v0.7.5 | Agent skills visible and manageable in the chat sidebar |
| **Tool Call Visualization** | v0.7.0 | Tool calls displayed in both Web UI and CLI TUI |
| **Session Tool Call Display** | v0.7.10 | All tool calls and results shown when loading session data |
| **Configurable Log Level** | v0.7.15 | uvicorn log level configurable via environment variable |

### 📦 Ecosystem and Integrations

| Addition | Release | Details |
|----------|---------|---------|
| **CrewAI Package** | v0.7.5 | Published `kagent-crewai` for CrewAI framework integration |
| **Skills Package** | v0.7.12 | Published `kagent-skills` for reusable agent capabilities |
| **Grafana MCP SecretRef** | v0.7.14 | Helm chart support for Grafana MCP server with secret references |
| **MCPServer Timeout Propagation** | v0.7.14 | MCPServer CRD timeout properly propagated to RemoteMCPServer |
| **A2A SDK Bump** | v0.7.16 | Updated to a2a-sdk v0.3.23 |
| **OIDC Authentication Proposal** | v0.7.16 | Enhancement proposal for OIDC-based authentication |

### 👥 Community Growth

Over this period, the project welcomed **15+ new contributors**, demonstrating strong community engagement:

@fl-sean03, @dobesv, @MatteoMori8, @nujragan93, @AayushSaini101, @opspawn, @Daniel-Vaz, @lithammer, @CriszelGipala, @lets-call-n-walk, @ddjain, @sahitya-chandra, @Dhouti, @pmuir, @apexlnc, @antedotee, and others.

---

## Supported Platforms

| Platform | Architecture | Distribution |
|----------|-------------|-------------|
| Linux | amd64, arm64 | Container images, CLI binaries |
| macOS | amd64 (Intel), arm64 (Apple Silicon) | CLI binaries |
| Windows | amd64 | CLI binaries |
| Kubernetes | Any (via Helm) | Helm charts for production deployment |

**Python SDK:** Python 3.10+ (tested across 3.10, 3.11, 3.12)

---

## What's Next — Roadmap Highlights

The kagent team is actively working on the following areas:

- **OIDC Authentication** — Identity-based access control for multi-tenant deployments (EP-476, in design)
- **Enhanced Memory** — Hybrid search (vector + keyword), re-ranking, and memory consolidation
- **v1beta1 API stabilization** — Moving toward stable API contracts
- **Extended CNCF integration** — Continued advancement within the CNCF ecosystem

Track the full roadmap on the [kagent Project Board](https://github.com/orgs/kagent-dev/projects/3).

---

## Getting Started

```bash
# Install kagent CLI
curl -sL https://github.com/kagent-dev/kagent/releases/latest/download/kagent-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m) -o kagent
chmod +x kagent

# Deploy to your Kubernetes cluster
kagent deploy

# Access the Web UI
kubectl port-forward -n kagent svc/kagent-ui 3000:80
```

**Documentation:** [kagent.dev/docs](https://kagent.dev/docs/kagent)
**GitHub:** [github.com/kagent-dev/kagent](https://github.com/kagent-dev/kagent)
**Discord:** [discord.gg/Fu3k65f2k3](https://discord.gg/Fu3k65f2k3)
**CNCF Slack:** [#kagent](https://cloud-native.slack.com/archives/C08ETST0076)

---

<div align="center">
  <p><strong>kagent</strong> is a <a href="https://cncf.io">Cloud Native Computing Foundation</a> project.</p>
  <p>Licensed under <a href="https://www.apache.org/licenses/LICENSE-2.0">Apache 2.0</a></p>
</div>
