package utils

import (
	"io"
	"os"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

type WriterHook struct {
	Writer    io.Writer
	LogLevels []logrus.Level
}

func (hook *WriterHook) Fire(entry *logrus.Entry) error {
	for _, level := range hook.LogLevels {
		if entry.Level == level {
			line, err := entry.String()
			if err != nil {
				return err
			}
			_, err = hook.Writer.Write([]byte(line))
			return err
		}
	}
	return nil // do nothing if level not in LogLevels
}

func (hook *WriterHook) Levels() []logrus.Level {
	return hook.LogLevels
}

func InitLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint:     true,
		TimestampFormat: time.RFC3339,
	})

	logger.SetOutput(io.Discard) // prevent duplicate log

	appModeEnv := os.Getenv("APP_MODE")
	isDev := appModeEnv == "" || strings.ToLower(appModeEnv) == "dev"

	infoWriter, err := rotatelogs.New(
		"./logs/app-info.%Y-%m-%d.log",
		rotatelogs.WithLinkName("./logs/app-info.log"),
		rotatelogs.WithRotationTime(24*time.Hour),
		rotatelogs.WithMaxAge(14*24*time.Hour),
	)
	if err != nil {
		panic("failed to create info log rotator: " + err.Error())
	}

	errorWriter, err := rotatelogs.New(
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

	logger.AddHook(&WriterHook{
		Writer: infoOutput,
		LogLevels: []logrus.Level{
			logrus.InfoLevel,
			logrus.WarnLevel,
			logrus.DebugLevel,
		},
	})

	logger.AddHook(&WriterHook{
		Writer: errorOutput,
		LogLevels: []logrus.Level{
			logrus.ErrorLevel,
			logrus.FatalLevel,
			logrus.PanicLevel,
		},
	})

	logger.SetLevel(logrus.DebugLevel)
	return logger
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
