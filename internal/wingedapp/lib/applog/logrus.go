package applog

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
)

// LogrusLogger wraps logrus to implement the Logger interface.
type LogrusLogger struct {
	entry *logrus.Entry
}

// NewLogrus creates a new logrus-backed Logger.
func NewLogrus(serviceName string) *LogrusLogger {
	log := logrus.New()
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})

	entry := log.WithField("service", serviceName)
	return &LogrusLogger{entry: entry}
}

// NewLogrusFromEntry creates a Logger from an existing logrus.Entry.
func NewLogrusFromEntry(entry *logrus.Entry) *LogrusLogger {
	return &LogrusLogger{entry: entry}
}

// SetLevel sets the logging level.
func (l *LogrusLogger) SetLevel(level logrus.Level) {
	l.entry.Logger.SetLevel(level)
}

func (l *LogrusLogger) toLogrusFields(fields []Field) logrus.Fields {
	lf := make(logrus.Fields, len(fields))
	for _, f := range fields {
		lf[f.Key] = f.Value
	}
	return lf
}

func (l *LogrusLogger) extractRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(CtxKeyRequestID).(string); ok {
		return id
	}
	return ""
}

func (l *LogrusLogger) withContext(ctx context.Context) *logrus.Entry {
	entry := l.entry
	if reqID := l.extractRequestID(ctx); reqID != "" {
		entry = entry.WithField("request_id", reqID)
	}
	return entry
}

func (l *LogrusLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.withContext(ctx).WithFields(l.toLogrusFields(fields)).Debug(msg)
}

func (l *LogrusLogger) Info(ctx context.Context, msg string, fields ...Field) {
	l.withContext(ctx).WithFields(l.toLogrusFields(fields)).Info(msg)
}

func (l *LogrusLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.withContext(ctx).WithFields(l.toLogrusFields(fields)).Warn(msg)
}

func (l *LogrusLogger) Error(ctx context.Context, msg string, err error, fields ...Field) {
	entry := l.withContext(ctx).WithFields(l.toLogrusFields(fields))
	if err != nil {
		entry = entry.WithError(err)
	}
	entry.Error(msg)
}

func (l *LogrusLogger) Fatal(ctx context.Context, msg string, err error, fields ...Field) {
	entry := l.withContext(ctx).WithFields(l.toLogrusFields(fields))
	if err != nil {
		entry = entry.WithError(err)
	}
	entry.Fatal(msg)
}

func (l *LogrusLogger) WithFields(fields ...Field) Logger {
	return &LogrusLogger{
		entry: l.entry.WithFields(l.toLogrusFields(fields)),
	}
}

// Entry returns the underlying logrus.Entry for migration compatibility.
func (l *LogrusLogger) Entry() *logrus.Entry {
	return l.entry
}
