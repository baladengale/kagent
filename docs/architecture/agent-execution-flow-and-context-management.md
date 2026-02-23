# Agent Execution Flow, MCP Tool Invocation, and Context Management

This document provides a deep analysis of how kagent agents execute, invoke MCP tools, interact with LLMs, and manage conversation context. It covers the complete request lifecycle, the tool-calling loop, context window strategies, and extension points for controlling agent behavior.

**Audience:** Developers exploring kagent internals, contributors building custom agents, and anyone investigating context optimization strategies.

---

## Table of Contents

- [1. Architecture Overview](#1-architecture-overview)
- [2. Request Lifecycle: End-to-End Flow](#2-request-lifecycle-end-to-end-flow)
- [3. The Agent Loop: LLM and Tool Invocation](#3-the-agent-loop-llm-and-tool-invocation)
  - [3.1 When Does the LLM Get Called?](#31-when-does-the-llm-get-called)
  - [3.2 How Tool Calls Are Dispatched](#32-how-tool-calls-are-dispatched)
  - [3.3 Parallel vs Sequential Tool Execution](#33-parallel-vs-sequential-tool-execution)
  - [3.4 The Complete Loop Cycle](#34-the-complete-loop-cycle)
- [4. MCP Tool Integration](#4-mcp-tool-integration)
  - [4.1 Discovery: Go Controller Side](#41-discovery-go-controller-side)
  - [4.2 Runtime: Python ADK Side](#42-runtime-python-adk-side)
  - [4.3 Tool Filtering](#43-tool-filtering)
  - [4.4 Header Propagation](#44-header-propagation)
- [5. Conversation History and Session Management](#5-conversation-history-and-session-management)
  - [5.1 Session Storage](#51-session-storage)
  - [5.2 How History Is Built for Each LLM Call](#52-how-history-is-built-for-each-llm-call)
  - [5.3 Event Persistence Rules](#53-event-persistence-rules)
- [6. Context Window Management](#6-context-window-management)
  - [6.1 Current State in Kagent](#61-current-state-in-kagent)
  - [6.2 Google ADK Events Compaction](#62-google-adk-events-compaction)
  - [6.3 Compaction Configuration](#63-compaction-configuration)
  - [6.4 Custom Summarizer](#64-custom-summarizer)
  - [6.5 Limitations of Current Compaction](#65-limitations-of-current-compaction)
- [7. Controlling Agent Behavior: Extension Points](#7-controlling-agent-behavior-extension-points)
  - [7.1 Callback Hooks](#71-callback-hooks)
  - [7.2 Tool Output Control](#72-tool-output-control)
  - [7.3 Flow Control Actions](#73-flow-control-actions)
  - [7.4 Model Parameters](#74-model-parameters)
- [8. Opportunities: Context Optimization Techniques](#8-opportunities-context-optimization-techniques)
  - [8.1 Observation Masking](#81-observation-masking)
  - [8.2 LLM-Based Summarization](#82-llm-based-summarization)
  - [8.3 Token-Based Sliding Window](#83-token-based-sliding-window)
  - [8.4 Tool Output Truncation](#84-tool-output-truncation)
  - [8.5 Selective History Pruning](#85-selective-history-pruning)
- [9. Key Source Files Reference](#9-key-source-files-reference)
- [10. Diagrams](#10-diagrams)

---

## 1. Architecture Overview

Kagent's agent execution spans three layers:

```
┌──────────────────────────────────────────────────────────────────┐
│                         UI (Next.js)                             │
│    User sends message ──► A2A protocol ──► Go HTTP Server        │
└────────────────────────────────┬─────────────────────────────────┘
                                 │
                                 ▼
┌──────────────────────────────────────────────────────────────────┐
│                    Go HTTP Server + Database                      │
│  Routes /api/chat/* ──► Proxies to Python ADK runtime            │
│  Stores: sessions, events, agent configs, tool metadata          │
└────────────────────────────────┬─────────────────────────────────┘
                                 │
                                 ▼
┌──────────────────────────────────────────────────────────────────┐
│                   Python Agent Runtime (ADK)                      │
│  ┌─────────────┐   ┌──────────────┐   ┌──────────────────────┐  │
│  │ A2aAgent    │──▶│ Google ADK   │──▶│ LLM Provider         │  │
│  │ Executor    │   │ Runner       │   │ (OpenAI/Anthropic/   │  │
│  │ (kagent)    │   │ (core loop)  │   │  Gemini/Ollama/etc.) │  │
│  └─────────────┘   └──────┬───────┘   └──────────────────────┘  │
│                            │                                      │
│                            ▼                                      │
│                    ┌──────────────┐                                │
│                    │ MCP Servers  │  (external tool servers)      │
│                    │ (via HTTP/   │                                │
│                    │  SSE)        │                                │
│                    └──────────────┘                                │
└──────────────────────────────────────────────────────────────────┘
```

**Key principle:** Kagent wraps Google's ADK (Agent Development Kit). The core agent loop -- deciding when to call the LLM, when to execute tools, and when to finish -- is delegated to Google ADK's `Runner` and `BaseLlmFlow`. Kagent adds Kubernetes-native lifecycle management, A2A protocol bridging, session persistence, MCP toolset wrappers, and multi-provider LLM support.

---

## 2. Request Lifecycle: End-to-End Flow

Here is the complete flow from a user message to the final response:

```
User Message (UI)
    │
    ▼
[1] Go HTTP Server receives A2A request
    │
    ▼
[2] Go server forwards to Python ADK runtime (A2A protocol)
    │
    ▼
[3] A2aAgentExecutor.execute() receives the request
    │   - Converts A2A message to ADK format (request_converter.py)
    │   - Resolves a FRESH Runner instance per request
    │   - Prepares or retrieves Session (with full event history)
    │   - Sets request headers into session state
    │
    ▼
[4] Runner.run_async() starts the agent loop
    │   - Loads full conversation history from session
    │   - Enters the LLM flow loop (SingleFlow or AutoFlow)
    │
    ▼
[5] ┌─── AGENT LOOP (repeats until done) ───────────────────┐
    │                                                         │
    │  [5a] Preprocess: Build LLM request                     │
    │       - Gather conversation history from session events │
    │       - Resolve tool definitions from MCP toolsets      │
    │       - Apply request processors                        │
    │                                                         │
    │  [5b] Call LLM with: system prompt + history + tools    │
    │       - Run before_model_callback (can short-circuit)   │
    │       - Send request to LLM provider                    │
    │       - Run after_model_callback (can modify response)  │
    │                                                         │
    │  [5c] Process LLM response:                             │
    │       IF response contains function_call(s):            │
    │         - Run before_tool_callback per tool              │
    │         - Execute tool(s) via MCP / local function      │
    │         - Run after_tool_callback per tool               │
    │         - Accumulate function_response events            │
    │         - LOOP BACK to [5a] (LLM sees tool results)     │
    │       IF response is text (no tool calls):              │
    │         - This is the FINAL response                    │
    │         - EXIT LOOP                                     │
    │                                                         │
    └─────────────────────────────────────────────────────────┘
    │
    ▼
[6] Events converted to A2A format (event_converter.py)
    │   - Streamed to frontend via event queue
    │   - Non-partial events persisted to session
    │
    ▼
[7] Final TaskStatusUpdateEvent (completed/failed) sent
    │   - Runner closed (MCP connections cleaned up)
    │
    ▼
User sees response in UI
```

**Source:** `python/packages/kagent-adk/src/kagent/adk/_agent_executor.py` (lines 143-375)

---

## 3. The Agent Loop: LLM and Tool Invocation

### 3.1 When Does the LLM Get Called?

The LLM is called in these situations:

1. **Initial call:** When the user sends a message, the LLM receives the full conversation history + system prompt + available tool definitions.

2. **After every tool execution:** When the LLM requests tool(s) and they complete, the LLM is called AGAIN with the tool results appended to the conversation. This is the critical insight: **the LLM is called after EVERY round of tool execution, not just at the end.**

3. **Agent transfer:** When a sub-agent completes, the parent agent's LLM may be called to process the sub-agent's response.

The flow is implemented in `BaseLlmFlow.run_async()`:

```
while True:
    event = _run_one_step_async()   # One LLM call + optional tool execution
    yield event
    if event.is_final_response:     # LLM returned text, not tool calls
        break
```

### 3.2 How Tool Calls Are Dispatched

When the LLM response contains `function_call` parts:

1. The `BaseLlmFlow._postprocess_async()` detects function calls in the LLM response
2. It calls `functions.handle_function_calls_async()`
3. Each function call is dispatched to the appropriate tool (MCP toolset, local function, or agent tool)
4. The tool results are collected as `function_response` events
5. These responses are appended to the conversation history
6. The loop continues (LLM is called again with the updated history)

### 3.3 Parallel vs Sequential Tool Execution

**Tool calls within a single LLM turn execute in PARALLEL by default:**

```python
# From google/adk/flows/llm_flows/functions.py
tasks = [
    asyncio.create_task(
        _execute_single_function_call_async(invocation_context, fc, tools_dict, ...)
    )
    for fc in filtered_function_calls
]
function_response_events = await asyncio.gather(*tasks)
```

When an LLM requests multiple tools in one response (e.g., "call tool_A AND tool_B"), they execute concurrently via `asyncio.gather()`. The results are merged into a single event before the next LLM call.

**Tool calls across LLM turns are sequential:** If the LLM first calls tool_A, sees the result, then decides to call tool_B, these are separate turns and execute sequentially.

**Requirements for true parallel execution:**
- Tool functions must be `async`
- Tools must use non-blocking I/O (`await asyncio.sleep()` not `time.sleep()`)
- MCP tools naturally support this since they use async HTTP/SSE transports

### 3.4 The Complete Loop Cycle

```
Turn 1: User message ──► LLM ──► "I need to call get_pods and get_services"
                                   (two function_calls in one response)
                                        │
                                        ▼
         Execute get_pods() ──────┐
         Execute get_services() ──┤  (PARALLEL via asyncio.gather)
                                  │
                                  ▼
         Merge results into single function_response event
                                  │
                                  ▼
Turn 2: [history + tool results] ──► LLM ──► "Now I need to call describe_pod"
                                              (one function_call)
                                                   │
                                                   ▼
         Execute describe_pod()
                                                   │
                                                   ▼
Turn 3: [history + tool result] ──► LLM ──► "Here is the analysis..."
                                             (final text response, no tool calls)
                                             EXIT LOOP
```

**Key insight:** Each "turn" is a full LLM API call. The conversation history grows with every turn. With N sequential tool-calling rounds, you make N+1 LLM calls (initial + one per tool round).

---

## 4. MCP Tool Integration

### 4.1 Discovery: Go Controller Side

MCP tools are discovered at the Kubernetes controller level:

1. **CRD creation:** User creates `RemoteMCPServer` or `MCPServer` custom resources
2. **Controller reconciliation:** The `MCPServerToolController` and `RemoteMCPServerController` reconcile these resources every 60 seconds
3. **Tool listing:** The controller connects to MCP servers using SSE or Streamable HTTP transport and calls `session.ListTools()`
4. **Storage:** Discovered tools are stored in the database (`tool_server_tools` table) with name, description, and server reference
5. **API exposure:** Tools are available via REST API at `/api/tools` and `/api/toolservers`

```go
// From go/internal/controller/reconciler/reconciler.go
session, err := client.Connect(ctx, transport, nil)
result, err := session.ListTools(ctx, &mcp.ListToolsParams{})
```

### 4.2 Runtime: Python ADK Side

When a request arrives, tools are used at runtime:

1. **Agent configuration:** The `AgentConfig.to_agent()` method creates `KAgentMcpToolset` instances for each configured MCP server
2. **Runner creation:** A fresh `Runner` is created per request (MCP connections are NOT shared between requests)
3. **Tool resolution:** When the LLM flow preprocesses a request, it calls `toolset.get_tools()` which connects to the MCP server and fetches available tools
4. **Tool invocation:** When the LLM calls a tool, the `McpToolset` proxies the call to the MCP server via the established transport
5. **Cleanup:** After the request completes, `runner.close()` cleans up all MCP connections

```python
# From python/packages/kagent-adk/src/kagent/adk/types.py
tools.append(
    KAgentMcpToolset(
        connection_params=http_tool.params,
        tool_filter=http_tool.tools,        # Optional: limit which tools are exposed
        header_provider=tool_header_provider, # Optional: inject auth headers
    )
)
```

### 4.3 Tool Filtering

Filtering happens at multiple levels:

| Level | Mechanism | Description |
|-------|-----------|-------------|
| **CRD config** | `Agent.spec.tools[].toolNames` | Agent specifies which tools from an MCP server to use |
| **Runtime filter** | `KAgentMcpToolset(tool_filter=[...])` | Only matching tools are exposed to the LLM |
| **LLM selection** | Model decides | The LLM picks which available tools to call based on the task |

If `tool_filter` is empty, ALL tools from the MCP server are available.

### 4.4 Header Propagation

Headers flow through a layered system:

```
RemoteMCPServer CRD ──► base headers (from Secrets)
         │
         ▼
Agent Tool config ──► allowed_headers (forwarded from A2A request)
         │
         ▼
STS integration ──► auth tokens (take precedence over allowed headers)
         │
         ▼
Final headers ──► sent with every MCP protocol request
```

**Source:** `python/packages/kagent-adk/src/kagent/adk/types.py` (lines 34-84)

---

## 5. Conversation History and Session Management

### 5.1 Session Storage

Kagent implements a custom `SessionService` (`_session_service.py`) that persists sessions to the Go backend:

- **Sessions** are stored with: `id`, `app_name`, `user_id`, `state`, `name`
- **Events** are stored per session with: `id`, `data` (serialized Event JSON)
- Sessions are retrieved with ALL events loaded in chronological order

```python
# From _session_service.py
response = await self.client.get(
    f"/api/sessions/{session_id}",
    params={"user_id": user_id, "order": "asc"},
)
```

### 5.2 How History Is Built for Each LLM Call

Every time the LLM is called, the Google ADK Runner constructs the conversation from the session's events:

1. **Session retrieval:** All stored events are loaded from the backend
2. **Event replay:** Events are replayed into the session object via `super().append_event()`
3. **Content construction:** The ADK's `BaseLlmFlow` request processor (`_contents_llm_request_processor`) builds the `contents` list from events:
   - `user` role messages from user events
   - `model` role messages from LLM response events (including `function_call` parts)
   - `user` role messages with `function_response` parts from tool result events
4. **Full payload:** System instruction + conversation contents + tool definitions are sent to the LLM

**Critical implication:** The ENTIRE conversation history is sent with every LLM call. There is no built-in truncation at the kagent layer. This means:
- Long conversations consume more tokens per call
- Eventually, the conversation may exceed the model's context window
- Cost grows linearly with conversation length

### 5.3 Event Persistence Rules

| Event Type | Persisted? | Notes |
|-----------|-----------|-------|
| `partial=True` (streaming chunks) | No | Sent to frontend only, not saved |
| `partial=False` (complete events) | Yes | Saved to backend via `append_event` |
| System events (header updates) | Yes | Saved for session state management |
| Compaction events | Yes | Replace summarized events |

---

## 6. Context Window Management

### 6.1 Current State in Kagent

**Kagent does NOT currently implement any context window management.** The full conversation history is passed to every LLM call. Context limits are entirely delegated to the underlying LLM provider's SDK, which may truncate or error out.

This is a significant area for improvement, especially for agents handling:
- Long-running tasks with many tool calls
- Conversations with large tool outputs (e.g., kubectl output, log dumps)
- Multi-turn debugging sessions

### 6.2 Google ADK Events Compaction

Google ADK provides a built-in **Events Compaction** mechanism that kagent could leverage. It uses a sliding window approach:

```
Events:  [E1] [E2] [E3] [E4] [E5] [E6] [E7] [E8] [E9]
                         ▲                         ▲
                    compaction_interval=3      compaction_interval=3
                         │                         │
                         ▼                         ▼
After:   [Summary_1]    [E3] [E4] [E5] [Summary_2] [E8] [E9]
          (E1-E2)        overlap    (E4-E7)         overlap
```

How it works:
1. When the number of completed invocations reaches `compaction_interval`, compaction triggers
2. An LLM (the agent's model or a custom summarizer model) summarizes the older events
3. The summary is written as a new event with a "compaction" action
4. Subsequent LLM calls use the summary instead of the raw events

### 6.3 Compaction Configuration

To enable compaction in kagent, you would configure it on the `App` object:

```python
from google.adk.apps.app import App, EventsCompactionConfig

app = App(
    name='my-agent',
    root_agent=root_agent,
    events_compaction_config=EventsCompactionConfig(
        compaction_interval=3,   # Compact every 3 invocations
        overlap_size=1,          # Keep last 1 invocation from previous window
    ),
)
```

**Note:** Kagent currently creates `Runner` instances directly without using `App`. To leverage compaction, kagent would need to either:
- Start using `App` objects (recommended)
- Implement custom compaction in the `SessionService`

### 6.4 Custom Summarizer

You can use a different (cheaper/faster) model for summarization:

```python
from google.adk.apps.llm_event_summarizer import LlmEventSummarizer
from google.adk.models import Gemini

summarizer = LlmEventSummarizer(llm=Gemini(model="gemini-2.5-flash"))

events_compaction_config=EventsCompactionConfig(
    compaction_interval=3,
    overlap_size=1,
    summarizer=summarizer,
)
```

### 6.5 Limitations of Current Compaction

1. **Turn-based only:** Compaction triggers every N invocations, NOT based on token count
2. **No token threshold:** Cannot say "compact when context exceeds 50% of max tokens"
3. **No selective compaction:** Cannot choose to compact only tool outputs while keeping user messages intact
4. **No observation masking:** Cannot selectively hide or truncate specific tool results
5. **SequentialAgent incompatibility:** Requires a model for summarization; fails with agents that don't have a direct LLM

These limitations are recognized by the community (see [google/adk-python#4146](https://github.com/google/adk-python/issues/4146)).

---

## 7. Controlling Agent Behavior: Extension Points

### 7.1 Callback Hooks

Google ADK provides six callback hooks on the `Agent` class:

```
before_agent ──► before_model ──► [LLM Call] ──► after_model
                                                      │
                                          ┌───────────┴───────────┐
                                          │ (if tool calls)       │ (if text response)
                                          ▼                       ▼
                                    before_tool              after_agent
                                         │
                                    [Tool Execution]
                                         │
                                    after_tool
                                         │
                                    (loop back to before_model)
```

| Callback | When It Fires | Can Short-Circuit? | Use Cases |
|----------|--------------|-------------------|-----------|
| `before_agent_callback` | Before agent starts | Yes (return content) | Auth checks, rate limiting |
| `before_model_callback` | Before each LLM call | Yes (return LlmResponse) | Caching, request filtering |
| `after_model_callback` | After each LLM response | Yes (modify/replace response) | Response filtering, logging |
| `before_tool_callback` | Before each tool execution | Yes (return dict to skip tool) | Policy enforcement, argument validation |
| `after_tool_callback` | After each tool execution | Yes (modify result) | **Output truncation**, result caching |
| `on_model_error_callback` | When LLM call fails | Yes (provide fallback) | Error handling, retry logic |

**Key insight for context management:** `after_tool_callback` is the primary extension point for **observation masking** and **tool output truncation**. You can intercept tool results and truncate/summarize them before they enter the conversation history.

### 7.2 Tool Output Control

The `EventActions` class provides per-event flow control:

```python
# skip_summarization: Tell the LLM NOT to summarize this tool's output
# Useful when tool output is already user-ready
event.actions.skip_summarization = True

# escalate: Signal that current agent cannot handle the request
event.actions.escalate = True

# transfer_to_agent: Hand off to another agent
event.actions.transfer_to_agent = "specialist-agent"
```

### 7.3 Flow Control Actions

| Action | Effect | Use Case |
|--------|--------|----------|
| `skip_summarization` | LLM won't summarize tool output | Tool already produces final answer |
| `escalate` | Return to parent agent | Current agent lacks capability |
| `transfer_to_agent` | Switch to named agent | Delegate to specialist |
| `state_delta` | Update session state | Store intermediate data |
| `end_of_agent` | Signal agent completion | Clean exit from agent loop |

### 7.4 Model Parameters

Kagent exposes LLM parameters through the CRD configuration:

```yaml
# Agent CRD model config
spec:
  modelConfig:
    model: gpt-4o
    type: openai
    maxTokens: 4096         # Max output tokens
    temperature: 0.7        # Creativity control
    topP: 0.9               # Nucleus sampling
    frequencyPenalty: 0.0   # Repetition control
    presencePenalty: 0.0    # Topic diversity
    reasoningEffort: medium # For reasoning models
```

**Note:** These control the LLM's OUTPUT behavior. They do NOT control the input context window size or conversation truncation.

---

## 8. Opportunities: Context Optimization Techniques

This section outlines techniques that could be implemented to improve context management in kagent. These are opportunities for contribution, not current features.

### 8.1 Observation Masking

**Concept:** Replace verbose tool outputs with concise summaries or markers before they enter the conversation history.

**Implementation approach:**

```python
# Custom after_tool_callback for observation masking
def observation_masking_callback(tool_context, tool_response):
    """Truncate large tool outputs to prevent context bloat."""
    if isinstance(tool_response, dict):
        output = str(tool_response.get("output", ""))
        if len(output) > MAX_OBSERVATION_LENGTH:
            # Option 1: Simple truncation
            tool_response["output"] = output[:MAX_OBSERVATION_LENGTH] + "\n...[truncated]"

            # Option 2: LLM-based summarization
            # summary = summarize_with_llm(output)
            # tool_response["output"] = summary
    return tool_response

agent = Agent(
    name="my-agent",
    after_tool_callback=observation_masking_callback,
    # ...
)
```

**Where to implement in kagent:** Extend `AgentConfig.to_agent()` in `types.py` to accept callback configuration, or implement as a middleware in the `KAgentMcpToolset`.

### 8.2 LLM-Based Summarization

**Concept:** Use a smaller/cheaper LLM to summarize conversation history when it grows beyond a threshold.

**Two approaches:**

1. **Use Google ADK's built-in compaction** (easiest):
   - Configure `EventsCompactionConfig` on the kagent `App`
   - Uses sliding window with LLM summarization
   - Limited to turn-based triggers

2. **Custom summarization in SessionService** (more flexible):
   - Implement in `_session_service.py`
   - Before returning session to runner, check token count
   - If over threshold, summarize older events with a fast model
   - Replace events with summary event

```python
# Conceptual: Custom session service with summarization
async def get_session(self, app_name, user_id, session_id):
    session = await super().get_session(app_name, user_id, session_id)

    token_count = estimate_tokens(session.events)
    if token_count > MAX_CONTEXT_TOKENS * 0.7:  # 70% threshold
        older_events = session.events[:-KEEP_RECENT_N]
        summary = await summarize_events(older_events, summarizer_model)
        session.events = [summary_event] + session.events[-KEEP_RECENT_N:]

    return session
```

### 8.3 Token-Based Sliding Window

**Concept:** Instead of turn-based compaction, use token counting to manage the window.

**Implementation approach:**

```python
# Token-based window management
def apply_token_window(events, max_tokens, model_name):
    """Keep most recent events that fit within token budget."""
    total_tokens = 0
    keep_from_index = len(events)

    # Walk backwards from most recent
    for i in range(len(events) - 1, -1, -1):
        event_tokens = count_tokens(events[i], model_name)
        if total_tokens + event_tokens > max_tokens:
            break
        total_tokens += event_tokens
        keep_from_index = i

    # Optionally prepend a summary of dropped events
    dropped = events[:keep_from_index]
    if dropped:
        summary = summarize_events(dropped)
        return [summary] + events[keep_from_index:]

    return events
```

**Note:** Google ADK issue [#4146](https://github.com/google/adk-python/issues/4146) tracks a feature request for token-based compaction.

### 8.4 Tool Output Truncation

**Concept:** Limit the size of tool outputs before they enter the conversation.

**Implementation options:**

1. **Character limit:** Truncate outputs beyond N characters
2. **Structured extraction:** For JSON/YAML outputs, extract only relevant fields
3. **Streaming compression:** For large outputs, stream through a summarizer
4. **Content-type aware:** Different strategies for different tool output types

```python
# Content-type aware truncation
TRUNCATION_RULES = {
    "kubectl_get": {"max_items": 10, "fields": ["name", "status", "age"]},
    "log_search": {"max_lines": 50, "strategy": "head_tail"},
    "default": {"max_chars": 4000, "strategy": "truncate"},
}
```

### 8.5 Selective History Pruning

**Concept:** Not all conversation events are equally important. Prune less important ones first.

**Priority order (keep longest):**
1. System prompt (always keep)
2. Most recent user message (always keep)
3. Most recent N tool call/response pairs
4. Earlier user messages (summarize)
5. Earlier tool results (drop or summarize first)

---

## 9. Key Source Files Reference

### Kagent-specific

| File | Purpose |
|------|---------|
| `python/packages/kagent-adk/src/kagent/adk/_agent_executor.py` | Main execution entry point; per-request runner lifecycle |
| `python/packages/kagent-adk/src/kagent/adk/types.py` | Agent config → Google ADK Agent creation, tool setup |
| `python/packages/kagent-adk/src/kagent/adk/_session_service.py` | Session persistence to Go backend |
| `python/packages/kagent-adk/src/kagent/adk/_mcp_toolset.py` | MCP toolset wrapper with error handling |
| `python/packages/kagent-adk/src/kagent/adk/converters/event_converter.py` | ADK event → A2A event conversion |
| `python/packages/kagent-adk/src/kagent/adk/converters/request_converter.py` | A2A request → ADK run args conversion |
| `python/packages/kagent-adk/src/kagent/adk/converters/part_converter.py` | Part format conversions (function_call, function_response) |
| `python/packages/kagent-adk/src/kagent/adk/models/_litellm.py` | LiteLLM integration (Anthropic, Ollama, Bedrock) |
| `go/internal/controller/reconciler/reconciler.go` | MCP server connection and tool listing |
| `go/internal/httpserver/handlers/tools.go` | Tool REST API |
| `go/internal/httpserver/handlers/toolservers.go` | Tool server REST API |
| `go/internal/mcp/mcp_handler.go` | A2A MCP handler (list_agents, invoke_agent) |
| `go/api/v1alpha2/agent_types.go` | Agent CRD definition |

### Google ADK (external dependency, v1.25.0+)

| Module | Purpose |
|--------|---------|
| `google.adk.runners.Runner` | Core agent loop orchestration |
| `google.adk.flows.llm_flows.BaseLlmFlow` | LLM call → tool execution → loop cycle |
| `google.adk.flows.llm_flows.functions` | Tool dispatch, parallel execution, callback invocation |
| `google.adk.agents.LlmAgent` | Agent with callbacks, tool definitions, model binding |
| `google.adk.tools.mcp_tool.McpToolset` | MCP tool integration base |
| `google.adk.events.EventActions` | Flow control (skip_summarization, escalate, transfer) |
| `google.adk.apps.app.EventsCompactionConfig` | Context compaction configuration |

---

## 10. Diagrams

### Agent Loop Decision Flow

```
                    ┌──────────────────┐
                    │   User Message   │
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │  Build LLM       │
                    │  Request:        │
                    │  - System prompt │
                    │  - History       │◄──────────────────────────┐
                    │  - Tool defs     │                           │
                    └────────┬─────────┘                           │
                             │                                     │
                             ▼                                     │
                    ┌──────────────────┐                           │
                    │  before_model    │──── short-circuit? ──►    │
                    │  callback        │     (return cached)       │
                    └────────┬─────────┘                           │
                             │ (no short-circuit)                  │
                             ▼                                     │
                    ┌──────────────────┐                           │
                    │   LLM API Call   │                           │
                    └────────┬─────────┘                           │
                             │                                     │
                             ▼                                     │
                    ┌──────────────────┐                           │
                    │  after_model     │                           │
                    │  callback        │                           │
                    └────────┬─────────┘                           │
                             │                                     │
                    ┌────────┴─────────┐                           │
                    │                  │                            │
              has tool calls?    text only?                        │
                    │                  │                            │
                    ▼                  ▼                            │
           ┌──────────────┐  ┌──────────────┐                     │
           │ before_tool  │  │ FINAL        │                     │
           │ callback     │  │ RESPONSE     │                     │
           │ (per tool)   │  │ (exit loop)  │                     │
           └──────┬───────┘  └──────────────┘                     │
                  │                                                │
                  ▼                                                │
           ┌──────────────┐                                        │
           │  Execute     │   ← Tools run in PARALLEL              │
           │  Tools       │     via asyncio.gather()               │
           │  (MCP/local) │                                        │
           └──────┬───────┘                                        │
                  │                                                │
                  ▼                                                │
           ┌──────────────┐                                        │
           │  after_tool  │                                        │
           │  callback    │                                        │
           │  (per tool)  │                                        │
           └──────┬───────┘                                        │
                  │                                                │
                  ▼                                                │
           ┌──────────────┐                                        │
           │  Merge tool  │                                        │
           │  results     │────────────────────────────────────────┘
           │  into event  │   (append to history, loop back)
           └──────────────┘
```

### MCP Tool Lifecycle

```
┌─────────────────────────────────────────────────────────────────┐
│                    DISCOVERY (Go Controller)                     │
│                                                                  │
│  RemoteMCPServer CR ──► Controller reconciles every 60s          │
│                              │                                   │
│                              ▼                                   │
│                    Connect via SSE/HTTP transport                 │
│                              │                                   │
│                              ▼                                   │
│                    session.ListTools() ──► Store in DB            │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                   CONFIGURATION (Agent CRD)                      │
│                                                                  │
│  Agent CR specifies:                                             │
│    - Which MCP servers to use                                    │
│    - Which specific tools (tool_filter)                          │
│    - Headers to propagate                                        │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                  RUNTIME (Python ADK per-request)                 │
│                                                                  │
│  1. Fresh Runner created with KAgentMcpToolset instances         │
│  2. Toolset connects to MCP server, fetches tool schemas         │
│  3. Tool schemas converted to LLM-compatible function defs       │
│  4. LLM decides which tool(s) to call                            │
│  5. McpToolset proxies call to MCP server                        │
│  6. Result returned as function_response                         │
│  7. Runner.close() cleans up all MCP connections                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## Summary: Key Answers

| Question | Answer |
|----------|--------|
| **When is LLM called?** | After every tool execution round. LLM is called N+1 times for N tool rounds. |
| **How many tools per LLM call?** | LLM decides. Multiple tools in one response execute in parallel. |
| **Is there conversation truncation?** | No. Kagent sends full history every time. No built-in truncation. |
| **How to control tool count?** | Use `tool_filter` to limit available tools. LLM decides what to use. |
| **How to control tool output?** | Use `after_tool_callback` to truncate/transform results. |
| **How to manage context limits?** | Currently not managed. Can use ADK's `EventsCompactionConfig` or custom session-level summarization. |
| **Can tool calls be intercepted?** | Yes, via `before_tool_callback` (skip execution) and `after_tool_callback` (modify results). |
| **Is observation masking supported?** | Not built-in. Implement via `after_tool_callback` or custom session service. |

---

**Last updated:** 2026-02-23
**Related docs:** [Controller Reconciliation](controller-reconciliation.md), [CLAUDE.md](../../CLAUDE.md)
