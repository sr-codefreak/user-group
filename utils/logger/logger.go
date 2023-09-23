package logger

import (
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

type myLogger struct {
	sync.Mutex
	l *logrus.Logger
}

var lo = myLogger{}

func NewLogger() *logrus.Logger {

	logFile, err := os.Create("log.json")
	multi := io.MultiWriter(os.Stdout)
	if err == nil {
		multi = io.MultiWriter(logFile, os.Stdout)
	}

	if lo.l == nil {
		lo.Lock()
		defer lo.Unlock()
		if lo.l == nil {

			l := &logrus.Logger{}
			l.SetFormatter(&logrus.TextFormatter{})

			l.SetOutput(multi)
			l.SetLevel(logrus.InfoLevel)
			lo.l = l
		}
	}
	return lo.l
}

func GetLogger() *logrus.Logger {
	if lo.l == nil {
		_ = NewLogger()
	}
	return lo.l
}
