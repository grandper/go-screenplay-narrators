package godogadapter_test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grandper/go-screenplay/v2/screenplay"

	"github.com/grandper/go-screenplay-narrators/godogadapter"
)

func TestGodogAdapter(t *testing.T) {
	t.Parallel()

	t.Run("announces a beat on the way in", func(t *testing.T) {
		t.Parallel()

		logged := narrateInAScenario(t, func(adapter *godogadapter.Adapter) {
			adapter.Narrate(screenplay.Event{
				Kind:    screenplay.KindBeat,
				Phase:   screenplay.PhaseBegin,
				Message: "Wanda sends a GET request to /birds",
			})
		})

		assert.Equal(t, []string{"Wanda sends a GET request to /birds"}, logged)
	})

	t.Run("stays silent on a successful end phase", func(t *testing.T) {
		t.Parallel()

		logged := narrateInAScenario(t, func(adapter *godogadapter.Adapter) {
			adapter.Narrate(screenplay.Event{Kind: screenplay.KindBeat, Phase: screenplay.PhaseEnd})
		})

		assert.Empty(t, logged)
	})

	t.Run("indents according to the depth", func(t *testing.T) {
		t.Parallel()

		logged := narrateInAScenario(t, func(adapter *godogadapter.Adapter) {
			adapter.Narrate(screenplay.Event{
				Kind:    screenplay.KindBeat,
				Phase:   screenplay.PhaseBegin,
				Depth:   2,
				Message: "sets the Accept header",
			})
		})

		assert.Equal(t, []string{"        sets the Accept header"}, logged)
	})

	t.Run("renders an answer", func(t *testing.T) {
		t.Parallel()

		logged := narrateInAScenario(t, func(adapter *godogadapter.Adapter) {
			adapter.Narrate(screenplay.Event{
				Kind:   screenplay.KindBeat,
				Phase:  screenplay.PhaseEnd,
				Answer: 200,
			})
		})

		assert.Equal(t, []string{"=> 200"}, logged)
	})

	t.Run("renders a failure", func(t *testing.T) {
		t.Parallel()

		logged := narrateInAScenario(t, func(adapter *godogadapter.Adapter) {
			adapter.Narrate(screenplay.Event{
				Kind:    screenplay.KindBeat,
				Phase:   screenplay.PhaseEnd,
				Message: "Wanda sends a GET request",
				Err:     errors.New("connection refused"),
			})
		})

		assert.Equal(t, []string{"✗ Wanda sends a GET request: connection refused"}, logged)
	})

	t.Run("never reads the message as a format", func(t *testing.T) {
		t.Parallel()

		logged := narrateInAScenario(t, func(adapter *godogadapter.Adapter) {
			adapter.Narrate(screenplay.Event{
				Kind:    screenplay.KindBeat,
				Phase:   screenplay.PhaseBegin,
				Message: "Wanda enters the text 100%d",
			})
		})

		assert.Equal(t, []string{"Wanda enters the text 100%d"}, logged)
	})

	t.Run("honors a custom indentation", func(t *testing.T) {
		t.Parallel()

		logged := narrateInAScenario(t, func(adapter *godogadapter.Adapter) {
			adapter.WithIndent(">>").Narrate(screenplay.Event{
				Kind:    screenplay.KindBeat,
				Phase:   screenplay.PhaseBegin,
				Depth:   1,
				Message: "sets the Accept header",
			})
		})

		assert.Equal(t, []string{">>sets the Accept header"}, logged)
	})
}

// feature is a one-step scenario, so a test can narrate from inside a godog
// step and read what the step logged.
const feature = `Feature: narration
  Scenario: Wanda requests the list of birds
    When Wanda narrates
`

// narrateInAScenario runs a godog scenario whose only step narrates through an
// adapter built from the step's context, and returns what the step logged.
func narrateInAScenario(t *testing.T, narrate func(adapter *godogadapter.Adapter)) []string {
	t.Helper()

	var logged []string

	suite := godog.TestSuite{
		ScenarioInitializer: func(scenario *godog.ScenarioContext) {
			scenario.Step(`^Wanda narrates$`, func(ctx context.Context) {
				narrate(godogadapter.New(ctx))
				logged = godog.LoggedMessages(ctx)
			})
		},
		Options: &godog.Options{
			Format:          "progress",
			Output:          io.Discard,
			FeatureContents: []godog.Feature{{Name: "narration.feature", Contents: []byte(feature)}},
			TestingT:        t,
		},
	}

	require.Zero(t, suite.Run())

	return logged
}
