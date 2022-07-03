package jsEngine

import (
	"github.com/dop251/goja"
	"github.com/mcfly722/goPackages/context"
)

// Console ...
type Console struct {
	context context.Context
}

// APIConsole ...
func APIConsole(context context.Context, eventLoop EventLoop, runtime *goja.Runtime) {

	log := runtime.NewObject()
	log.Set("log", func(msg string) {
		context.Log(50, msg)
	})

	runtime.Set("console", log)
}