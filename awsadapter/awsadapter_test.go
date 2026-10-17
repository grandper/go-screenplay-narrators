package awsadapter_test

import (
	"bytes"
	"errors"
	"log"
	"testing"

	"github.com/aws/smithy-go/logging"
	"github.com/stretchr/testify/assert"

	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-narrators/awsadapter"
)

func TestAWSAdapter(t *testing.T) {
	t.Parallel()

	t.Run("announces a beat on the way in", func(t *testing.T) {
		t.Parallel()

		var buffer bytes.Buffer
		adapter := awsadapter.New(standardLogger(&buffer))

		adapter.Narrate(screenplay.Event{
			Kind:    screenplay.KindBeat,
			Level:   screenplay.LevelInfo,
			Phase:   screenplay.PhaseBegin,
			Message: "Wanda sends a GET request to /birds",
		})

		assert.Equal(t, "DEBUG Wanda sends a GET request to /birds\n", buffer.String())
	})

	t.Run("stays silent on a successful end phase", func(t *testing.T) {
		t.Parallel()

		var buffer bytes.Buffer
		adapter := awsadapter.New(standardLogger(&buffer))

		adapter.Narrate(screenplay.Event{Kind: screenplay.KindBeat, Phase: screenplay.PhaseEnd})

		assert.Empty(t, buffer.String())
	})

	t.Run("indents according to the depth", func(t *testing.T) {
		t.Parallel()

		var buffer bytes.Buffer
		adapter := awsadapter.New(standardLogger(&buffer))

		adapter.Narrate(screenplay.Event{
			Kind:    screenplay.KindBeat,
			Level:   screenplay.LevelInfo,
			Phase:   screenplay.PhaseBegin,
			Depth:   2,
			Message: "sets the Accept header",
		})

		assert.Equal(t, "DEBUG         sets the Accept header\n", buffer.String())
	})

	t.Run("renders an answer", func(t *testing.T) {
		t.Parallel()

		var buffer bytes.Buffer
		adapter := awsadapter.New(standardLogger(&buffer))

		adapter.Narrate(screenplay.Event{
			Kind:   screenplay.KindBeat,
			Level:  screenplay.LevelInfo,
			Phase:  screenplay.PhaseEnd,
			Answer: 200,
		})

		assert.Equal(t, "DEBUG => 200\n", buffer.String())
	})

	t.Run("renders a failure as a warning", func(t *testing.T) {
		t.Parallel()

		var buffer bytes.Buffer
		adapter := awsadapter.New(standardLogger(&buffer))

		adapter.Narrate(screenplay.Event{
			Kind:    screenplay.KindBeat,
			Level:   screenplay.LevelError,
			Phase:   screenplay.PhaseEnd,
			Message: "Wanda sends a GET request",
			Err:     errors.New("connection refused"),
		})

		assert.Equal(t, "WARN ✗ Wanda sends a GET request: connection refused\n", buffer.String())
	})

	t.Run("maps every screenplay level onto a classification", func(t *testing.T) {
		t.Parallel()

		levels := map[screenplay.Level]logging.Classification{
			screenplay.LevelDebug: logging.Debug,
			screenplay.LevelInfo:  logging.Debug,
			screenplay.LevelWarn:  logging.Warn,
			screenplay.LevelError: logging.Warn,
			screenplay.Level(42):  logging.Debug,
		}

		for level, want := range levels {
			var got logging.Classification
			adapter := awsadapter.New(logging.LoggerFunc(
				func(classification logging.Classification, _ string, _ ...any) { got = classification },
			))

			adapter.Narrate(screenplay.Event{
				Kind:    screenplay.KindAside,
				Level:   level,
				Phase:   screenplay.PhaseBegin,
				Message: "an aside",
			})

			assert.Equal(t, want, got)
		}
	})

	t.Run("never reads the message as a format", func(t *testing.T) {
		t.Parallel()

		var buffer bytes.Buffer
		adapter := awsadapter.New(standardLogger(&buffer))

		adapter.Narrate(screenplay.Event{
			Kind:    screenplay.KindBeat,
			Level:   screenplay.LevelInfo,
			Phase:   screenplay.PhaseBegin,
			Message: "Wanda enters the text 100%d",
		})

		assert.Equal(t, "DEBUG Wanda enters the text 100%d\n", buffer.String())
	})

	t.Run("honors a custom indentation", func(t *testing.T) {
		t.Parallel()

		var buffer bytes.Buffer
		adapter := awsadapter.New(standardLogger(&buffer)).WithIndent(">>")

		adapter.Narrate(screenplay.Event{
			Kind:    screenplay.KindBeat,
			Level:   screenplay.LevelInfo,
			Phase:   screenplay.PhaseBegin,
			Depth:   1,
			Message: "sets the Accept header",
		})

		assert.Equal(t, "DEBUG >>sets the Accept header\n", buffer.String())
	})
}

// standardLogger returns the SDK's standard logger without its prefix and
// timestamp, so a test can compare the lines it writes.
func standardLogger(buffer *bytes.Buffer) logging.Logger {
	return logging.StandardLogger{Logger: log.New(buffer, "", 0)}
}
