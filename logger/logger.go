package logger

import (
	"go.uber.org/zap"
)

// Instance of logger
var Instance *zap.Logger

func init() {
	Instance, _ = zap.NewDevelopment()
}
