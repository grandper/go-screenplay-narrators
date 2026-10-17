// Package testingadapter renders go-screenplay narration on the log of a test,
// so the narrative shows up in go test -v next to the failures of testify or of
// any other assertion library.
package testingadapter

import (
	"strings"
	"testing"

	"github.com/grandper/go-screenplay/v2/screenplay"
)

// Adapter writes narration to the log of a testing.TB.
type Adapter struct {
	t      testing.TB
	indent string
}

// New returns an Adapter writing to the log of the given test: a *testing.T, a
// *testing.B, a *testing.F, or the T() of a testify suite. If t is nil, the
// adapter stays silent. Each nesting level is indented by four spaces.
func New(t testing.TB) *Adapter {
	return &Adapter{t: t, indent: "    "}
}

// WithIndent overrides the per-level indentation string.
func (a *Adapter) WithIndent(indent string) *Adapter {
	a.indent = indent
	return a
}

// Narrate implements screenplay.Adapter.
func (a *Adapter) Narrate(event screenplay.Event) {
	if a.t == nil {
		return
	}

	// Announce steps on the way in; only report the way out when it failed or
	// produced an answer, to avoid doubling every line.
	if event.Phase == screenplay.PhaseEnd && event.Err == nil && event.Answer == nil {
		return
	}

	pad := strings.Repeat(a.indent, event.Depth)

	switch {
	case event.Err != nil:
		a.t.Logf("%s✗ %s: %v", pad, event.Message, event.Err)
	case event.Answer != nil:
		a.t.Logf("%s=> %v", pad, event.Answer)
	default:
		a.t.Logf("%s%s", pad, event.Message)
	}
}

// Adapter implements the screenplay.Adapter interface.
var _ screenplay.Adapter = (*Adapter)(nil)
