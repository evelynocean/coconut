package coconutLog

import (
	"github.com/sirupsen/logrus"
)

type Logger struct {
	Log *logrus.Logger
}

type Entry struct {
	Entry *logrus.Entry
}

func New() *Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})

	return &Logger{
		Log: log,
	}
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Log.Printf(format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Log.Errorf(format, args...)
}

func (l *Logger) WithFields(fields map[string]interface{}) (e *Entry) {
	e = &Entry{
		Entry: l.Log.WithFields(fields),
	}

	return e
}

func (l *Logger) WithError(err error) (e *Entry) {
	e = &Entry{
		Entry: l.Log.WithError(err),
	}

	return e
}

func (e *Entry) Debugf(format string, args ...interface{}) {
	e.Entry.Printf(format, args...)
}

func (e *Entry) Errorf(format string, args ...interface{}) {
	e.Entry.Errorf(format, args...)
}
