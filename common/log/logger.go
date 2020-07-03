package log

import (
	"fmt"

	"github.com/inconshreveable/log15"
)

var (
	defaultLogger = Logger(&DefaultLogger{Logger: New("hotstuff", "logger")})
)

type Logger interface {
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})

	Error(v ...interface{})
	Errorf(format string, v ...interface{})

	Info(v ...interface{})
	Infof(format string, v ...interface{})

	Warning(v ...interface{})
	Warningf(format string, v ...interface{})

	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})

	Panic(v ...interface{})
	Panicf(format string, v ...interface{})
}

func SetLogger(l Logger) {
	defaultLogger = l
}

func GetLogger(ctx ...interface{}) Logger {
	if len(ctx) == 0 {
		return defaultLogger
	}
	return &DefaultLogger{Logger: New(ctx...)}
}

// DefaultLogger is a default implementation of the Logger interface.
type DefaultLogger struct {
	log15.Logger
}

func (l *DefaultLogger) Debug(v ...interface{}) {
	if len(v) > 0 {
		ctx := v[1:]
		l.Logger.Debug(v[0].(string), ctx...)
	}
}

func (l *DefaultLogger) Debugf(format string, v ...interface{}) {
	l.Logger.Debug(fmt.Sprintf(format, v...))
}

func (l *DefaultLogger) Error(v ...interface{}) {
	if len(v) > 0 {
		ctx := v[1:]
		l.Logger.Error(v[0].(string), ctx...)
	}
}

func (l *DefaultLogger) Errorf(format string, v ...interface{}) {
	l.Logger.Error(fmt.Sprintf(format, v...))
}

func (l *DefaultLogger) Info(v ...interface{}) {
	if len(v) > 0 {
		ctx := v[1:]
		l.Logger.Info(v[0].(string), ctx...)
	}
}

func (l *DefaultLogger) Infof(format string, v ...interface{}) {
	l.Logger.Info(fmt.Sprintf(format, v...))
}

func (l *DefaultLogger) Warning(v ...interface{}) {
	if len(v) > 0 {
		ctx := v[1:]
		l.Logger.Warn(v[0].(string), ctx...)
	}
}

func (l *DefaultLogger) Warningf(format string, v ...interface{}) {
	l.Logger.Warn(fmt.Sprintf(format, v...))
}

func (l *DefaultLogger) Fatal(v ...interface{}) {
	if len(v) > 0 {
		ctx := v[1:]
		l.Logger.Crit(v[0].(string), ctx...)
	}
}

func (l *DefaultLogger) Fatalf(format string, v ...interface{}) {
	l.Logger.Crit(fmt.Sprintf(format, v...))
}

func (l *DefaultLogger) Panic(v ...interface{}) {
	panic(fmt.Sprint(v...))
}

func (l *DefaultLogger) Panicf(format string, v ...interface{}) {
	panic(fmt.Sprintf(format, v...))
}
