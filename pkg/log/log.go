// Package log initialize and configure a logrus logger.
package log

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

const (
	outputStdout = "stdout"
	outputStderr = "stderr"
	outputTest   = "test"
	outputSyslog = "syslog"
)

type LogrusLogger struct {
	logger *logrus.Logger
}

func NewLogger(logLevel string, logOutput string) (*LogrusLogger, error) {
	if logLevel == "" {
		logLevel = "info"
	}

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return nil, err
	}

	output, hook, err := getOutput(logOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to init logger: %v", err)
	}

	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}

	log := &logrus.Logger{
		Out:       output,
		Formatter: formatter,
		Hooks:     make(logrus.LevelHooks),
		Level:     level,
	}

	if logOutput == outputSyslog || logOutput == outputTest {
		log.Hooks.Add(hook)
	}

	return &LogrusLogger{logger: log}, nil
}

func (l *LogrusLogger) Infof(format string, args ...interface{}) {
	l.logger.Logf(logrus.InfoLevel, format, args...)
}

func (l *LogrusLogger) Errorf(format string, args ...interface{}) {
	l.logger.Logf(logrus.ErrorLevel, format, args...)
}

func (l *LogrusLogger) Debugf(format string, args ...interface{}) {
	l.logger.Logf(logrus.DebugLevel, format, args...)
}

func (l *LogrusLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Logf(logrus.FatalLevel, format, args...)
}
