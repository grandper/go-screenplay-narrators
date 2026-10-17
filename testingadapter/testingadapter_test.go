package testingadapter_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-narrators/testingadapter"
)

func TestTestingAdapter(t *testing.T) {
	t.Parallel()

	t.Run("announces a beat on the way in", func(t *testing.T) {
		t.Parallel()

		test := &recordingTest{TB: t}
		adapter := testingadapter.New(test)

		adapter.Narrate(screenplay.Event{
			Kind:    screenplay.KindBeat,
			Phase:   screenplay.PhaseBegin,
			Message: "Wanda sends a GET request to /birds",
		})

		assert.Equal(t, []string{"Wanda sends a GET request to /birds"}, test.logged)
	})

	t.Run("stays silent on a successful end phase", func(t *testing.T) {
		t.Parallel()

		test := &recordingTest{TB: t}
		adapter := testingadapter.New(test)

		adapter.Narrate(screenplay.Event{Kind: screenplay.KindBeat, Phase: screenplay.PhaseEnd})

		assert.Empty(t, test.logged)
	})

	t.Run("indents according to the depth", func(t *testing.T) {
		t.Parallel()

		test := &recordingTest{TB: t}
		adapter := testingadapter.New(test)

		adapter.Narrate(screenplay.Event{
			Kind:    screenplay.KindBeat,
			Phase:   screenplay.PhaseBegin,
			Depth:   2,
			Message: "sets the Accept header",
		})

		assert.Equal(t, []string{"        sets the Accept header"}, test.logged)
	})

	t.Run("renders an answer", func(t *testing.T) {
		t.Parallel()

		test := &recordingTest{TB: t}
		adapter := testingadapter.New(test)

		adapter.Narrate(screenplay.Event{
			Kind:   screenplay.KindBeat,
			Phase:  screenplay.PhaseEnd,
			Answer: 200,
		})

		assert.Equal(t, []string{"=> 200"}, test.logged)
	})

	t.Run("renders a failure", func(t *testing.T) {
		t.Parallel()

		test := &recordingTest{TB: t}
		adapter := testingadapter.New(test)

		adapter.Narrate(screenplay.Event{
			Kind:    screenplay.KindBeat,
			Phase:   screenplay.PhaseEnd,
			Message: "Wanda sends a GET request",
			Err:     errors.New("connection refused"),
		})

		assert.Equal(t, []string{"✗ Wanda sends a GET request: connection refused"}, test.logged)
	})

	t.Run("never reads the message as a format", func(t *testing.T) {
		t.Parallel()

		test := &recordingTest{TB: t}
		adapter := testingadapter.New(test)

		adapter.Narrate(screenplay.Event{
			Kind:    screenplay.KindBeat,
			Phase:   screenplay.PhaseBegin,
			Message: "Wanda enters the text 100%d",
		})

		assert.Equal(t, []string{"Wanda enters the text 100%d"}, test.logged)
	})

	t.Run("honors a custom indentation", func(t *testing.T) {
		t.Parallel()

		test := &recordingTest{TB: t}
		adapter := testingadapter.New(test).WithIndent(">>")

		adapter.Narrate(screenplay.Event{
			Kind:    screenplay.KindBeat,
			Phase:   screenplay.PhaseBegin,
			Depth:   1,
			Message: "sets the Accept header",
		})

		assert.Equal(t, []string{">>sets the Accept header"}, test.logged)
	})

	t.Run("stays silent without a test", func(t *testing.T) {
		t.Parallel()

		adapter := testingadapter.New(nil)

		assert.NotPanics(t, func() {
			adapter.Narrate(screenplay.Event{Kind: screenplay.KindAside, Message: "an aside"})
		})
	})
}

// recordingTest is a testing.TB that records the lines it is asked to log, so
// a test can compare them.
type recordingTest struct {
	testing.TB

	logged []string
}

// Logf records the formatted line.
func (r *recordingTest) Logf(format string, args ...any) {
	r.logged = append(r.logged, fmt.Sprintf(format, args...))
}
