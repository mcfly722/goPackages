package jsEngine_test

import (
	"fmt"
	"testing"

	"github.com/dop251/goja"
	"github.com/mcfly722/goPackages/context"
)

func Test_CallHandler(t *testing.T) {

	rootContext := context.NewRootContext(context.NewConsoleLogDebugger(100, true))

	eventLoop := NewEventLoop(goja.New(), scripts)
	eventLoop.addAPI(apiConsole)

	script := `
	function handler(param1, param2) {
		 log('handler executed '+param1+ ','+param2)

		 someUnknownFunctionCall()

		 return (param1 / param2)
	}

	handle(handler,2,3)
	`

	handle := func(handler *goja.Callable, args ...goja.Value) {

		value, err := (*handler)(nil, args...)
		if err != nil {
			fmt.Println(fmt.Sprintf("handler error:%v", err))
		} else {
			fmt.Println(fmt.Sprintf("handler returned value:%v", value))
		}

	}

	log := func(msg string) { fmt.Println(msg) }

	runtime.Set("handle", handle)
	runtime.Set("log", log)

}
