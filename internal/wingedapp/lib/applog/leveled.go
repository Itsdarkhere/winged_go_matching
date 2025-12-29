package applog

import "context"

// LeveledLoggerAdapter adapts applog.Logger to retryablehttp.LeveledLogger interface.
// retryablehttp.LeveledLogger expects: Error, Info, Debug, Warn(msg string, keysAndValues ...interface{})
type LeveledLoggerAdapter struct {
	logger Logger
}

// NewLeveledLogger creates an adapter for use with retryablehttp.Client.
func NewLeveledLogger(l Logger) *LeveledLoggerAdapter {
	return &LeveledLoggerAdapter{logger: l}
}

func (a *LeveledLoggerAdapter) toFields(keysAndValues []interface{}) []Field {
	fields := make([]Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues)-1; i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			continue
		}
		fields = append(fields, F(key, keysAndValues[i+1]))
	}
	return fields
}

func (a *LeveledLoggerAdapter) Error(msg string, keysAndValues ...interface{}) {
	a.logger.Error(context.Background(), msg, nil, a.toFields(keysAndValues)...)
}

func (a *LeveledLoggerAdapter) Info(msg string, keysAndValues ...interface{}) {
	a.logger.Info(context.Background(), msg, a.toFields(keysAndValues)...)
}

func (a *LeveledLoggerAdapter) Debug(msg string, keysAndValues ...interface{}) {
	a.logger.Debug(context.Background(), msg, a.toFields(keysAndValues)...)
}

func (a *LeveledLoggerAdapter) Warn(msg string, keysAndValues ...interface{}) {
	a.logger.Warn(context.Background(), msg, a.toFields(keysAndValues)...)
}
