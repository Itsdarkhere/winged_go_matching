package applog

import "context"

// NopLogger is a no-op logger that discards all log entries.
// Useful for optional logging where nil checks would be verbose.
type NopLogger struct{}

func (n *NopLogger) Debug(ctx context.Context, msg string, fields ...Field) {}
func (n *NopLogger) Info(ctx context.Context, msg string, fields ...Field)  {}
func (n *NopLogger) Warn(ctx context.Context, msg string, fields ...Field)  {}
func (n *NopLogger) Error(ctx context.Context, msg string, err error, fields ...Field) {
}
func (n *NopLogger) Fatal(ctx context.Context, msg string, err error, fields ...Field) {
}
func (n *NopLogger) WithFields(fields ...Field) Logger { return n }
