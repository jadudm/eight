package env

import (
	"go.uber.org/zap"
)

func SetupLogging() {
	zap.ReplaceGlobals(zap.Must(zap.NewProduction()))
}
