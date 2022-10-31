package logzap

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// Logger ...
type Logger struct {
	*zap.SugaredLogger //*zap.Logger
}

//type ctxLogger struct{}

// ContextWithLogger adds logger to context
func ContextWithLogger(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, Logger{}, l)
}

// LoggerFromContext returns logger from context
func LoggerFromContext(ctx context.Context) *Logger {
	if l, ok := ctx.Value(Logger{}).(*Logger); ok {
		return l
	}
	return &Logger{zap.L().Sugar()}
}

// New ...
func New(logLevel string) *Logger {

	loggerConfig := zap.NewProductionEncoderConfig()
	//loggerConfig.EncodeTime = syslogTimeEncoder //zapcore.ISO8601TimeEncoder
	loggerConfig.TimeKey = "timestamp"
	loggerConfig.EncodeTime = zapcore.TimeEncoderOfLayout("02.01.2006 03:04:05 PM")

	fileEncoder := zapcore.NewJSONEncoder(loggerConfig) // файл будет в JSON формате
	consoleEncoder := zapcore.NewConsoleEncoder(loggerConfig)

	// задаём файл для логов
	ex, _ := os.Executable()
	fileLog := fileNameWithoutExtension(filepath.Base(ex)) + ".log"
	writer := zapcore.AddSync(getLogWriter(fileLog))
	//logFile, _ := os.OpenFile(fileLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	//writer := zapcore.AddSync(logFile)

	// определяем уровень логирования
	defaultLogLevel, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		defaultLogLevel = zapcore.DebugLevel
	}

	// будем писать и в файл и в консоль
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writer, defaultLogLevel),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel),
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{logger.Sugar()}
}

func fileNameWithoutExtension(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

func getLogWriter(fileLog string) zapcore.WriteSyncer {
	fileWriter := &lumberjack.Logger{
		Filename:   fileLog, // Имя файла
		MaxSize:    10,      // Размер в МБ до ротации файла
		MaxBackups: 5,       // Максимальное количество файлов, сохраненных до перезаписи
		MaxAge:     30,      // Максимальное количество дней для хранения файлов
		Compress:   true,    // Следует ли сжимать файлы логов с помощью gzip
	}
	return zapcore.AddSync(fileWriter)
}

// func (logger *Logger) Error(args ...interface{}) {
// 	logger.SugaredLogger.Error(args...)
// }

// func (logger *Logger) Errorf(format string, args ...interface{}) {
// 	logger.SugaredLogger.Errorf(format, args...)
// }

// func (logger *Logger) Fatal(args ...interface{}) {
// 	logger.SugaredLogger.Fatal(args...)
// }
// func (logger *Logger) Fatalf(format string, args ...interface{}) {
// 	logger.SugaredLogger.Fatalf(format, args...)
// }

// func (logger *Logger) Info(args ...interface{}) {
// 	logger.SugaredLogger.Info(args...)
// }
// func (logger *Logger) Infof(format string, args ...interface{}) {
// 	logger.SugaredLogger.Infof(format, args)
// }

// func (logger *Logger) Warn(args ...interface{}) {
// 	logger.SugaredLogger.Warn(args...)
// }
// func (logger *Logger) Warnf(format string, args ...interface{}) {
// 	logger.SugaredLogger.Warnf(format, args...)
// }

// func (logger *Logger) Debug(args ...interface{}) {
// 	logger.SugaredLogger.Debug(args...)
// }
// func (logger *Logger) Debugf(format string, args ...interface{}) {
// 	logger.SugaredLogger.Debugf(format, args...)
// }
