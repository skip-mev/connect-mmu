package log

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapConfig is the configuration for a zap logger.
type ZapConfig struct {
	// StdOutLogLevel is the log level for the standard out logger.
	StdOutLogLevel string
}

// NewDefaultZapConfig creates a default configuration for a zap logger.
func NewDefaultZapConfig() ZapConfig {
	return ZapConfig{
		StdOutLogLevel: "info",
	}
}

func NewZapLogger(config ZapConfig) *zap.Logger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	// set up the primary output to always include os.Stderr.
	logLevel := zapcore.InfoLevel
	if err := logLevel.Set(config.StdOutLogLevel); err != nil {
		fmt.Fprintf(os.Stderr, "failed to set log level on std out: %v\nfalling back to info", err)
		logLevel = zapcore.InfoLevel // Fallback to info if setting fails
	}

	// set up the primary output to always include os.Stderr
	stdCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stderr),
		logLevel,
	)

	return zap.New(
		stdCore,
		zap.AddCaller(),
		zap.Fields(zapcore.Field{Key: "pid", Type: zapcore.Int64Type, Integer: int64(os.Getpid())}),
	)
}
