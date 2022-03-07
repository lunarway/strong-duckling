package test

import (
	"io"
	"testing"

	"github.com/go-kit/kit/log"
)

func NewLogger(t *testing.T) log.Logger {
	return log.NewLogger(&logger{t})
}

var _ io.Writer = &logger{}

// logger is an io.Writer used for reporting logs to the test runner.
type logger struct {
	t *testing.T
}

func (l *logger) Write(p []byte) (int, error) {
	l.t.Logf("%s", p)
	return len(p), nil
}
