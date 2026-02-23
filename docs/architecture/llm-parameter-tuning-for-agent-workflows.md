# LLM Parameter Tuning Guide for Kagent Agent Workflows

This document provides recommended LLM parameter configurations for complex agent workflows in Kagent, where multiple MCP tools, remote agents, and multi-step reasoning are involved. Recommendations are prioritized by implementation effort: **config-only changes first**, code changes only when necessary.

---

## Table of Contents

1. [Parameter Reference](#1-parameter-reference)
2. [Recommended Values for Complex Agent Workflows](#2-recommended-values-for-complex-agent-workflows)
3. [Priority 1: Config-Only Changes (ModelConfig CRD)](#3-priority-1-config-only-changes-modelconfig-crd)
4. [Priority 2: Agent CRD Config Changes](#4-priority-2-agent-crd-config-changes)
5. [Priority 3: Helm Values Tuning](#5-priority-3-helm-values-tuning)
6. [Priority 4: Code Changes (If Required)](#6-priority-4-code-changes-if-required)
7. [Provider-Specific Recommendations](#7-provider-specific-recommendations)
8. [Anti-Patterns to Avoid](#8-anti-patterns-to-avoid)
9. [Example Configurations](#9-example-configurations)

---

## 1. Parameter Reference

### What Each Parameter Controls

| Parameter | What It Does | Impact on Agent Workflows |
|-----------|-------------|---------------------------|
| **temperature** | Controls randomness (0.0 = deterministic, 2.0 = maximum randomness) | Lower = more consistent tool selection and argument formatting |
| **top_p** | Nucleus sampling - considers tokens whose cumulative probability reaches this threshold | Complements temperature; controls diversity of token choices |
| **top_k** | Only considers the top K most probable tokens (Anthropic/Vertex AI only) | Limits vocabulary; useful for constraining outputs |
| **max_tokens** | Maximum tokens in the response | Prevents runaway costs; must be large enough for tool call JSON |
| **frequency_penalty** | Penalizes tokens based on how often they appear in the response so far (-2.0 to 2.0) | Reduces repetitive tool calls or redundant reasoning |
| **presence_penalty** | Penalizes tokens that have appeared at all in the response so far (-2.0 to 2.0) | Encourages exploration of different tools/approaches |
| **seed** | For reproducible outputs (OpenAI only, best-effort) | Useful for debugging specific agent behaviors |
| **reasoning_effort** | Controls thinking depth before responding (OpenAI o-series models: minimal/low/medium/high) | Directly affects tool selection accuracy vs latency/cost |
| **stream** | Whether to stream LLM responses token-by-token | Affects perceived latency and A2A protocol behavior |
| **n** | Number of completions to generate per request | **Always 1 for agents** - multiple completions break tool-calling flow |

### Where Each Parameter Is Configurable

| Parameter | ModelConfig CRD | Helm Values | Code Change Required |
|-----------|:-:|:-:|:-:|
| temperature | Yes | No (but could add) | No |
| top_p | Yes | No | No |
| top_k | Yes (Anthropic/Vertex) | No | No |
| max_tokens | Yes | No | No |
| frequency_penalty | Yes (OpenAI) | No | No |
| presence_penalty | Yes (OpenAI) | No | No |
| seed | Yes (OpenAI) | No | No |
| reasoning_effort | Yes (OpenAI) | No | No |
| stream | Agent CRD | No | No |
| n | Yes (OpenAI) | No | No |

---

## 2. Recommended Values for Complex Agent Workflows

### The Core Problem

In complex Kagent workflows with multiple MCP tools:
- The LLM must **reliably select the correct tool** from many available options
- Tool call arguments must be **precisely formatted JSON**
- Multi-step chains must maintain **coherent state** across turns
- Token usage should be **efficient** to control costs
- The agent must not **hallucinate tool names** or **repeat failed calls**

### Recommended Baseline Values

| Parameter | Recommended Value | Rationale |
|-----------|:-:|-----------|
| **temperature** | `0.1` - `0.3` | Low enough for consistent tool selection, high enough to avoid degenerate loops |
| **top_p** | `0.9` - `0.95` | Slightly constrained; eliminates very unlikely tokens while keeping sufficient diversity |
| **top_k** | `40` (Anthropic only) | Good balance for tool-heavy workflows |
| **max_tokens** | `4096` - `8192` | Must accommodate complex tool call JSON + reasoning; 4096 minimum for multi-tool responses |
| **frequency_penalty** | `0.1` - `0.3` | Gently discourages repetitive tool invocations without breaking structured output |
| **presence_penalty** | `0.0` - `0.1` | Keep low; higher values can corrupt tool argument formatting |
| **seed** | (unset for production, set for debugging) | Use only when reproducing specific issues |
| **reasoning_effort** | `medium` - `high` | Higher effort improves tool selection accuracy; `medium` balances cost vs quality |
| **stream** | `true` | Recommended for better perceived latency in UI; the framework handles streaming correctly |
| **n** | `1` (or unset) | **Never set > 1 for tool-calling agents** |

---

## 3. Priority 1: Config-Only Changes (ModelConfig CRD)

**Effort: Lowest** - Just apply YAML to your cluster.

### 3.1 OpenAI ModelConfig for Complex Agents

```yaml
apiVersion: kagent.dev/v1alpha2
kind: ModelConfig
metadata:
  name: agent-workflow-openai
  namespace: kagent
spec:
  model: gpt-4.1  # or gpt-4.1-mini for cost optimization
  provider: OpenAI
  apiKeySecret: kagent-openai
  apiKeySecretKey: OPENAI_API_KEY
  openAI:
    temperature: "0.2"
    maxTokens: 8192
    topP: "0.95"
    frequencyPenalty: "0.1"
    presencePenalty: "0.0"
    # For o-series models (o3, o4-mini), uncomment:
    # reasoningEffort: "medium"
```

**Why these values:**
- `temperature: 0.2` - Consistent tool selection without being fully deterministic (which can cause loops)
- `maxTokens: 8192` - Allows room for complex multi-tool responses with structured JSON
- `topP: 0.95` - Eliminates bottom 5% unlikely tokens
- `frequencyPenalty: 0.1` - Gently discourages repeating the same tool call pattern
- `presencePenalty: 0.0` - No presence penalty keeps tool argument formatting clean

### 3.2 Anthropic ModelConfig for Complex Agents

```yaml
apiVersion: kagent.dev/v1alpha2
kind: ModelConfig
metadata:
  name: agent-workflow-anthropic
  namespace: kagent
spec:
  model: claude-sonnet-4-20250514
  provider: Anthropic
  apiKeySecret: kagent-anthropic
  apiKeySecretKey: ANTHROPIC_API_KEY
  anthropic:
    temperature: "0.2"
    maxTokens: 8192
    topP: "0.9"
    topK: 40
```

**Why these values:**
- `topK: 40` - Limits token choices; Anthropic models respond well to this for structured output
- `topP: 0.9` - Slightly more constrained than OpenAI since Anthropic models are already quite focused

### 3.3 Ollama ModelConfig for Complex Agents

```yaml
apiVersion: kagent.dev/v1alpha2
kind: ModelConfig
metadata:
  name: agent-workflow-ollama
  namespace: kagent
spec:
  model: llama3.2
  provider: Ollama
  ollama:
    host: "host.docker.internal:11434"
    options:
      num_ctx: "65536"        # Large context for multi-tool conversations
      temperature: "0.2"
      top_p: "0.9"
      top_k: "40"
      repeat_penalty: "1.1"  # Ollama equivalent of frequency_penalty
      num_predict: "8192"    # Max tokens equivalent
```

**Why these values:**
- `num_ctx: 65536` - Critical for complex workflows; tool definitions consume significant context
- `repeat_penalty: 1.1` - Ollama's mechanism to reduce repetitive tool calls

### 3.4 Azure OpenAI / GeminiVertexAI / Bedrock

For Azure OpenAI, the same temperature/topP/maxTokens recommendations apply via the `azureOpenAI` config block.

For GeminiVertexAI, use the `geminiVertexAI` config block with temperature/topP/topK.

Bedrock configuration is limited to region; model parameters are controlled at the Bedrock model deployment level.

---

## 4. Priority 2: Agent CRD Config Changes

**Effort: Low** - Agent-level YAML changes.

### 4.1 Enable Streaming

```yaml
apiVersion: kagent.dev/v1alpha2
kind: Agent
metadata:
  name: complex-troubleshooter
  namespace: kagent
spec:
  modelConfigRef: agent-workflow-openai
  description: "Complex troubleshooting agent with multiple tools"
  declarative:
    stream: true  # Enable streaming for better UX
    systemMessage: |
      You are a Kubernetes troubleshooting agent.
      When using tools, follow these rules:
      1. Call one tool at a time and wait for results
      2. Validate tool output before proceeding
      3. If a tool call fails, try an alternative approach
      ...
```

**Why streaming matters:**
- Token-by-token delivery reduces perceived latency
- The Go HTTP server correctly proxies A2A streaming to the UI
- No accuracy impact - the LLM generates the same output whether streamed or not

### 4.2 System Message Engineering (Config, Not Code)

The system message in the Agent CRD is one of the most impactful levers. For complex tool workflows:

```yaml
systemMessage: |
  You are an expert agent with access to the following tool categories:
  - Kubernetes tools (via Helm MCP server)
  - Monitoring tools (via Prometheus MCP server)
  - Documentation tools (via search MCP server)

  ## Tool Usage Guidelines
  - Always verify your understanding before calling a tool
  - Use the most specific tool available for the task
  - If a tool returns an error, analyze the error before retrying
  - Never call the same tool with the same arguments more than twice
  - Prefer tools that return structured data over raw text

  ## Response Format
  - Be concise in your reasoning
  - Show tool results to the user, don't just summarize
  - If multiple tools are needed, explain your plan first
```

This is **configuration only** and has enormous impact on token efficiency and tool selection accuracy.

---

## 5. Priority 3: Helm Values Tuning

**Effort: Medium** - Changes to `helm/kagent/values.yaml` or override files.

### 5.1 Default Provider Configuration

The Helm chart defines default providers. Update these to include parameter tuning:

```yaml
# In your Helm values override file
providers:
  default: openAI
  openAI:
    provider: OpenAI
    model: "gpt-4.1"  # Upgrade from gpt-4.1-mini for complex workflows
    apiKeySecretRef: kagent-openai
    apiKeySecretKey: OPENAI_API_KEY
    # Note: Provider-specific params (temperature, etc.) are NOT
    # currently exposed in Helm values. Use ModelConfig CRD instead.
```

### 5.2 What Helm Values Don't Cover (Yet)

The current Helm chart (`helm/kagent/values.yaml`) does **not** expose provider-specific parameters like temperature, maxTokens, etc. These must be set via the ModelConfig CRD directly. This is actually fine for production - you typically want different ModelConfig resources for different agents with different tuning profiles.

---

## 6. Priority 4: Code Changes (If Required)

**Effort: Highest** - Only pursue if config-only changes are insufficient.

### 6.1 Gap: Anthropic Parameters Not Passed Through ADK

**Issue:** The Go translator (`adk_api_translator.go` lines 778-805) does NOT pass Anthropic's temperature, topP, topK, or maxTokens to the Python ADK config. The CRD accepts these values, but they are lost during translation.

**Current code path:**
```
CRD AnthropicConfig (has temperature, topP, topK, maxTokens)
    → Go translator (only passes model, baseUrl, headers)
        → Python ADK Anthropic type (only has base_url)
            → LiteLLM (receives no parameters)
```

**Recommended fix (if needed):**

1. Update Python ADK `Anthropic` type in `python/packages/kagent-adk/src/kagent/adk/types.py`:
```python
class Anthropic(BaseLLM):
    base_url: str | None = None
    temperature: float | None = None
    top_p: float | None = None
    top_k: int | None = None
    max_tokens: int | None = None
    type: Literal["anthropic"]
```

2. Update Go ADK Anthropic type in `go/pkg/adk/types.go` to include these fields.

3. Update the translator in `go/internal/controller/translator/agent/adk_api_translator.go` to pass through the values.

**Priority:** Medium-high if using Anthropic as provider. The parameters are accepted by the CRD but silently dropped.

### 6.2 Gap: Azure OpenAI Parameters Not Passed Through

Similar to Anthropic - the `AzureOpenAI` translator passes configuration via environment variables but does NOT forward temperature, topP, or maxTokens from the CRD to the Python ADK. These parameters exist in the CRD (`AzureOpenAIConfig`) but are lost.

**Priority:** Medium if using Azure OpenAI.

### 6.3 Gap: GeminiVertexAI Parameters Not Passed Through

The GeminiVertexAI CRD defines temperature, topP, topK, maxOutputTokens, but the translator does not forward them.

**Priority:** Medium if using Vertex AI.

### 6.4 Potential Enhancement: Retry/Backoff Configuration

Currently there is no CRD-level configuration for LLM call retry behavior (timeout, retry count, backoff). For complex agent workflows with unreliable networks or rate limits:

- The `timeout` field exists for OpenAI but not other providers
- No retry count or backoff configuration exists

**Priority:** Low - LiteLLM has built-in retry logic. Only pursue if you observe timeout issues.

---

## 7. Provider-Specific Recommendations

### OpenAI (Most Complete Parameter Support)

| Use Case | Model | temperature | maxTokens | topP | frequencyPenalty | reasoningEffort |
|----------|-------|:-:|:-:|:-:|:-:|:-:|
| Complex multi-tool troubleshooting | gpt-4.1 | 0.2 | 8192 | 0.95 | 0.1 | - |
| Cost-efficient simple workflows | gpt-4.1-mini | 0.3 | 4096 | 0.95 | 0.0 | - |
| Deep reasoning + tool use | o4-mini | - | 16384 | - | - | medium |
| Maximum accuracy | o3 | - | 16384 | - | - | high |

**Note on o-series models:** Temperature and topP are not applicable; use `reasoningEffort` instead.

### Anthropic

| Use Case | Model | temperature | maxTokens | topP | topK |
|----------|-------|:-:|:-:|:-:|:-:|
| Complex multi-tool workflows | claude-sonnet-4 | 0.2 | 8192 | 0.9 | 40 |
| Cost-efficient workflows | claude-haiku-4-5 | 0.3 | 4096 | 0.9 | 40 |
| Maximum accuracy | claude-opus-4 | 0.1 | 8192 | 0.9 | 30 |

**Note:** Requires code change (Priority 4, section 6.1) to actually pass these parameters through.

### Ollama (Local Models)

| Use Case | Model | num_ctx | temperature | top_p | top_k | repeat_penalty |
|----------|-------|:-:|:-:|:-:|:-:|:-:|
| Complex workflows | llama3.2:70b | 65536 | 0.2 | 0.9 | 40 | 1.1 |
| Resource-constrained | llama3.2:8b | 32768 | 0.1 | 0.85 | 30 | 1.15 |

---

## 8. Anti-Patterns to Avoid

### DO NOT: Set temperature to 0.0

A temperature of exactly 0 can cause deterministic loops where the agent repeatedly calls the same failing tool. Use `0.1` minimum to inject just enough variance.

### DO NOT: Set max_tokens too low

Tool-calling responses include structured JSON for function calls. A response with 3 tool calls can easily be 500-1000 tokens. Setting `maxTokens: 512` will cause truncated JSON and tool call failures. Minimum `4096` for any multi-tool workflow.

### DO NOT: Use n > 1

Setting `n > 1` generates multiple completions. The agent framework expects exactly one completion with tool calls. Multiple completions will break the tool execution loop.

### DO NOT: Set high presence_penalty with tools

`presencePenalty > 0.3` can cause the LLM to avoid re-using JSON keys it has already used in the response (like `"name"`, `"arguments"`), leading to malformed tool call JSON.

### DO NOT: Set high frequency_penalty with structured output

`frequencyPenalty > 0.5` penalizes repeated tokens. JSON tool calls inherently repeat tokens like `{`, `"`, `:`. High penalty degrades JSON formatting quality.

### DO NOT: Ignore context window limits

Complex agent workflows with many MCP tools consume significant context for tool definitions alone. For example:
- Each MCP tool definition: ~100-300 tokens
- 20 tools: ~2,000-6,000 tokens just for definitions
- Multi-turn conversation with tool results: 10,000+ tokens per turn

Ensure your model's context window and `maxTokens` accommodate this.

---

## 9. Example Configurations

### Production-Ready: Kubernetes Troubleshooting Agent

```yaml
# ModelConfig
apiVersion: kagent.dev/v1alpha2
kind: ModelConfig
metadata:
  name: k8s-troubleshooter-model
  namespace: kagent
spec:
  model: gpt-4.1
  provider: OpenAI
  apiKeySecret: kagent-openai
  apiKeySecretKey: OPENAI_API_KEY
  openAI:
    temperature: "0.2"
    maxTokens: 8192
    topP: "0.95"
    frequencyPenalty: "0.1"
    presencePenalty: "0.0"
---
# Agent
apiVersion: kagent.dev/v1alpha2
kind: Agent
metadata:
  name: k8s-troubleshooter
  namespace: kagent
spec:
  modelConfigRef: k8s-troubleshooter-model
  description: "Kubernetes cluster troubleshooting agent"
  declarative:
    stream: true
    systemMessage: |
      You are a Kubernetes troubleshooting expert. You have access to tools
      for querying cluster state, reading logs, checking metrics, and
      managing Helm releases.

      ## Guidelines
      - Start by gathering context: check pod status, events, and logs
      - Use Prometheus tools for metric analysis only when symptoms suggest resource issues
      - Present findings with evidence from tool outputs
      - Suggest actionable remediation steps
    tools:
      - mcpServerRef:
          name: kubectl-mcp-server
      - mcpServerRef:
          name: prometheus-mcp-server
      - mcpServerRef:
          name: helm-mcp-server
```

### Cost-Optimized: Simple Query Agent

```yaml
apiVersion: kagent.dev/v1alpha2
kind: ModelConfig
metadata:
  name: simple-query-model
  namespace: kagent
spec:
  model: gpt-4.1-mini
  provider: OpenAI
  apiKeySecret: kagent-openai
  apiKeySecretKey: OPENAI_API_KEY
  openAI:
    temperature: "0.3"
    maxTokens: 4096
    topP: "0.95"
    frequencyPenalty: "0.0"
    presencePenalty: "0.0"
```

### Debugging Configuration (Reproducible Outputs)

```yaml
apiVersion: kagent.dev/v1alpha2
kind: ModelConfig
metadata:
  name: debug-model
  namespace: kagent
spec:
  model: gpt-4.1
  provider: OpenAI
  apiKeySecret: kagent-openai
  apiKeySecretKey: OPENAI_API_KEY
  openAI:
    temperature: "0.0"
    maxTokens: 8192
    topP: "1.0"
    frequencyPenalty: "0.0"
    presencePenalty: "0.0"
    seed: 42  # For reproducibility (best-effort)
```

---

## Summary: Priority Order for Tuning

| Priority | What to Change | Where | Effort | Impact |
|:-:|---|---|:-:|:-:|
| **1** | temperature, maxTokens, topP in ModelConfig CRD | `kubectl apply -f modelconfig.yaml` | Lowest | High |
| **2** | System message engineering in Agent CRD | `kubectl apply -f agent.yaml` | Low | Very High |
| **3** | Enable streaming in Agent CRD | Agent spec `.declarative.stream: true` | Low | Medium (UX) |
| **4** | frequencyPenalty tuning (OpenAI) | ModelConfig CRD | Lowest | Medium |
| **5** | reasoningEffort for o-series models (OpenAI) | ModelConfig CRD | Lowest | High |
| **6** | Ollama num_ctx / repeat_penalty | ModelConfig CRD `ollama.options` | Lowest | High (Ollama) |
| **7** | Fix Anthropic parameter passthrough gap | Code change (3 files) | Medium | High (Anthropic users) |
| **8** | Fix Azure/Vertex parameter passthrough gap | Code change (3 files) | Medium | Medium |

**Key takeaway:** 80% of the tuning impact comes from Priority 1-3 (all config-only, zero code changes). The most impactful single change is a well-crafted system message combined with `temperature: 0.2` and adequate `maxTokens`.

---

**Related documents:**
- [Agent Execution Flow and Context Management](agent-execution-flow-and-context-management.md)
- [Controller Reconciliation](controller-reconciliation.md)

**Last Updated:** 2026-02-23
