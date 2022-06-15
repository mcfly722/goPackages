package plugins_test

import (
	"os"
	"os/signal"
	"testing"

	"github.com/mcfly722/goPackages/context"
	"github.com/mcfly722/goPackages/plugins"
)

type plugin struct {
	path      string
	terminate chan bool
}

func newPlugin(path string) plugins.Plugin {
	return &plugin{
		path:      path,
		terminate: make(chan bool),
	}
}

func (plugin *plugin) Go(current context.Context) {
loop:
	for {
		select {
		case <-plugin.terminate:
			current.Log(102, "terminate")
			break loop
		case <-current.OnDone():
			break loop
		}

	}
}

func (plugin *plugin) Dispose(current context.Context) {}

func (plugin *plugin) Terminate() {
	plugin.terminate <- true
}

func Test_AsServer(t *testing.T) {
	pluginsPath := ""

	rootCtx := context.NewRootContext(context.NewConsoleLogDebugger())

	pluginsManager := plugins.NewPluginsManager(pluginsPath, "*.go", 1, newPlugin)

	rootCtx.NewContextFor(pluginsManager, "pluginsManager", "pluginsManager")

	{ // handle ctrl+c for gracefully shutdown using context
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			<-c
			rootCtx.Log(1, "CTRL+C signal")
			rootCtx.Terminate()
		}()
	}

	rootCtx.Wait()

	rootCtx.Log(0, "done")
}
