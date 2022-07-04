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

// enables js execution module functionality
func ExampleExec() {
	eventLoop := jsEngine.NewEventLoop(goja.New(), []jsEngine.Script{})
	eventLoop.Import(jsEngine.Exec{})
}

func Test_ExecProcess(t *testing.T) {

	script := jsEngine.NewScript("test", `
		function onDone(exitCode){
				console.log("exitCode="+exitCode)
		}

		function onStdout(content){
				console.log("content="+content)
		}

		p = Exec.NewCommand("ping.exe", ["-n","2", "0.0.0.0"]).SetTimeoutMs(30*1000).SetOnDone(onDone).SetOnStdoutString(onStdout).SetPath('C:/Users').StartNewProcess()

	`)

	rootContext := context.NewRootContext(context.NewConsoleLogDebugger(100, true))

	eventLoop := jsEngine.NewEventLoop(goja.New(), []jsEngine.Script{script})

	eventLoop.Import(jsEngine.Console{})
	eventLoop.Import(jsEngine.Scheduler{})
	eventLoop.Import(jsEngine.Exec{})

	rootContext.NewContextFor(eventLoop, "jsEngine", "eventLoop")

	{ // handle ctrl+c for gracefully shutdown using context
		ctrlC := make(chan os.Signal, 1)
		signal.Notify(ctrlC, os.Interrupt)
		go func() {
			<-ctrlC
			rootContext.Log(3, "CTRL+C signal")
			rootContext.Cancel()
		}()
	}

	{ // wait 5 second till exit
		go func() {
			time.Sleep(10 * time.Second)
			rootContext.Log(1, "timeout")
			rootContext.Cancel()
		}()
	}

	rootContext.Wait()
}
