package jsEngine

import (
	"github.com/mcfly722/goPackages/logger"
)

// JSLog ...
type JSLog struct{}

// Initialize ...
func (jsLog *JSLog) Initialize(runtime *JSRuntime) error {
	log := func(msg string) {
		runtime.Logger.LogEvent(logger.EventTypeInfo, runtime.Name, msg)
	}

	runtime.VM.Set("log", log)

	runtime.Logger.LogEvent(logger.EventTypeInfo, runtime.Name, "log initialized")

	return nil
}

// Dispose ...
func (jsLog *JSLog) Dispose(runtime *JSRuntime) {
	runtime.Logger.LogEvent(logger.EventTypeInfo, runtime.Name, "log disposed")
}
