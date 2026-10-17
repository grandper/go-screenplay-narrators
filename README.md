# go-screenplay-narrators

[![Test](https://github.com/grandper/go-screenplay-narrators/actions/workflows/go-test.yml/badge.svg)](https://github.com/grandper/go-screenplay-narrators/actions/workflows/go-test.yml)
[![Lint](https://github.com/grandper/go-screenplay-narrators/actions/workflows/go-lint.yml/badge.svg)](https://github.com/grandper/go-screenplay-narrators/actions/workflows/go-lint.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/grandper/go-screenplay-narrators.svg)](https://pkg.go.dev/search?q=github.com%2Fgrandper%2Fgo-screenplay-narrators)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

`go-screenplay-narrators` is a collection of packages that renders the narration of
[go-screenplay](https://github.com/grandper/go-screenplay) through the logging
libraries your project already uses.

**Main features:**

- A narrator adapter for [zap](https://github.com/uber-go/zap), [zerolog](https://github.com/rs/zerolog)
  and [logrus](https://github.com/sirupsen/logrus), which keep the actor, the kind, the depth and
  the error as fields.
- A narrator adapter for [godog](https://github.com/cucumber/godog), which logs the narrative on the
  step of the scenario that produced it.
- A narrator adapter for `testing.TB`, which logs the narrative in `go test -v`, next to the failures
  of [testify](https://github.com/stretchr/testify) or of any other assertion library.
- A narrator adapter for the [logger of the AWS SDK for Go v2](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-logging.html),
  which writes the narrative next to the SDK's own logs.
- One Go module per adapter: importing one of them never pulls the dependencies of the others.

## Installation

Every adapter is its own Go module. Install only the ones you use:
```bash
go get github.com/grandper/go-screenplay-narrators/zapadapter
go get github.com/grandper/go-screenplay-narrators/zerologadapter
go get github.com/grandper/go-screenplay-narrators/logrusadapter
go get github.com/grandper/go-screenplay-narrators/testingadapter
go get github.com/grandper/go-screenplay-narrators/godogadapter
go get github.com/grandper/go-screenplay-narrators/awsadapter
```
The library requires Go 1.27 or later.

## What is narration?
As actors perform actions and ask questions, go-screenplay can emit a readable,
hierarchical narrative of what is happening:
```
Wanda attempts to request the list of birds
    Wanda attempts to set the header 'Accept' to 'application/json'
    Wanda attempts to send a GET request to /birds
Wanda should see that the status code should be equal to 200
Wanda asks for the status code of the last response
    => 200
```
Narration is **backend independent**: go-screenplay does not force a logging library
on you. It exposes a small `screenplay.Adapter` interface, and ships two reference
adapters, one over the standard `log` package and one over `log/slog`.

Your project probably already logs through something else: zap in the services,
testify in the unit tests, godog in the acceptance tests, the logger of the AWS SDK
next to the calls to AWS. This is where `go-screenplay-narrators` comes in. Each
package of this repository is an adapter that turns every narrated step into a line of
that library, so that the narrative of the actors lands where you already read your logs.

By default **nothing is narrated**: attaching an adapter is entirely opt-in.

## Attaching an adapter
An adapter implements the one-method `screenplay.Adapter` interface of go-screenplay.
Attach it to a `Narrator`, and give the narrator to the `Production`, exactly as you do
with the reference adapters of go-screenplay (`narrator/logadapter` and `narrator/slogadapter`):
```go
import (
    "go.uber.org/zap"

    "github.com/grandper/go-screenplay-narrators/zapadapter"
    "github.com/grandper/go-screenplay/v2/screenplay"
)

logger, _ := zap.NewProduction()
narrator := screenplay.NewNarrator(zapadapter.New(logger))
production := screenplay.NewProduction(screenplay.WithNarrator(narrator))

stage := production.SetTheStage(screenplay.CastWhereEveryoneCan(MakeHTTPRequests))
wanda := stage.ActorNamed("Wanda") // already narrating
```
Every adapter follows the same rules:

- a step is announced on the way in;
- the way out is only reported when the step failed or a question produced an answer, so that
  every line is not doubled;
- a failure is reported at the error level of the logging library;
- a nil logger falls back on the default logger of the logging library (the `testing` adapter,
  which has no default, stays silent).

You can attach several adapters at once; every event is delivered to each of them:
```go
narrator := screenplay.NewNarrator(zapadapter.New(logger))
narrator.AttachAdapter(awsadapter.New(cfg.Logger))
```

## zap — [`zapadapter`](zapadapter)
The `zap` adapter emits one structured entry per step, and keeps the actor, the kind, the
depth, the answer and the error as fields:
```go
logger, _ := zap.NewProduction()
narrator := screenplay.NewNarrator(zapadapter.New(logger))
```
When Wanda fails to send a request, the entry reads:
```json
{"level":"error","ts":1792267029.36,"caller":"zapadapter/zapadapter.go:51","msg":"Wanda attempts to send a GET request to /birds","kind":"beat","depth":1,"actor":"Wanda","error":"dial tcp 127.0.0.1:8080: connect: connection refused"}
```
`zapadapter.New(nil)` writes to `zap.L()`, the global logger.

## zerolog — [`zerologadapter`](zerologadapter)
The `zerolog` adapter emits one structured event per step, with the same fields:
```go
logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
narrator := screenplay.NewNarrator(zerologadapter.New(&logger))
```
When Wanda fails to send a request, the event reads:
```json
{"level":"error","kind":"beat","depth":1,"actor":"Wanda","error":"dial tcp 127.0.0.1:8080: connect: connection refused","time":"2026-10-17T21:57:09+02:00","message":"Wanda attempts to send a GET request to /birds"}
```
`zerologadapter.New(nil)` writes to `log.Logger`, the global logger of `github.com/rs/zerolog/log`.

## logrus — [`logrusadapter`](logrusadapter)
The `logrus` adapter emits one entry per step, with the same fields. The error is stored
under `logrus.ErrorKey`, as `WithError` does:
```go
logger := logrus.New()
logger.SetFormatter(&logrus.JSONFormatter{})
narrator := screenplay.NewNarrator(logrusadapter.New(logger))
```
When Wanda fails to send a request, the entry reads:
```json
{"actor":"Wanda","depth":1,"error":"dial tcp 127.0.0.1:8080: connect: connection refused","kind":"beat","level":"error","msg":"Wanda attempts to send a GET request to /birds","time":"2026-10-17T21:57:09+02:00"}
```
`logrusadapter.New(nil)` writes to `logrus.StandardLogger()`.

## testing and testify — [`testingadapter`](testingadapter)
Testify has no logger of its own: its failures are reported on the `testing.T` of the test.
The `testing` adapter logs the narrative on that same `testing.TB`, so the narrative and the
failures read in one place:
```go
func TestBirds(t *testing.T) {
    narrator := screenplay.NewNarrator(testingadapter.New(t))
    production := screenplay.NewProduction(screenplay.WithNarrator(narrator))
    wanda := production.SetTheStage(screenplay.CastWhereEveryoneCan(MakeHTTPRequests)).ActorNamed("Wanda")
    ...
}
```
In a testify suite, pass the `T()` of the suite:
```go
func (s *BirdsSuite) SetupTest() {
    narrator := screenplay.NewNarrator(testingadapter.New(s.T()))
    ...
}
```
Run with `go test -v`, the test reports the narrative, indented by depth:
```
=== RUN   TestBirds
    testingadapter.go:52: Wanda attempts to request the list of birds
    testingadapter.go:52:     Wanda attempts to send a GET request to /birds
```
The adapter holds the `testing.TB` it was given: build it inside the test that narrates, and
never narrate once that test has returned. Change the indentation with `WithIndent`:
```go
adapter := testingadapter.New(t).WithIndent("  ")
```

## godog — [`godogadapter`](godogadapter)
The `godog` adapter logs the narrative on the step being run, as `godog.Logf` does: on the
`testing.T` of the step when the suite runs under `go test`, on the standard output otherwise.
The adapter is built from the context of the scenario, so build the narrator in a `Before`
hook and keep it in the context:
```go
type stageKey struct{}

func InitializeScenario(scenario *godog.ScenarioContext) {
    scenario.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
        narrator := screenplay.NewNarrator(godogadapter.New(ctx))
        production := screenplay.NewProduction(screenplay.WithNarrator(narrator))
        stage := production.SetTheStage(screenplay.CastWhereEveryoneCan(MakeHTTPRequests))

        return context.WithValue(ctx, stageKey{}, stage), nil
    })
}
```
Run with `go test -v`, the step reports the narrative, indented by depth:
```
Wanda attempts to request the list of birds
    Wanda attempts to send a GET request to /birds
```
The lines are also returned by `godog.LoggedMessages(ctx)`. Change the indentation with
`WithIndent`:
```go
adapter := godogadapter.New(ctx).WithIndent("  ")
```

## AWS SDK for Go v2 — [`awsadapter`](awsadapter)
The `aws` adapter writes the narrative to a `logging.Logger` of `github.com/aws/smithy-go`, the
logger the AWS SDK for Go v2 uses. Pass it the `Logger` of your `aws.Config`, and the narrative
lands next to the SDK's own logs:
```go
cfg, err := config.LoadDefaultConfig(ctx, config.WithClientLogMode(aws.LogRetries|aws.LogRequest))
if err != nil {
    return err
}

narrator := screenplay.NewNarrator(awsadapter.New(cfg.Logger))
```
The SDK only knows two classifications: a warning or an error is logged as `WARN`, everything
else as `DEBUG`. With the SDK's standard logger, the lines read:
```
SDK 2026/10/17 21:57:09 DEBUG Wanda attempts to request the list of birds
SDK 2026/10/17 21:57:09 DEBUG     Wanda attempts to send a GET request to /birds
```
`awsadapter.New(nil)` writes to the SDK's default logger, `logging.NewStandardLogger(os.Stderr)`.
Change the indentation with `WithIndent`:
```go
adapter := awsadapter.New(cfg.Logger).WithIndent("  ")
```

## Keeping your dependencies to a minimum
Each adapter lives in its own module, with its own `go.mod`. Importing `zapadapter` adds `zap`,
go-screenplay and their own dependencies to your `go.mod`, and nothing else: zerolog, logrus,
godog and the AWS SDK stay out of your build.

## License
Licensed under MIT License.
