package plugins

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/mcfly722/goPackages/context"
	"github.com/mcfly722/goPackages/logger"
)

// IPlugin ...
type IPlugin interface {
	OnLoad(relativeName string, body string)
	OnUpdate(relativeName string, body string)
	OnDispose(relativeName string)
}

type plugin struct {
	self         IPlugin
	relativeName string
	modification time.Time
	body         string
	parentLogger *logger.Logger
	onUpdate     chan string
	onTerminate  chan bool
}

func newPlugin(parentLogger *logger.Logger, path string, relativeName string, pluginsConstructor func() IPlugin) (*plugin, error) {
	fullPluginFileName := fmt.Sprintf("%v%v", path, relativeName)

	file, err := os.Stat(fullPluginFileName)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadFile(fullPluginFileName)
	if err != nil {
		return nil, err
	}

	plugin := &plugin{
		self:         pluginsConstructor(),
		relativeName: relativeName,
		modification: file.ModTime(),
		body:         string(body[:]),
		parentLogger: parentLogger,
		onUpdate:     make(chan string),
		onTerminate:  make(chan bool),
	}

	return plugin, nil
}

func (plugin *plugin) Go(current context.Context) {
	plugin.parentLogger.LogEvent(logger.EventTypeInfo, "pluginsManager", fmt.Sprintf("%v loading", plugin.relativeName))
	plugin.self.OnLoad(plugin.relativeName, plugin.body)
loop:
	for {
		select {
		case body := <-plugin.onUpdate:
			plugin.body = body
			plugin.parentLogger.LogEvent(logger.EventTypeInfo, "pluginsManager", fmt.Sprintf("%v updating", plugin.relativeName))
			plugin.self.OnUpdate(plugin.relativeName, plugin.body)
			break loop
		case <-plugin.onTerminate:
			break loop
		case <-current.OnDone():
			break loop
		}
	}
}

func (plugin *plugin) Dispose() {
	plugin.self.OnDispose(plugin.relativeName)
	plugin.parentLogger.LogEvent(logger.EventTypeInfo, "pluginsManager", fmt.Sprintf("%v unloaded", plugin.relativeName))
}
