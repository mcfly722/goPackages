package jsEngine_test

import (
	"fmt"
	"os"
	"os/signal"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/mcfly722/goPackages/context"
	"github.com/mcfly722/goPackages/jsEngine"
)

type testModule struct{}

func (module testModule) Constructor(context context.Context, eventLoop jsEngine.EventLoop, runtime *goja.Runtime) {

	handle := func(handler *goja.Callable, args ...goja.Value) {

		value, err := (*handler)(nil, args...)
		if err != nil {
			fmt.Println(fmt.Sprintf("handler error:%v", err))
		} else {
			fmt.Println(fmt.Sprintf("handler returned value:%v", value))
		}

	}
	runtime.Set("handle", handle)
}

func testScript(scriptBody string, timeoutSec int) {
	script := jsEngine.NewScript("test", scriptBody)

	rootContext := context.NewRootContext(context.NewConsoleLogDebugger(100, true))

	eventLoop := jsEngine.NewEventLoop(goja.New(), []jsEngine.Script{script})

	eventLoop.Import(jsEngine.Console{})
	eventLoop.Import(testModule{})

	rootContext.NewContextFor(eventLoop, "jsEngine", "eventLoop")

	{ // handle ctrl+c for gracefully shutdown using context
		ctrlC := make(chan os.Signal, 1)
		signal.Notify(ctrlC, os.Interrupt)
		go func() {
			<-ctrlC
			rootContext.Log(2, "CTRL+C signal")
			rootContext.Cancel()
		}()
	}

	{ // wait 1 second till exit
		time.Sleep(time.Duration(timeoutSec) * time.Second)
		rootContext.Log(2, "timeout")
		rootContext.Cancel()
	}

	rootContext.Wait()
}

func Test_CallHandler(t *testing.T) {
	testScript(`
	function handler(param1, param2) {
		 Console.Log('handler executed '+param1+ ','+param2)

		 //someUnknownFunctionCall()

		 return (param1 / param2)
	}

	handle(handler,2,3)
	`, 1)
}

func Test_ScriptException(t *testing.T) {
	testScript("console.log(123)", 1)
}
