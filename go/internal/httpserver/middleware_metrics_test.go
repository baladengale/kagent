package httpserver

import (
	"testing"
)

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "health endpoint",
			path:     "/health",
			expected: "/health",
		},
		{
			name:     "version endpoint",
			path:     "/version",
			expected: "/version",
		},
		{
			name:     "agents list",
			path:     "/api/agents",
			expected: "/api/agents",
		},
		{
			name:     "agent with namespace and name",
			path:     "/api/agents/default/my-agent",
			expected: "/api/agents",
		},
		{
			name:     "sessions list",
			path:     "/api/sessions",
			expected: "/api/sessions",
		},
		{
			name:     "session by ID",
			path:     "/api/sessions/abc-123",
			expected: "/api/sessions",
		},
		{
			name:     "tasks by ID",
			path:     "/api/tasks/task-123",
			expected: "/api/tasks",
		},
		{
			name:     "tools list",
			path:     "/api/tools",
			expected: "/api/tools",
		},
		{
			name:     "tool servers list",
			path:     "/api/toolservers",
			expected: "/api/toolservers",
		},
		{
			name:     "model configs list",
			path:     "/api/modelconfigs",
			expected: "/api/modelconfigs",
		},
		{
			name:     "model configs with namespace",
			path:     "/api/modelconfigs/default/my-config",
			expected: "/api/modelconfigs",
		},
		{
			name:     "unknown path",
			path:     "/unknown/path",
			expected: "/other",
		},
		{
			name:     "feedback",
			path:     "/api/feedback",
			expected: "/api/feedback",
		},
		{
			name:     "memories with detail",
			path:     "/api/memories/default/my-memory",
			expected: "/api/memories",
		},
		{
			name:     "a2a with detail",
			path:     "/api/a2a/default/my-agent",
			expected: "/api/a2a",
		},
		{
			name:     "mcp endpoint",
			path:     "/mcp/some/path",
			expected: "/mcp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePath(tt.path)
			if result != tt.expected {
				t.Errorf("normalizePath(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}
