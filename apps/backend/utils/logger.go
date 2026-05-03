package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a structured logger using uber-go/zap.
// In development, we use a human-readable format.
// In production, we use JSON format for log aggregation tools (Datadog, CloudWatch, etc.)
func NewLogger(isDevelopment bool) (*zap.Logger, error) {
	var config zap.Config

	if isDevelopment {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
	}

	return config.Build()
}
