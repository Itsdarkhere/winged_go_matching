package applog

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// LogfireLogger implements Logger using OpenTelemetry traces sent to Logfire.
type LogfireLogger struct {
	tracer     trace.Tracer
	baseAttrs  []attribute.KeyValue
	severities map[string]attribute.KeyValue
}

// NewLogfire creates a new Logfire-backed Logger.
func NewLogfire(serviceName string) *LogfireLogger {
	tracer := otel.Tracer(serviceName)
	return &LogfireLogger{
		tracer: tracer,
		baseAttrs: []attribute.KeyValue{
			attribute.String("service", serviceName),
		},
		severities: map[string]attribute.KeyValue{
			"debug": attribute.String("level", "debug"),
			"info":  attribute.String("level", "info"),
			"warn":  attribute.String("level", "warn"),
			"error": attribute.String("level", "error"),
			"fatal": attribute.String("level", "fatal"),
		},
	}
}

func (l *LogfireLogger) toOtelAttrs(fields []Field) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(fields)+len(l.baseAttrs))
	attrs = append(attrs, l.baseAttrs...)
	for _, f := range fields {
		attrs = append(attrs, toOtelAttr(f))
	}
	return attrs
}

func toOtelAttr(f Field) attribute.KeyValue {
	switch v := f.Value.(type) {
	case string:
		return attribute.String(f.Key, v)
	case int:
		return attribute.Int(f.Key, v)
	case int64:
		return attribute.Int64(f.Key, v)
	case float64:
		return attribute.Float64(f.Key, v)
	case bool:
		return attribute.Bool(f.Key, v)
	default:
		return attribute.String(f.Key, fmt.Sprintf("%v", v))
	}
}

func (l *LogfireLogger) extractRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(CtxKeyRequestID).(string); ok {
		return id
	}
	return ""
}

func (l *LogfireLogger) log(ctx context.Context, level, msg string, err error, fields []Field) {
	attrs := l.toOtelAttrs(fields)
	attrs = append(attrs, l.severities[level])
	attrs = append(attrs, attribute.String("message", msg))

	if reqID := l.extractRequestID(ctx); reqID != "" {
		attrs = append(attrs, attribute.String("request_id", reqID))
	}

	_, span := l.tracer.Start(ctx, msg, trace.WithAttributes(attrs...))
	defer span.End()

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

func (l *LogfireLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, "debug", msg, nil, fields)
}

func (l *LogfireLogger) Info(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, "info", msg, nil, fields)
}

func (l *LogfireLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, "warn", msg, nil, fields)
}

func (l *LogfireLogger) Error(ctx context.Context, msg string, err error, fields ...Field) {
	l.log(ctx, "error", msg, err, fields)
}

func (l *LogfireLogger) Fatal(ctx context.Context, msg string, err error, fields ...Field) {
	l.log(ctx, "fatal", msg, err, fields)
}

func (l *LogfireLogger) WithFields(fields ...Field) Logger {
	newAttrs := make([]attribute.KeyValue, len(l.baseAttrs), len(l.baseAttrs)+len(fields))
	copy(newAttrs, l.baseAttrs)
	for _, f := range fields {
		newAttrs = append(newAttrs, toOtelAttr(f))
	}
	return &LogfireLogger{
		tracer:     l.tracer,
		baseAttrs:  newAttrs,
		severities: l.severities,
	}
}
