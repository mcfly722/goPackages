package jsEngine

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/mcfly722/goPackages/logger"
)

// JSRun ...
type JSRun struct{}

// Initialize ...
func (jsRun *JSRun) Initialize(runtime *JSRuntime) error {

	run := func(cmd string, args []string, timeoutMilliseconds int64, onSuccess goja.Callable) {

		output := fmt.Sprintf("command:%v\nargs:%v\ntimeout:%v", cmd, args, timeoutMilliseconds)

		runtime.CallCallback(&onSuccess, runtime.VM.ToValue(output))

	}

	runtime.VM.Set("run", run)

	runtime.Logger.LogEvent(logger.EventTypeInfo, runtime.Name, "runner initialized")

	return nil
}

// Dispose ...
func (jsRun *JSRun) Dispose(runtime *JSRuntime) {
	runtime.Logger.LogEvent(logger.EventTypeInfo, runtime.Name, "runner disposed")
}
