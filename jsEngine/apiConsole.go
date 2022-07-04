package jsEngine

import (
	"github.com/dop251/goja"
	"github.com/mcfly722/goPackages/context"
)

// Console module
type Console struct {
	context   context.Context
	eventLoop EventLoop
	runtime   *goja.Runtime
}

// Log ...
func (console *Console) Log(msg string) {
	console.context.Log(50, msg)
}

// Constructor ...
func (console Console) Constructor(context context.Context, eventLoop EventLoop, runtime *goja.Runtime) {
	runtime.Set("Console", &Console{
		context:   context,
		eventLoop: eventLoop,
		runtime:   runtime,
	})
}
