package client

// Package logger contract for client HTTP operations.
//
// T6.1 / ADR-007 / val-006-no-logger-injection-client.

import (
	"context"
	"log/slog"
)

// Logger is the minimal interface the client uses for observability.
// Implementations: NoopLogger (default), SlogLogger (stdlib slog adapter),
// or any consumer-provided struct that satisfies the two methods.
//
// Design rules (.claude/rules/dependencies.md):
//   - No third-party logging library dependency.
//   - kv pairs MUST be even-length: key1, value1, key2, value2, ...
//   - Implementations MUST NOT panic; if a panic occurs in the consumer's
//     Logger, doRequest still completes (defensive recover wrapper used).
type Logger interface {
	Debug(msg string, kv ...any)
	Error(msg string, kv ...any)
}

// NoopLogger is the default Logger that discards all messages.
type NoopLogger struct{}

// Debug implements Logger.
func (NoopLogger) Debug(string, ...any) {}

// Error implements Logger.
func (NoopLogger) Error(string, ...any) {}

// SlogLogger adapts a *slog.Logger to the Logger interface. Provided as
// a convenience for consumers who already use the Go standard log/slog;
// not required, and importing this type does NOT force a slog dependency
// because slog is in the standard library.
type SlogLogger struct {
	L *slog.Logger
}

// Debug implements Logger.
func (s SlogLogger) Debug(msg string, kv ...any) {
	if s.L == nil {
		return
	}
	s.L.LogAttrs(context.Background(), slog.LevelDebug, msg, kvToAttrs(kv)...)
}

// Error implements Logger.
func (s SlogLogger) Error(msg string, kv ...any) {
	if s.L == nil {
		return
	}
	s.L.LogAttrs(context.Background(), slog.LevelError, msg, kvToAttrs(kv)...)
}

// kvToAttrs converts a flat key/value slice to slog.Attr.
// Keys are coerced to string; values are passed through slog.Any.
// Odd-length kv slices append a "MISSING" placeholder rather than panic.
func kvToAttrs(kv []any) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(kv)/2)
	for i := 0; i+1 < len(kv); i += 2 {
		key, ok := kv[i].(string)
		if !ok {
			key = "INVALID_KEY"
		}
		attrs = append(attrs, slog.Any(key, kv[i+1]))
	}
	if len(kv)%2 == 1 {
		attrs = append(attrs, slog.String("MISSING", "value-for-last-key"))
	}
	return attrs
}

// logSafe calls fn (Debug or Error) on lg while swallowing any panic so
// the HTTP request completes. (.claude/rules/error-handling.md §Forbidden
// allows recover() in adapter boundaries when the alternative is breaking
// an unrelated request.)
func logSafe(fn func(string, ...any), msg string, kv ...any) {
	defer func() { _ = recover() }()
	fn(msg, kv...)
}
