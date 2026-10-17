package logrusadapter_test

import (
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-narrators/logrusadapter"
)

func TestLogrusAdapter(t *testing.T) {
	t.Parallel()

	t.Run("emits a structured entry for a beat", func(t *testing.T) {
		t.Parallel()

		logger, hook := test.NewNullLogger()
		adapter := logrusadapter.New(logger)

		adapter.Narrate(screenplay.Event{
			Kind:    screenplay.KindBeat,
			Level:   screenplay.LevelInfo,
			Phase:   screenplay.PhaseBegin,
			Depth:   1,
			Actor:   "Wanda",
			Message: "Wanda sends a GET request",
		})

		entry := single(t, hook)
		assert.Equal(t, "Wanda sends a GET request", entry.Message)
		assert.Equal(t, logrus.InfoLevel, entry.Level)
		assert.Equal(t, "beat", entry.Data["kind"])
		assert.Equal(t, "Wanda", entry.Data["actor"])
		assert.Equal(t, 1, entry.Data["depth"])
	})

	t.Run("stays silent on a successful end phase", func(t *testing.T) {
		t.Parallel()

		logger, hook := test.NewNullLogger()
		adapter := logrusadapter.New(logger)

		adapter.Narrate(screenplay.Event{Kind: screenplay.KindBeat, Phase: screenplay.PhaseEnd})

		assert.Empty(t, hook.AllEntries())
	})

	t.Run("records an answer", func(t *testing.T) {
		t.Parallel()

		logger, hook := test.NewNullLogger()
		adapter := logrusadapter.New(logger)

		adapter.Narrate(screenplay.Event{
			Kind:   screenplay.KindBeat,
			Level:  screenplay.LevelInfo,
			Phase:  screenplay.PhaseEnd,
			Answer: 200,
		})

		entry := single(t, hook)
		assert.Equal(t, 200, entry.Data["answer"])
	})

	t.Run("records a failure at the error level", func(t *testing.T) {
		t.Parallel()

		logger, hook := test.NewNullLogger()
		adapter := logrusadapter.New(logger)
		failure := errors.New("connection refused")

		adapter.Narrate(screenplay.Event{
			Kind:    screenplay.KindBeat,
			Level:   screenplay.LevelError,
			Phase:   screenplay.PhaseEnd,
			Message: "Wanda sends a GET request",
			Err:     failure,
		})

		entry := single(t, hook)
		assert.Equal(t, logrus.ErrorLevel, entry.Level)
		assert.Equal(t, failure, entry.Data[logrus.ErrorKey])
	})

	t.Run("maps every screenplay level onto its logrus level", func(t *testing.T) {
		t.Parallel()

		levels := map[screenplay.Level]logrus.Level{
			screenplay.LevelDebug: logrus.DebugLevel,
			screenplay.LevelInfo:  logrus.InfoLevel,
			screenplay.LevelWarn:  logrus.WarnLevel,
			screenplay.LevelError: logrus.ErrorLevel,
			screenplay.Level(42):  logrus.InfoLevel,
		}

		for level, want := range levels {
			logger, hook := test.NewNullLogger()
			logger.SetLevel(logrus.DebugLevel)
			adapter := logrusadapter.New(logger)

			adapter.Narrate(screenplay.Event{
				Kind:    screenplay.KindAside,
				Level:   level,
				Phase:   screenplay.PhaseBegin,
				Message: "an aside",
			})

			entry := single(t, hook)
			assert.Equal(t, want, entry.Level)
		}
	})

	t.Run("falls back on the standard logger", func(t *testing.T) {
		t.Parallel()

		adapter := logrusadapter.New(nil)

		assert.NotPanics(t, func() {
			adapter.Narrate(screenplay.Event{Kind: screenplay.KindAside, Level: screenplay.LevelDebug})
		})
	})
}

func single(t *testing.T, hook *test.Hook) *logrus.Entry {
	t.Helper()

	entries := hook.AllEntries()
	require.Len(t, entries, 1)

	return entries[0]
}
