package utils

import (
	"io"
	"os"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

// WriterHook is a logrus hook that writes logs of specified levels to a given io.Writer.
type WriterHook struct {
	Writer    io.Writer
	LogLevels map[logrus.Level]struct{}
}

// NewWriterHook creates a new WriterHook for the specified writer and log levels.
func NewWriterHook(writer io.Writer, levels []logrus.Level) *WriterHook {
	levelMap := make(map[logrus.Level]struct{}, len(levels))
	for _, lvl := range levels {
		levelMap[lvl] = struct{}{}
	}
	return &WriterHook{Writer: writer, LogLevels: levelMap}
}

// Fire writes the log entry to the Writer if its level is enabled in the hook.
func (hook *WriterHook) Fire(entry *logrus.Entry) error {
	if _, ok := hook.LogLevels[entry.Level]; ok {
		line, err := entry.String()
		if err != nil {
			return err
		}
		_, err = hook.Writer.Write([]byte(line))
		return err
	}
	return nil // do nothing if level not in LogLevels
}

// Levels returns the log levels enabled for this WriterHook.
func (hook *WriterHook) Levels() []logrus.Level {
	levels := make([]logrus.Level, 0, len(hook.LogLevels))
	for lvl := range hook.LogLevels {
		levels = append(levels, lvl)
	}
	return levels
}

// RotatelogsNewFunc defines the function signature for rotatelogs.New, allowing injection for testing.
type RotatelogsNewFunc func(string, ...rotatelogs.Option) (*rotatelogs.RotateLogs, error)

// InitLoggerWithCreators creates a logrus.Logger with hooks for info and error logs, allowing injection of log writer creators for testing.
func InitLoggerWithCreators(
	infoLogCreator RotatelogsNewFunc,
	errorLogCreator RotatelogsNewFunc,
) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint:     true,
		TimestampFormat: time.RFC3339,
	})

	logger.SetOutput(io.Discard) // prevent duplicate log

	appModeEnv := os.Getenv("APP_MODE")
	isDev := appModeEnv == "" || strings.ToLower(appModeEnv) == "dev"

	infoWriter, err := infoLogCreator(
		"./logs/app-info.%Y-%m-%d.log",
		rotatelogs.WithLinkName("./logs/app-info.log"),
		rotatelogs.WithRotationTime(24*time.Hour),
		rotatelogs.WithMaxAge(14*24*time.Hour),
	)
	if err != nil {
		panic("failed to create info log rotator: " + err.Error())
	}

	errorWriter, err := errorLogCreator(
		"./logs/app-error.%Y-%m-%d.log",
		rotatelogs.WithLinkName("./logs/app-error.log"),
		rotatelogs.WithRotationTime(24*time.Hour),
		rotatelogs.WithMaxAge(14*24*time.Hour),
	)
	if err != nil {
		panic("failed to create error log rotator: " + err.Error())
	}

	var infoOutput io.Writer = infoWriter
	var errorOutput io.Writer = errorWriter

	if isDev {
		infoOutput = io.MultiWriter(infoWriter, os.Stdout)
		errorOutput = io.MultiWriter(errorWriter, os.Stdout)
	}

	logger.AddHook(NewWriterHook(infoOutput, []logrus.Level{
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.DebugLevel,
	}))

	logger.AddHook(NewWriterHook(errorOutput, []logrus.Level{
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}))

	logger.SetLevel(logrus.DebugLevel)
	return logger
}

// InitLogger creates a logrus.Logger for production use, writing to rotating log files.
func InitLogger() *logrus.Logger {
	return InitLoggerWithCreators(rotatelogs.New, rotatelogs.New)
}

// Lumberjack
// func InitLogger() *logrus.Logger {
// 	logger := logrus.New()

// 	logger.SetFormatter(&logrus.JSONFormatter{
// 		TimestampFormat: time.RFC3339,
// 	})

// 	logger.SetOutput(io.Discard) // prevent duplicate log

// 	infoFile := &lumberjack.Logger{
// 		Filename:   "./logs/app-info.log",
// 		MaxSize:    10,
// 		MaxBackups: 7,
// 		MaxAge:     14,
// 		Compress:   true,
// 	}

// 	errorFile := &lumberjack.Logger{
// 		Filename:   "./logs/app-error.log",
// 		MaxSize:    10,
// 		MaxBackups: 7,
// 		MaxAge:     14,
// 		Compress:   true,
// 	}

// 	stdout := os.Stdout

// 	logger.AddHook(&WriterHook{
// 		Writer: io.MultiWriter(infoFile, stdout),
// 		LogLevels: []logrus.Level{
// 			logrus.InfoLevel,
// 			logrus.WarnLevel,
// 			logrus.DebugLevel,
// 		},
// 	})

// 	logger.AddHook(&WriterHook{
// 		Writer: io.MultiWriter(errorFile, stdout),
// 		LogLevels: []logrus.Level{
// 			logrus.ErrorLevel,
// 			logrus.FatalLevel,
// 			logrus.PanicLevel,
// 		},
// 	})

// 	logger.SetLevel(logrus.DebugLevel)
// 	return logger
// }
