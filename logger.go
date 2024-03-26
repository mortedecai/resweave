package resweave

import (
	"fmt"

	"go.uber.org/zap"
)

type LogHolder interface {
	Logger() *zap.SugaredLogger
	SetLogger(logger *zap.SugaredLogger, recursive bool)
	LoggerName() string

	Infow(msg string, args ...interface{})
	Errorw(msg string, args ...interface{})
	Debugw(msg string, args ...interface{})
}

type RecurserFunc func(l *zap.SugaredLogger)

type logholder struct {
	logger   *zap.SugaredLogger
	recurser RecurserFunc
	name     string
}

func NewLogholder(name string, r RecurserFunc) *logholder {
	return &logholder{name: name, logger: nil, recurser: r}
}

func (l *logholder) Logger() *zap.SugaredLogger {
	return l.logger
}

func (l *logholder) setLoggerName(logger *zap.SugaredLogger) *zap.SugaredLogger {
	if logger == nil {
		fmt.Println("logger was nil")
		return logger
	}
	return logger.Named(l.name)
}

func (l *logholder) SetLogger(logger *zap.SugaredLogger, recursive bool) {
	l.logger = l.setLoggerName(logger)
	if recursive && l.recurser != nil {
		l.recurser(l.logger)
	}
}

func (l *logholder) LoggerName() string {
	return l.name
}

func (l *logholder) Infow(msg string, args ...interface{}) {
	if l.Logger() != nil {
		l.Logger().Infow(msg, args...)
	}
}

func (l *logholder) Errorw(msg string, args ...interface{}) {
	if l.Logger() != nil {
		l.Logger().Errorw(msg, args...)
	}
}

func (l *logholder) Debugw(msg string, args ...interface{}) {
	if l.Logger() != nil {
		l.Logger().Debugw(msg, args...)
	}
}
