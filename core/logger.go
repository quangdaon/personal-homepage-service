package core

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"personal-homepage-service/config"
	"time"
)

func NewLogger(cfg config.Config) (*zap.Logger, error) {
	// Get the current UTC date to create a new file per run
	runTimestamp := time.Now().UTC().Format("2006-01-02T15-04-05")
	logFile := fmt.Sprintf("%v/personal-homepage-service-%s.log", cfg.LogsDirectory, runTimestamp)

	// Set up lumberjack for daily rotation
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFile, // Unique file for each run
		MaxSize:    100,     // MB before it rolls
		MaxBackups: 7,       // Keep last 7 logs
		MaxAge:     30,      // Days
		Compress:   true,    // Compress rotated logs
	}

	// Zap core setup
	writeSyncer := zapcore.AddSync(lumberjackLogger)
	encoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:      "ts",
		LevelKey:     "level",
		MessageKey:   "msg",
		CallerKey:    "caller",
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeLevel:  zapcore.CapitalLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	})

	core := zapcore.NewCore(encoder, writeSyncer, zap.InfoLevel)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger, nil
}
