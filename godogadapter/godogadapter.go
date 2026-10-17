// Package godogadapter renders go-screenplay narration through
// github.com/cucumber/godog, so the narrative is logged on the step of the
// scenario that produced it.
package godogadapter

import (
	"context"
	"strings"

	"github.com/cucumber/godog"

	"github.com/grandper/go-screenplay/v2/screenplay"
)

// Adapter writes narration to the log of a godog scenario.
type Adapter struct {
	ctx    context.Context
	indent string
}

// New returns an Adapter writing to the log of the scenario that carries the
// given context, as godog.Logf does: the testing.T of the step when the
// scenario runs under go test, the standard output otherwise. If ctx is nil,
// context.Background() is used. Each nesting level is indented by four
// spaces.
func New(ctx context.Context) *Adapter {
	if ctx == nil {
		ctx = context.Background()
	}

	return &Adapter{ctx: ctx, indent: "    "}
}

// WithIndent overrides the per-level indentation string.
func (a *Adapter) WithIndent(indent string) *Adapter {
	a.indent = indent
	return a
}

// Narrate implements screenplay.Adapter.
func (a *Adapter) Narrate(event screenplay.Event) {
	// Announce steps on the way in; only report the way out when it failed or
	// produced an answer, to avoid doubling every line.
	if event.Phase == screenplay.PhaseEnd && event.Err == nil && event.Answer == nil {
		return
	}

	pad := strings.Repeat(a.indent, event.Depth)

	switch {
	case event.Err != nil:
		godog.Logf(a.ctx, "%s✗ %s: %v", pad, event.Message, event.Err)
	case event.Answer != nil:
		godog.Logf(a.ctx, "%s=> %v", pad, event.Answer)
	default:
		godog.Logf(a.ctx, "%s%s", pad, event.Message)
	}
}

// Adapter implements the screenplay.Adapter interface.
var _ screenplay.Adapter = (*Adapter)(nil)
