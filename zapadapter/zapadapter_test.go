package zapadapter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-narrators/zapadapter"
)

func TestZapAdapter(t *testing.T) {
	t.Parallel()

	t.Run("emits a structured entry for a beat", func(t *testing.T) {
		t.Parallel()

		core, logs := observer.New(zapcore.DebugLevel)
		adapter := zapadapter.New(zap.New(core))

		adapter.Narrate(screenplay.Event{
			Kind:    screenplay.KindBeat,
			Level:   screenplay.LevelInfo,
			Phase:   screenplay.PhaseBegin,
			Depth:   1,
			Actor:   "Wanda",
			Message: "Wanda sends a GET request",
		})

		entry := single(t, logs)
		assert.Equal(t, "Wanda sends a GET request", entry.Message)
		assert.Equal(t, zapcore.InfoLevel, entry.Level)
		assert.Equal(t, "beat", entry.ContextMap()["kind"])
		assert.Equal(t, "Wanda", entry.ContextMap()["actor"])
		assert.Equal(t, int64(1), entry.ContextMap()["depth"])
	})

	t.Run("stays silent on a successful end phase", func(t *testing.T) {
		t.Parallel()

		core, logs := observer.New(zapcore.DebugLevel)
		adapter := zapadapter.New(zap.New(core))

		adapter.Narrate(screenplay.Event{Kind: screenplay.KindBeat, Phase: screenplay.PhaseEnd})

		assert.Zero(t, logs.Len())
	})

	t.Run("records an answer", func(t *testing.T) {
		t.Parallel()

		core, logs := observer.New(zapcore.DebugLevel)
		adapter := zapadapter.New(zap.New(core))

		adapter.Narrate(screenplay.Event{
			Kind:   screenplay.KindBeat,
			Level:  screenplay.LevelInfo,
			Phase:  screenplay.PhaseEnd,
			Answer: 200,
		})

		entry := single(t, logs)
		assert.Equal(t, int64(200), entry.ContextMap()["answer"])
	})

	t.Run("records a failure at the error level", func(t *testing.T) {
		t.Parallel()

		core, logs := observer.New(zapcore.DebugLevel)
		adapter := zapadapter.New(zap.New(core))

		adapter.Narrate(screenplay.Event{
			Kind:    screenplay.KindBeat,
			Level:   screenplay.LevelError,
			Phase:   screenplay.PhaseEnd,
			Message: "Wanda sends a GET request",
			Err:     errors.New("connection refused"),
		})

		entry := single(t, logs)
		assert.Equal(t, zapcore.ErrorLevel, entry.Level)
		assert.Equal(t, "connection refused", entry.ContextMap()["error"])
	})

	t.Run("maps every screenplay level onto its zap level", func(t *testing.T) {
		t.Parallel()

		levels := map[screenplay.Level]zapcore.Level{
			screenplay.LevelDebug: zapcore.DebugLevel,
			screenplay.LevelInfo:  zapcore.InfoLevel,
			screenplay.LevelWarn:  zapcore.WarnLevel,
			screenplay.LevelError: zapcore.ErrorLevel,
			screenplay.Level(42):  zapcore.InfoLevel,
		}

		for level, want := range levels {
			core, logs := observer.New(zapcore.DebugLevel)
			adapter := zapadapter.New(zap.New(core))

			adapter.Narrate(screenplay.Event{
				Kind:    screenplay.KindAside,
				Level:   level,
				Phase:   screenplay.PhaseBegin,
				Message: "an aside",
			})

			entry := single(t, logs)
			assert.Equal(t, want, entry.Level)
		}
	})

	t.Run("falls back on the global logger", func(t *testing.T) {
		t.Parallel()

		adapter := zapadapter.New(nil)

		assert.NotPanics(t, func() {
			adapter.Narrate(screenplay.Event{Kind: screenplay.KindAside, Message: "an aside"})
		})
	})
}

func single(t *testing.T, logs *observer.ObservedLogs) observer.LoggedEntry {
	t.Helper()

	entries := logs.All()
	require.Len(t, entries, 1)

	return entries[0]
}
