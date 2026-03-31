package logger

import "go.uber.org/zap"

func New() (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"stdout", "app.log"}
	cfg.ErrorOutputPaths = []string{"stderr", "app.log"}

	return cfg.Build()
}
