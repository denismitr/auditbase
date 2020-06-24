package logger

import (
	"bytes"
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
	SQL(query string, args []interface{})
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
	_ = l.errLogger.Output(2, fmt.Sprintf("\nERROR in ["+l.env+"] env: %s", err.Error()))
}

// Debugf - prints debug message with params
func (l *StdLogger) Debugf(format string, args ...interface{}) {
	if l.env == Prod {
		return
	}

	_ = l.debugLogger.Output(2, fmt.Sprintf("\nDEBUG in ["+l.env+"] "+format, args...))
}

func (l *StdLogger) SQL(query string, args []interface{}) {
	if l.env == Prod {
		return
	}

	var buf = &bytes.Buffer{}

	buf.WriteString("\nSQL in [")
	buf.WriteString(l.env)
	buf.WriteString("]: ")
	buf.WriteString(query)
	buf.WriteString(" \nArgs: ")

	for i := range args {
		if i + 1 < len(args) {
			buf.WriteString(fmt.Sprintf("{%#v}, ", args[i]))
		} else {
			buf.WriteString(fmt.Sprintf("{%#v}", args[i]))
		}
	}

	_ = l.debugLogger.Output(2, buf.String())
}