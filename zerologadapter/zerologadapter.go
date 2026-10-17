// Package zerologadapter renders go-screenplay narration through
// github.com/rs/zerolog, preserving structure (actor, kind, depth, error) as
// fields.
package zerologadapter

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/grandper/go-screenplay/v2/screenplay"
)

// Adapter writes narration to a *zerolog.Logger.
type Adapter struct {
	logger *zerolog.Logger
}

// New returns an Adapter writing to the given logger. If logger is nil, the
// global logger of github.com/rs/zerolog/log is used.
func New(logger *zerolog.Logger) *Adapter {
	if logger == nil {
		logger = &log.Logger
	}

	return &Adapter{logger: logger}
}

// Narrate implements screenplay.Adapter.
func (a *Adapter) Narrate(event screenplay.Event) {
	// One structured event per step (on begin) plus outcomes worth recording.
	if event.Phase == screenplay.PhaseEnd && event.Err == nil && event.Answer == nil {
		return
	}

	entry := a.logger.WithLevel(mapLevel(event.Level)).
		Str("kind", event.Kind.String()).
		Int("depth", event.Depth)

	if event.Actor != "" {
		entry = entry.Str("actor", event.Actor)
	}

	if event.Answer != nil {
		entry = entry.Interface("answer", event.Answer)
	}

	if event.Err != nil {
		entry = entry.Err(event.Err)
	}

	entry.Msg(event.Message)
}

// mapLevel maps a screenplay.Level onto a zerolog.Level.
func mapLevel(level screenplay.Level) zerolog.Level {
	switch level {
	case screenplay.LevelDebug:
		return zerolog.DebugLevel
	case screenplay.LevelInfo:
		return zerolog.InfoLevel
	case screenplay.LevelWarn:
		return zerolog.WarnLevel
	case screenplay.LevelError:
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

// Adapter implements the screenplay.Adapter interface.
var _ screenplay.Adapter = (*Adapter)(nil)
