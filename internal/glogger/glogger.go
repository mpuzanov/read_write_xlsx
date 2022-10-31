package glogger

import (
	"read_write_xlsx/pkg/logging"
	"read_write_xlsx/pkg/logzap"
	"strings"
)

// Logger represent common interface for logging function
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})

	Info(args ...interface{})
	Infof(format string, args ...interface{})

	Warn(args ...interface{})
	Warnf(format string, args ...interface{})

	Error(args ...interface{})
	Errorf(format string, args ...interface{})

	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

// BuildLogger создаем выбранный логгер
func BuildLogger(logType, Level string) Logger {
	var logger Logger
	switch strings.ToUpper(logType) {
	case "LOGRUS":
		logger = logging.New(Level, false) // logrus
	// case "STD":
	// 	logger = logstd.New(Level) // log
	default:
		logType = "ZAP"
		logger = logzap.New(Level) // zap
	}
	logger.Debugf("Используем logger: %s", logType)
	return logger
}
