package zerologadapter_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-narrators/zerologadapter"
)

func TestZerologAdapter(t *testing.T) {
	t.Parallel()

	t.Run("emits a structured event for a beat", func(t *testing.T) {
		t.Parallel()

		var buffer bytes.Buffer
		logger := zerolog.New(&buffer)
		adapter := zerologadapter.New(&logger)

		adapter.Narrate(screenplay.Event{
			Kind:    screenplay.KindBeat,
			Level:   screenplay.LevelInfo,
			Phase:   screenplay.PhaseBegin,
			Depth:   1,
			Actor:   "Wanda",
			Message: "Wanda sends a GET request",
		})

		record := decode(t, buffer.Bytes())
		assert.Equal(t, "Wanda sends a GET request", record["message"])
		assert.Equal(t, "info", record["level"])
		assert.Equal(t, "beat", record["kind"])
		assert.Equal(t, "Wanda", record["actor"])
		assert.InDelta(t, 1, record["depth"], 0)
	})

	t.Run("stays silent on a successful end phase", func(t *testing.T) {
		t.Parallel()

		var buffer bytes.Buffer
		logger := zerolog.New(&buffer)
		adapter := zerologadapter.New(&logger)

		adapter.Narrate(screenplay.Event{Kind: screenplay.KindBeat, Phase: screenplay.PhaseEnd})

		assert.Empty(t, buffer.String())
	})

	t.Run("records an answer", func(t *testing.T) {
		t.Parallel()

		var buffer bytes.Buffer
		logger := zerolog.New(&buffer)
		adapter := zerologadapter.New(&logger)

		adapter.Narrate(screenplay.Event{
			Kind:   screenplay.KindBeat,
			Level:  screenplay.LevelInfo,
			Phase:  screenplay.PhaseEnd,
			Answer: 200,
		})

		record := decode(t, buffer.Bytes())
		assert.InDelta(t, 200, record["answer"], 0)
	})

	t.Run("records a failure at the error level", func(t *testing.T) {
		t.Parallel()

		var buffer bytes.Buffer
		logger := zerolog.New(&buffer)
		adapter := zerologadapter.New(&logger)

		adapter.Narrate(screenplay.Event{
			Kind:    screenplay.KindBeat,
			Level:   screenplay.LevelError,
			Phase:   screenplay.PhaseEnd,
			Message: "Wanda sends a GET request",
			Err:     errors.New("connection refused"),
		})

		record := decode(t, buffer.Bytes())
		assert.Equal(t, "error", record["level"])
		assert.Equal(t, "connection refused", record["error"])
	})

	t.Run("maps every screenplay level onto its zerolog level", func(t *testing.T) {
		t.Parallel()

		levels := map[screenplay.Level]string{
			screenplay.LevelDebug: "debug",
			screenplay.LevelInfo:  "info",
			screenplay.LevelWarn:  "warn",
			screenplay.LevelError: "error",
			screenplay.Level(42):  "info",
		}

		for level, want := range levels {
			var buffer bytes.Buffer
			logger := zerolog.New(&buffer).Level(zerolog.DebugLevel)
			adapter := zerologadapter.New(&logger)

			adapter.Narrate(screenplay.Event{
				Kind:    screenplay.KindAside,
				Level:   level,
				Phase:   screenplay.PhaseBegin,
				Message: "an aside",
			})

			record := decode(t, buffer.Bytes())
			assert.Equal(t, want, record["level"])
		}
	})

	t.Run("falls back on the global logger", func(t *testing.T) {
		t.Parallel()

		adapter := zerologadapter.New(nil)

		assert.NotPanics(t, func() {
			adapter.Narrate(screenplay.Event{Kind: screenplay.KindAside, Level: screenplay.LevelDebug})
		})
	})
}

func decode(t *testing.T, data []byte) map[string]any {
	t.Helper()

	var record map[string]any
	require.NoError(t, json.Unmarshal(data, &record))

	return record
}
