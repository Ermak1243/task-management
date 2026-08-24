// Package logger предоставляет обёртку над zap-логгером с настраиваемым
// уровнем, форматом (console/json) и поддержкой режима разработки.
package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config хранит параметры логгера: уровень, формат и флаг dev-режима.
type Config struct {
	Level    string
	Encoding string
	IsDev    bool
}

// Logger — обёртка над *zap.SugaredLogger с методами разных уровней.
type Logger struct {
	sugar *zap.SugaredLogger
}

// New создаёт логгер с указанными параметрами. По умолчанию:
// level = "debug", encoding = "console". В dev-режиме используется
// DevelopmentEncoderConfig с подсветкой уровней.
func New(cfg Config) *Logger {
	if cfg.Level == "" {
		cfg.Level = "debug"
	}
	if cfg.Encoding == "" {
		cfg.Encoding = "console"
	}

	var logLevel zapcore.Level
	switch cfg.Level {
	case "debug":
		logLevel = zapcore.DebugLevel
	case "info":
		logLevel = zapcore.InfoLevel
	case "warn":
		logLevel = zapcore.WarnLevel
	case "error":
		logLevel = zapcore.ErrorLevel
	default:
		logLevel = zapcore.InfoLevel
	}

	writer := zapcore.AddSync(os.Stderr)
	encoderCfg := zap.NewProductionEncoderConfig()
	if cfg.IsDev {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
	}
	encoderCfg.TimeKey = "time"
	encoderCfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}
	encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	var encoder zapcore.Encoder
	if cfg.Encoding == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	core := zapcore.NewCore(encoder, writer, zap.NewAtomicLevelAt(logLevel))
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return &Logger{sugar: zapLogger.Sugar()}
}

func (l *Logger) Debug(args ...interface{})               { l.sugar.Debug(args...) }
func (l *Logger) Debugf(tmpl string, args ...interface{}) { l.sugar.Debugf(tmpl, args...) }
func (l *Logger) Info(args ...interface{})                { l.sugar.Info(args...) }
func (l *Logger) Infof(tmpl string, args ...interface{})  { l.sugar.Infof(tmpl, args...) }
func (l *Logger) Warn(args ...interface{})                { l.sugar.Warn(args...) }
func (l *Logger) Warnf(tmpl string, args ...interface{})  { l.sugar.Warnf(tmpl, args...) }
func (l *Logger) Error(args ...interface{})               { l.sugar.Error(args...) }
func (l *Logger) Errorf(tmpl string, args ...interface{}) { l.sugar.Errorf(tmpl, args...) }
func (l *Logger) Fatal(args ...interface{})               { l.sugar.Fatal(args...) }
func (l *Logger) Fatalf(tmpl string, args ...interface{}) { l.sugar.Fatalf(tmpl, args...) }
