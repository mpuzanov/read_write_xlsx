package logging

import (
	"context"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var e *logrus.Entry

// Logger ...
type Logger struct {
	*logrus.Entry
}

// GetLogger ...
func GetLogger() *Logger {
	return &Logger{e}
}

// ContextWithLogger adds logger to context
func ContextWithLogger(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, Logger{}, l)
}

// LoggerFromContext returns logger from context
func LoggerFromContext(ctx context.Context) *Logger {
	if l, ok := ctx.Value(Logger{}).(*Logger); ok {
		return l
	}
	return &Logger{logrus.NewEntry(logrus.New())}
}

// New ...
func New(Level string, IsJSON bool) *Logger {
	l := logrus.New()

	formatter := &logrus.TextFormatter{
		TimestampFormat:        "2006-01-02 15:04:05",
		FullTimestamp:          true,
		DisableLevelTruncation: true, // log level field configuration
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			fileName := path.Base(frame.File) + ":" + strconv.Itoa(frame.Line)
			//return frame.Function, fileName
			return "", fileName
		},
	}
	formatterJSON := &logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		PrettyPrint:     false,
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			fileName := path.Base(frame.File) + ":" + strconv.Itoa(frame.Line)
			//return frame.Function, fileName
			return "", fileName
		},
	}
	if IsJSON {
		l.SetFormatter(formatterJSON)
	} else {
		l.SetFormatter(formatter)
	}

	var fileLog string
	ex, _ := os.Executable()
	workDir := filepath.Dir(ex) // путь к программе
	fileLog = fileNameWithoutExtension(filepath.Base(ex)) + ".log"
	fileLog = filepath.Join(workDir, fileLog)
	// file, err := os.OpenFile(fileLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// if err != nil {
	// 	return err
	// }

	file := &lumberjack.Logger{
		Filename:   fileLog, // Имя файла
		MaxSize:    10,      // Размер в МБ до ротации файла
		MaxBackups: 5,       // Максимальное количество файлов, сохраненных до перезаписи
		MaxAge:     30,      // Максимальное количество дней для хранения файлов
		Compress:   true,    // Следует ли сжимать файлы логов с помощью gzip
	}

	multi := io.MultiWriter(file, os.Stdout) //, os.Stderr
	//l.SetOutput(os.Stdout)
	l.SetOutput(multi)

	l.SetReportCaller(true) // logrus show line number

	// установка уровня логирования
	loglevel, err := logrus.ParseLevel(Level)
	if err != nil {
		loglevel = logrus.DebugLevel
	}
	l.SetLevel(loglevel)

	e = logrus.NewEntry(l)
	return &Logger{e}
}

func fileNameWithoutExtension(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

// func (logger *Logger) Error(args ...interface{}) {
// 	logger.Entry.Error(args...)
// }
// func (logger *Logger) Errorf(format string, args ...interface{}) {
// 	logger.Entry.Errorf(format, args...)
// }

// func (logger *Logger) Fatal(args ...interface{}) {
// 	logger.Entry.Fatal(args...)
// }
// func (logger *Logger) Fatalf(format string, args ...interface{}) {
// 	logger.Entry.Fatalf(format, args...)
// }

// func (logger *Logger) Info(args ...interface{}) {
// 	logger.Entry.Info(args...)
// }
// func (logger *Logger) Infof(format string, args ...interface{}) {
// 	logger.Entry.Infof(format, args)
// }

// func (logger *Logger) Warn(args ...interface{}) {
// 	logger.Entry.Warn(args...)
// }
// func (logger *Logger) Warnf(format string, args ...interface{}) {
// 	logger.Entry.Warnf(format, args...)
// }

// func (logger *Logger) Debug(args ...interface{}) {
// 	logger.Entry.Debug(args...)
// }
// func (logger *Logger) Debugf(format string, args ...interface{}) {
// 	logger.Entry.Debugf(format, args...)
// }
