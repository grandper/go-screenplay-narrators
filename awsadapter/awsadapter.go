// Package awsadapter renders go-screenplay narration through the logger of the
// AWS SDK for Go v2 (github.com/aws/smithy-go/logging), so the narrative lands
// in the same place as the SDK's own logs.
package awsadapter

import (
	"os"
	"strings"

	"github.com/aws/smithy-go/logging"

	"github.com/grandper/go-screenplay/v2/screenplay"
)

// Adapter writes narration to a logging.Logger.
type Adapter struct {
	logger logging.Logger
	indent string
}

// New returns an Adapter writing to the given logger, typically the Logger of
// an aws.Config. If logger is nil, the SDK's default logger is used: a
// logging.StandardLogger writing to os.Stderr. Each nesting level is indented
// by four spaces.
func New(logger logging.Logger) *Adapter {
	if logger == nil {
		logger = logging.NewStandardLogger(os.Stderr)
	}

	return &Adapter{logger: logger, indent: "    "}
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
	classification := mapLevel(event.Level)

	switch {
	case event.Err != nil:
		a.logger.Logf(classification, "%s✗ %s: %v", pad, event.Message, event.Err)
	case event.Answer != nil:
		a.logger.Logf(classification, "%s=> %v", pad, event.Answer)
	default:
		a.logger.Logf(classification, "%s%s", pad, event.Message)
	}
}

// mapLevel maps a screenplay.Level onto one of the two classifications of the
// SDK: a warning or an error is a warning, anything else is debugging output.
func mapLevel(level screenplay.Level) logging.Classification {
	switch level {
	case screenplay.LevelDebug, screenplay.LevelInfo:
		return logging.Debug
	case screenplay.LevelWarn, screenplay.LevelError:
		return logging.Warn
	default:
		return logging.Debug
	}
}

// Adapter implements the screenplay.Adapter interface.
var _ screenplay.Adapter = (*Adapter)(nil)
