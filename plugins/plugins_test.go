package plugins_test

import (
	"os"
	"os/signal"
	"testing"
	"time"

	"github.com/mcfly722/goPackages/context"
	"github.com/mcfly722/goPackages/plugins"
)

type plugin struct {
	definition plugins.PluginDefinition
	counter    int
}

func newPlugin(definition plugins.PluginDefinition) context.ContextedInstance {
	return &plugin{
		definition: definition,
		counter:    0,
	}
}

func (plugin *plugin) Go(current context.Context) {
loop:
	for {
		select {
		case <-time.After(1 * time.Second):
			plugin.counter++

			if plugin.definition.Outdated() {
				current.Log(102, "terminate")
				current.Cancel()
			}

			if plugin.counter > 5 {
				current.Log(102, "terminate by counter1")
				current.Cancel()
			}

			break
		case _, opened := <-current.Opened():
			if !opened {
				break loop
			}
		}

	}
}

func Test_AsServer(t *testing.T) {
	pluginsPath := ""

	rootCtx := context.NewRootContext(context.NewConsoleLogDebugger(100, true))

	pluginsProvider := plugins.NewPluginsFromFilesProvider(pluginsPath, "*.go")

	pluginsManager := plugins.NewPluginsManager(pluginsProvider, 1, newPlugin)

	rootCtx.NewContextFor(pluginsManager, "pluginsManager", "pluginsManager")

	{ // handle ctrl+c for gracefully shutdown using context
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			<-c
			rootCtx.Log(1, "CTRL+C signal")
			rootCtx.Cancel()
		}()
	}

	rootCtx.Wait()

	rootCtx.Log(0, "done")
}
