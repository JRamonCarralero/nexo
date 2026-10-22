package logger_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/logger"
)

func TestJSONUsesCloudLoggingKeysAndTraceID(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf, "production", slog.LevelInfo)

	ctx := logger.WithTraceID(context.Background(), "abc123")
	log.WarnContext(ctx, "something odd", slog.String("task_id", "t1"))

	var rec map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &rec))
	assert.Equal(t, "WARNING", rec["severity"])
	assert.Equal(t, "something odd", rec["message"])
	assert.Equal(t, "abc123", rec["trace_id"])
	assert.Equal(t, "t1", rec["task_id"])
}

func TestNoTraceIDWhenAbsent(t *testing.T) {
	var buf bytes.Buffer
	logger.New(&buf, "production", slog.LevelInfo).Info("hello")

	var rec map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &rec))
	assert.NotContains(t, rec, "trace_id")
}
