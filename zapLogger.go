package main

import (
	"context"
	"ePrometna_Server/config"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

/*
	type Interface interface {
	  LogMode(LogLevel) Interface
	  Info(context.Context, string, ...interface{})
	  Warn(context.Context, string, ...interface{})
	  Error(context.Context, string, ...interface{})
	  Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error)
	}
*/
type gormZapLogger struct {
	logger.Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

func NewGormZapLogger() logger.Interface {
	var (
		infoStr      = "%s\t%s"
		warnStr      = "%s\t%s"
		errStr       = "%s\t%s"
		traceStr     = "%s\t[%.3fms] [rows:%v] %s"
		traceWarnStr = "%s %s\t[%.3fms] [rows:%v] %s"
		traceErrStr  = "%s %s\t[%.3fms] [rows:%v] %s"
	)

	return &gormZapLogger{
		Config: logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: false,
			Colorful:                  true,
		},
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
}

func (l *gormZapLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

func (l *gormZapLogger) Info(c context.Context, msg string, args ...any) {
	if l.LogLevel >= logger.Info {
		zap.S().Infof(l.infoStr+msg, args...)
	}
}

func (l *gormZapLogger) Warn(c context.Context, msg string, args ...any) {
	if l.LogLevel >= logger.Warn {
		zap.S().Warnf(l.warnStr+msg, args...)
	}
}

func (l *gormZapLogger) Error(c context.Context, msg string, args ...any) {
	if l.LogLevel >= logger.Error {
		zap.S().Errorf(l.errStr+msg, args...)
	}
}

func (l *gormZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}
	elapsed := time.Since(begin)

	switch {
	case err != nil && l.LogLevel >= logger.Error && (!errors.Is(err, logger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			zap.S().Errorf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			zap.S().Errorf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			zap.S().Warnf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			zap.S().Warnf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.LogLevel == logger.Info:
		sql, rows := fc()
		if rows == -1 {
			zap.S().Infof(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			zap.S().Infof(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}

func devLoggerSetup() error {
	logger, err := zap.NewDevelopment(zap.AddStacktrace(zap.ErrorLevel))
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
