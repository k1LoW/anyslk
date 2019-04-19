package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const rotationCount = 60
const rotationTime = 24 * time.Hour

// NewLogger ...
func NewLogger(dir string) *zap.Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	cores := []zapcore.Core{}
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	w := newLogWriter(dir)
	stdoutCore := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		zapcore.InfoLevel,
	)
	cores = append(cores, stdoutCore)
	logCore := zapcore.NewCore(
		encoder,
		zapcore.AddSync(w),
		zapcore.InfoLevel,
	)
	cores = append(cores, logCore)
	logger := zap.New(zapcore.NewTee(cores...))

	return logger
}

func newLogWriter(dir string) io.Writer {
	fileName := "anyslk.log"
	path, err := filepath.Abs(fmt.Sprintf("%s/%s", dir, fileName))
	if err != nil {
		log.Fatalf("Log setting error %v", err)
	}
	options := []rotatelogs.Option{
		rotatelogs.WithClock(rotatelogs.Local),
		rotatelogs.WithMaxAge(-1),
	}
	options = append(options, rotatelogs.WithRotationCount(rotationCount))
	options = append(options, rotatelogs.WithLinkName(path))
	options = append(options, rotatelogs.WithRotationTime(rotationTime))
	logSuffix := ".%Y%m%d"
	w, err := rotatelogs.New(
		path+logSuffix,
		options...,
	)
	if err != nil {
		log.Fatalf("Log setting error %v", err)
	}
	return w
}
