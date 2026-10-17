// Package zapadapter renders go-screenplay narration through go.uber.org/zap,
// preserving structure (actor, kind, depth, error) as fields.
package zapadapter

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/grandper/go-screenplay/v2/screenplay"
)

// Adapter writes narration to a *zap.Logger.
type Adapter struct {
	logger *zap.Logger
}

// New returns an Adapter writing to the given logger. If logger is nil,
// zap.L() is used.
func New(logger *zap.Logger) *Adapter {
	if logger == nil {
		logger = zap.L()
	}

	return &Adapter{logger: logger}
}

// Narrate implements screenplay.Adapter.
func (a *Adapter) Narrate(event screenplay.Event) {
	// One structured entry per step (on begin) plus outcomes worth recording.
	if event.Phase == screenplay.PhaseEnd && event.Err == nil && event.Answer == nil {
		return
	}

	fields := []zap.Field{
		zap.String("kind", event.Kind.String()),
		zap.Int("depth", event.Depth),
	}

	if event.Actor != "" {
		fields = append(fields, zap.String("actor", event.Actor))
	}

	if event.Answer != nil {
		fields = append(fields, zap.Any("answer", event.Answer))
	}

	if event.Err != nil {
		fields = append(fields, zap.Error(event.Err))
	}

	a.logger.Log(mapLevel(event.Level), event.Message, fields...)
}

// mapLevel maps a screenplay.Level onto a zapcore.Level.
func mapLevel(level screenplay.Level) zapcore.Level {
	switch level {
	case screenplay.LevelDebug:
		return zapcore.DebugLevel
	case screenplay.LevelInfo:
		return zapcore.InfoLevel
	case screenplay.LevelWarn:
		return zapcore.WarnLevel
	case screenplay.LevelError:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// Adapter implements the screenplay.Adapter interface.
var _ screenplay.Adapter = (*Adapter)(nil)
