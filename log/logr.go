package log

import (
	"github.com/go-logr/logr"
	"go.uber.org/zap"
)

type logrLogger struct {
	enabled bool
	s       *zap.SugaredLogger
}

func (l *logrLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	l.s.With(zap.Error(err)).Errorw(msg, keysAndValues...)
}

func (l *logrLogger) Info(msg string, keysAndValues ...interface{}) {
	if l.enabled {
		l.s.Infow(msg, keysAndValues...)
	}
}

func (l *logrLogger) Enabled() bool { return l.enabled }

func (l *logrLogger) V(level int) logr.InfoLogger {
	return &logrLogger{int(g.verbosity.get()) >= level, l.s}
}

func (l *logrLogger) WithName(name string) logr.Logger {
	return &logrLogger{true, l.s.Named(name)}
}

func (l *logrLogger) WithValues(keysAndValues ...interface{}) logr.Logger {
	return &logrLogger{true, l.s.With(keysAndValues...)}
}

func NewLogrLogger(ns ...string) logr.Logger {
	s := SDepth(1)
	for _, n := range ns {
		s = s.Named(n)
	}
	return &logrLogger{true, s}
}
