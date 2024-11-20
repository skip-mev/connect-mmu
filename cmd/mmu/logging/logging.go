package logging

import (
	"context"

	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/lib/log"
)

type loggerKey struct{}

func ConfigureLogger(level string) {
	config := log.NewDefaultZapConfig()
	config.StdOutLogLevel = level
	logger := log.NewZapLogger(config)
	zap.ReplaceGlobals(logger)
}

func Logger(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(loggerKey{}).(*zap.Logger)
	if !ok {
		return zap.L()
	}
	return logger
}

func LoggerContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerKey{}, zap.L())
}
