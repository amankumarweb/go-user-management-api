package logger

import "go.uber.org/zap"

// Log is the package-level logger used throughout the application.
var Log *zap.Logger

// Init initializes the production Zap logger.
// Call this once at application startup.
func Init() {
	var err error
	Log, err = zap.NewProduction()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
}

// Sync flushes any buffered log entries. Call this before application exit.
func Sync() {
	_ = Log.Sync()
}
