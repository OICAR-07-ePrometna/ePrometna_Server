package main

import (
	"context"
	"ePrometna_Server/config"
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
	ll                                  logger.LogLevel
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

func NewGormZapLogger() logger.Interface {
	var (
		infoStr      = "%s\n[info] "
		warnStr      = "%s\n[warn] "
		errStr       = "%s\n[error] "
		traceStr     = "%s\n[%.3fms] [rows:%v] %s"
		traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
		traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
	)

	return &gormZapLogger{
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
	newLogger.ll = level
	return &newLogger
}

func (l *gormZapLogger) Info(c context.Context, msg string, args ...any) {
	if l.ll >= logger.Info {
		zap.S().Infof(l.infoStr+msg, args...)
	}
}

func (l *gormZapLogger) Warn(c context.Context, msg string, args ...any) {
	if l.ll >= logger.Warn {
		zap.S().Warnf(l.warnStr+msg, args...)
	}
}

func (l *gormZapLogger) Error(c context.Context, msg string, args ...any) {
	if l.ll >= logger.Error {
		zap.S().Errorf(l.errStr+msg, args...)
	}
}

func (l *gormZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.ll >= logger.Silent {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()
	zap.S().Infof(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
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
