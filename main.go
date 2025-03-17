package main

import (
	"ePrometna_Server/config"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	if config.AppConfig.IsDevelopment {
		devLoggerSetup()
	} else {
		prodLoggerSetup()
	}

	zap.S().Infoln("Bokic")
}

func devLoggerSetup() error {
	logger, err := zap.NewDevelopment(zap.AddStacktrace(zap.WarnLevel))
	if err != nil {
		return err
	}
	_ = zap.ReplaceGlobals(logger)

	return nil
}

func prodLoggerSetup() error {
	_ = os.Mkdir(config.LOG_FOLDER, 0755)

	// console log, text
	// log level
	consoleLogLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= zapcore.InfoLevel
	})
	// log output
	consoleLogFile := zapcore.Lock(os.Stdout)
	// log configuration
	consoleLogConfig := zap.NewProductionEncoderConfig()
	// configure keys
	// configure types
	consoleLogConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleLogConfig.EncodeLevel = nil
	consoleLogConfig.EncodeCaller = nil
	// create encoder
	consoleLogEncoder := zapcore.NewConsoleEncoder(consoleLogConfig)

	// file log, text
	// log level
	fileLogLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= zapcore.DebugLevel
	})
	// log output, with rotation
	logPath := filepath.Join(config.LOG_FOLDER, config.LOG_FILE)
	lumberjackLogger := lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    config.LOG_FILE_MAX_SIZE,    // size in MB
		MaxAge:     config.LOG_FILE_MAX_AGE,     // maximum number of days to retain old log files
		MaxBackups: config.LOG_FILE_MAX_BACKUPS, // maximum number of old log files to retain
		LocalTime:  true,                        // time used for formatting the timestamps
		Compress:   false,
	}
	fileLogFile := zapcore.Lock(zapcore.AddSync(&lumberjackLogger))
	// log configuration
	fileLogConfig := zap.NewProductionEncoderConfig()
	// configure keys
	fileLogConfig.TimeKey = "timestamp"
	fileLogConfig.MessageKey = "message"
	// configure types
	fileLogConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	fileLogConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	// create encoder
	fileLogEncoder := zapcore.NewConsoleEncoder(fileLogConfig)

	// setup zap
	// duplicate log entries into multiple cores
	core := zapcore.NewTee(
		zapcore.NewCore(consoleLogEncoder, consoleLogFile, consoleLogLevel),
		zapcore.NewCore(fileLogEncoder, fileLogFile, fileLogLevel),
	)

	// create logger from core
	// options = annotate message with the filename, line number, and function name
	logger := zap.New(core, zap.AddCaller())
	defer logger.Sync()

	// replace global logger
	_ = zap.ReplaceGlobals(logger)

	return nil
}
