package jsEngine_test

import (
	"github.com/dop251/goja"
	"github.com/mcfly722/goPackages/jsEngine"
)

// enables js console functionality
func ExampleConsole() {
	eventLoop := jsEngine.NewEventLoop(goja.New(), []jsEngine.Script{})
	eventLoop.Import(jsEngine.Console{})
}
