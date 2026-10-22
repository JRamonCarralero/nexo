package httpserver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JRamonCarralero/nexo/backend/internal/httpserver"
	"github.com/JRamonCarralero/nexo/backend/internal/httpserver/api"
	"github.com/JRamonCarralero/nexo/backend/internal/modules/system"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/logger"
)

type fakePinger struct{ err error }

func (f fakePinger) Ping(context.Context) error { return f.err }

func newRouter(t *testing.T, pingErr error) (*gin.Engine, *bytes.Buffer) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	var logs bytes.Buffer
	log := logger.New(&logs, "test", slog.LevelDebug)
	r := httpserver.NewRouter(log, httpserver.Handlers{
		System: system.NewHandler(fakePinger{err: pingErr}, "test"),
	})
	return r, &logs
}

func do(r http.Handler, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(context.Background(), method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func decodeProblem(t *testing.T, body io.Reader) api.Problem {
	t.Helper()
	var p api.Problem
	require.NoError(t, json.NewDecoder(body).Decode(&p))
	return p
}

func TestHealthOK(t *testing.T) {
	r, _ := newRouter(t, nil)

	rec := do(r, http.MethodGet, "/api/v1/health", nil)

	require.Equal(t, http.StatusOK, rec.Code)
	var h api.Health
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&h))
	assert.Equal(t, api.HealthStatusOk, h.Status)
	assert.Equal(t, api.HealthChecksOk, h.Checks["database"])
	assert.Len(t, rec.Header().Get("X-Trace-Id"), 32)
}

func TestHealthDatabaseDownReturnsProblem(t *testing.T) {
	r, logs := newRouter(t, errors.New("connection refused"))

	rec := do(r, http.MethodGet, "/api/v1/health", nil)

	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
	p := decodeProblem(t, rec.Body)
	assert.Equal(t, "SERVICE_UNAVAILABLE", p.Code)
	assert.Equal(t, rec.Header().Get("X-Trace-Id"), p.TraceId)
	assert.NotContains(t, *p.Detail, "connection refused", "internal causes must not reach clients")

	assert.Contains(t, logs.String(), `"severity":"ERROR"`)
	assert.Contains(t, logs.String(), "connection refused", "the cause must be in the logs")
}

func TestUnknownRouteReturnsProblem(t *testing.T) {
	r, logs := newRouter(t, nil)

	rec := do(r, http.MethodGet, "/api/v1/nope", nil)

	require.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "ROUTE_NOT_FOUND", decodeProblem(t, rec.Body).Code)
	assert.Contains(t, logs.String(), `"severity":"WARNING"`)
}

func TestPanicIsRecoveredAsProblem(t *testing.T) {
	r, logs := newRouter(t, nil)
	r.GET("/boom", func(*gin.Context) { panic("kaboom") })

	rec := do(r, http.MethodGet, "/boom", nil)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "INTERNAL_ERROR", decodeProblem(t, rec.Body).Code)
	assert.Contains(t, logs.String(), "kaboom")
}

func TestIncomingTraceparentIsReused(t *testing.T) {
	r, logs := newRouter(t, nil)
	const traceID = "4bf92f3577b34da6a3ce929d0e0e4736"

	rec := do(r, http.MethodGet, "/api/v1/health", map[string]string{
		"traceparent": "00-" + traceID + "-00f067aa0ba902b7-01",
	})

	assert.Equal(t, traceID, rec.Header().Get("X-Trace-Id"))
	assert.Contains(t, logs.String(), traceID)
}
