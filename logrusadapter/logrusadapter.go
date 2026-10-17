// Package logrusadapter renders go-screenplay narration through
// github.com/sirupsen/logrus, preserving structure (actor, kind, depth, error)
// as fields.
package logrusadapter

import (
	"github.com/sirupsen/logrus"

	"github.com/grandper/go-screenplay/v2/screenplay"
)

// Adapter writes narration to a *logrus.Logger.
type Adapter struct {
	logger *logrus.Logger
}

// New returns an Adapter writing to the given logger. If logger is nil,
// logrus.StandardLogger() is used.
func New(logger *logrus.Logger) *Adapter {
	if logger == nil {
		logger = logrus.StandardLogger()
	}

	return &Adapter{logger: logger}
}

// Narrate implements screenplay.Adapter.
func (a *Adapter) Narrate(event screenplay.Event) {
	// One structured entry per step (on begin) plus outcomes worth recording.
	if event.Phase == screenplay.PhaseEnd && event.Err == nil && event.Answer == nil {
		return
	}

	fields := logrus.Fields{
		"kind":  event.Kind.String(),
		"depth": event.Depth,
	}

	if event.Actor != "" {
		fields["actor"] = event.Actor
	}

	if event.Answer != nil {
		fields["answer"] = event.Answer
	}

	if event.Err != nil {
		fields[logrus.ErrorKey] = event.Err
	}

	a.logger.WithFields(fields).Log(mapLevel(event.Level), event.Message)
}

// mapLevel maps a screenplay.Level onto a logrus.Level.
func mapLevel(level screenplay.Level) logrus.Level {
	switch level {
	case screenplay.LevelDebug:
		return logrus.DebugLevel
	case screenplay.LevelInfo:
		return logrus.InfoLevel
	case screenplay.LevelWarn:
		return logrus.WarnLevel
	case screenplay.LevelError:
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}

// Adapter implements the screenplay.Adapter interface.
var _ screenplay.Adapter = (*Adapter)(nil)
