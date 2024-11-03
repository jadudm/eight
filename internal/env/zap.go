package env

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func createLogger() *zap.Logger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	level := strings.ToLower(os.Getenv("DEBUG_LEVEL"))
	zap_level := zap.InfoLevel
	switch level {
	case "debug":
		zap_level = zap.DebugLevel
	case "info":
		zap_level = zap.InfoLevel
	case "warn":
		zap_level = zap.WarnLevel
	case "error":
		zap_level = zap.ErrorLevel
	default:
		zap_level = zap.InfoLevel
	}

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap_level),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:     encoderCfg,
		OutputPaths: []string{
			"stderr",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
		InitialFields: map[string]interface{}{
			"pid": os.Getpid(),
		},
	}

	return zap.Must(config.Build())
}

func SetupLogging() {
	zap.ReplaceGlobals(zap.Must(createLogger(), nil))
}
