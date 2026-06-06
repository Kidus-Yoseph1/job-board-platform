package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a custom wrapper around zap.Logger.
type Logger struct {
	*zap.SugaredLogger
}

var log *Logger

// Init initializes the custom zap logger based on the environment.
func Init(env string) {
	err := os.MkdirAll("logs", 0755)
	if err != nil {
		panic("failed to create logs directory: " + err.Error())
	}

	var fileEncoder, consoleEncoder zapcore.Encoder
	var level zapcore.Level

	if env == "production" {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		fileEncoder = zapcore.NewJSONEncoder(encoderConfig)
		consoleEncoder = zapcore.NewJSONEncoder(encoderConfig)
		level = zap.InfoLevel
	} else {
		fileEncConfig := zap.NewDevelopmentEncoderConfig()
		fileEncConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		fileEncConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		fileEncoder = zapcore.NewConsoleEncoder(fileEncConfig)

		consoleEncConfig := zap.NewDevelopmentEncoderConfig()
		consoleEncConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		consoleEncConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		consoleEncoder = zapcore.NewConsoleEncoder(consoleEncConfig)

		level = zap.DebugLevel
	}

	appFile, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic("failed to open logs/app.log: " + err.Error())
	}
	errorFile, err := os.OpenFile("logs/error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic("failed to open logs/error.log: " + err.Error())
	}

	appConsoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level)
	appFileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(appFile), level)

	errorConsoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stderr), zap.ErrorLevel)
	errorFileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(errorFile), zap.ErrorLevel)

	core := zapcore.NewTee(
		appConsoleCore,
		appFileCore,
		errorConsoleCore,
		errorFileCore,
	)

	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))

	log = &Logger{
		SugaredLogger: zapLogger.Sugar(),
	}
}

// Get returns the global custom logger instance.
func Get() *Logger {
	if log == nil {
		fallback, _ := zap.NewDevelopment()
		log = &Logger{SugaredLogger: fallback.Sugar()}
	}
	return log
}
