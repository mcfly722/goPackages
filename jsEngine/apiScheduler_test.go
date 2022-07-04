package jsEngine_test

import (
	"os"
	"os/signal"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/mcfly722/goPackages/context"
	"github.com/mcfly722/goPackages/jsEngine"
)

// enables js schuduler functionality
func ExampleScheduler() {
	eventLoop := jsEngine.NewEventLoop(goja.New(), []jsEngine.Script{})
	eventLoop.Import(jsEngine.Scheduler{})
}

func Test_SetInterval(t *testing.T) {

	script := jsEngine.NewScript("test", `
    timerId1 = setInterval(function(){
    	Console.Log("timer1")
    },1000)

    Console.Log("timer with id=" + timerId1 + " initialized")
	`)

	rootContext := context.NewRootContext(context.NewConsoleLogDebugger(100, true))

	eventLoop := jsEngine.NewEventLoop(goja.New(), []jsEngine.Script{script})

	eventLoop.Import(jsEngine.Console{})
	eventLoop.Import(jsEngine.Scheduler{})

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

	{ // wait 5 second till exit
		time.Sleep(5 * time.Second)
		rootContext.Log(2, "timeout")
		rootContext.Cancel()
	}

	rootContext.Wait()
}
