package logger

import (
	"fmt"
	"log"
	"os"
)

const Dev = "dev"
const Prod = "prod"

// Logger - provides small interface for logging errors and debuging
type Logger interface {
	Error(err error)
	Debugf(format string, args ...interface{})
}

// NewStdoutLogger creates StdLogger that uses stderr and stdout for logging
func NewStdoutLogger(env, namespace string) *StdLogger {
	debugLogger := log.New(os.Stdout, namespace+" ", log.Ldate|log.Ltime|log.Lshortfile)
	errLogger := log.New(os.Stderr, namespace+" ", log.Ldate|log.Ltime|log.Lshortfile)

	return &StdLogger{
		debugLogger: debugLogger,
		errLogger:   errLogger,
		env:         env,
	}
}

// StdLogger uses stderr and stdout for logging
type StdLogger struct {
	errLogger   *log.Logger
	debugLogger *log.Logger
	env         string
}

// Error - logs error message using error logger driver
func (l *StdLogger) Error(err error) {
	l.errLogger.Output(2, fmt.Sprintf("\nERROR in ["+l.env+"] env: %s", err.Error()))
}

// Debugf - prints debug message with params
func (l *StdLogger) Debugf(format string, args ...interface{}) {
	if l.env == Prod {
		return
	}

	l.debugLogger.Output(2, fmt.Sprintf("\nDEBUG in ["+l.env+"] "+format, args...))
}
