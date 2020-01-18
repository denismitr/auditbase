package utils

import (
	"log"
	"os"
)

const Dev = "dev"
const Prod = "prod"

type Logger interface {
	Error(err error)
	Debugf(format string, args ...interface{})
}

func NewStdoutLogger(env, namespace string) *StdLogger {
	logger := log.New(os.Stdout, namespace, log.Ldate|log.Ltime|log.Lshortfile)
	return &StdLogger{
		l:   logger,
		env: env,
	}
}

type StdLogger struct {
	l   *log.Logger
	env string
}

func (l *StdLogger) Error(err error) {
	l.l.Printf("error in [%s] env: ", err.Error())
}

// Debugf - prints debug message with params
func (l *StdLogger) Debugf(format string, args ...interface{}) {
	if l.env == Prod {
		return
	}

	var allArgs []interface{}
	allArgs = append(allArgs, l.env)
	allArgs = append(allArgs, args...)
	l.l.Printf("debug info in [%s] env: "+format, allArgs)
}
