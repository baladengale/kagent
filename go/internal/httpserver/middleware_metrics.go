package httpserver

import (
	"fmt"
	"net/http"
	"time"

	"github.com/kagent-dev/kagent/go/internal/metrics"
)

// metricsMiddleware records Prometheus metrics for HTTP requests.
func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		metrics.HTTPRequestsInFlight.Inc()
		defer metrics.HTTPRequestsInFlight.Dec()

		mw := &metricsResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(mw, r)

		duration := time.Since(start).Seconds()
		path := normalizePath(r.URL.Path)
		statusCode := fmt.Sprintf("%d", mw.statusCode)

		metrics.HTTPRequestsTotal.WithLabelValues(r.Method, path, statusCode).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
		metrics.HTTPResponseSizeBytes.WithLabelValues(r.Method, path).Observe(float64(mw.bytesWritten))
	})
}

type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (w *metricsResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *metricsResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += n
	return n, err
}

func (w *metricsResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// normalizePath reduces path cardinality for metrics by replacing dynamic
// segments with placeholders.
func normalizePath(path string) string {
	// Map known API prefixes to their base routes to avoid high-cardinality labels
	prefixes := []string{
		APIPathSessions,
		APIPathTasks,
		APIPathAgents,
		APIPathModelConfig,
		APIPathToolServers,
		APIPathTools,
		APIPathToolServerTypes,
		APIPathModelProviderConfigs,
		APIPathModels,
		APIPathMemories,
		APIPathNamespaces,
		APIPathFeedback,
		APIPathLangGraph,
		APIPathCrewAI,
		APIPathA2A,
		APIPathMCP,
		APIPathRuns,
	}

	for _, prefix := range prefixes {
		if len(path) > len(prefix) && path[:len(prefix)] == prefix {
			return prefix
		}
		if path == prefix {
			return prefix
		}
	}

	if path == APIPathHealth || path == APIPathVersion {
		return path
	}

	return "/other"
}
