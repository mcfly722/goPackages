package plugins_test

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"testing"

	"github.com/mcfly722/goPackages/context"
	"github.com/mcfly722/goPackages/logger"
	"github.com/mcfly722/goPackages/plugins"
)

type Plugin struct{}

// OnLoad ...
func (plugin *Plugin) OnLoad(pluginsFullPath string, relativeName string, body string) {
	log.Println(fmt.Sprintf("%v loaded", relativeName))
}

// OnUpdate ...
func (plugin *Plugin) OnUpdate(pluginsFullPath string, relativeName string, body string) {
	log.Println(fmt.Sprintf("%v updated", relativeName))
}

// OnUnload ...
func (plugin *Plugin) OnDispose(pluginsFullPath string, relativeName string) {
	log.Println(fmt.Sprintf("%v uloaded", relativeName))
}

// UpdateRequired ...
func (plugin *Plugin) UpdateRequired(pluginsFullPath string, relativeName string) bool {
	return false
}

type root struct{}

func (root *root) Go(current context.Context) {
loop:
	for {
		select {
		case <-current.OnDone():
			break loop
		}
	}
}

func (root *root) Dispose() {}

func Test_AsServer(t *testing.T) {

	rootCtx := context.NewContextFor(&root{})

	pluginsConstructor := func() plugins.IPlugin {
		return &Plugin{}
	}

	pluginsManager, err := plugins.NewPluginsManager("", "*.go", 3, pluginsConstructor)
	if err != nil {
		t.Fatal(err)
	}

	log := logger.NewLogger(5)
	log.SetOutputToConsole(true)
	pluginsManager.SetLogger(log)

	rootCtx.NewContextFor(pluginsManager)

	{ // handle ctrl+c for gracefully shutdown using context
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			<-c
			log.LogEvent(logger.EventTypeInfo, "Test_AsServer", "CTRL+C signal")
			rootCtx.OnDone() <- true
		}()
	}

	rootCtx.Wait()

	log.LogEvent(logger.EventTypeInfo, "Test_AsServer", "done")
}
