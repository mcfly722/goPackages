package plugins

import (
	"os"
	"path"
	"time"

	"github.com/mcfly722/goPackages/context"
)

// PluginDefinition ...
type pluginDefinition struct {
	manager           *Manager
	pluginsPath       string
	fileName          string
	updateIntervalSec int
	modificationTime  time.Time
}

func (manager *Manager) newPluginDefinition(pluginsPath string, fileName string, updateIntervalSec int) (*pluginDefinition, error) {

	fullDefinitionFileName := path.Join(pluginsPath, fileName)

	file, err := os.Stat(fullDefinitionFileName)
	if err != nil {
		return nil, err
	}

	pluginDefinition := &pluginDefinition{
		manager:           manager,
		pluginsPath:       pluginsPath,
		fileName:          fileName,
		updateIntervalSec: updateIntervalSec,
		modificationTime:  file.ModTime(),
	}

	return pluginDefinition, nil
}

func (definition *pluginDefinition) getID() string {
	return definition.fileName
}

// Go ...
func (definition *pluginDefinition) Go(current context.Context) {
	duration := time.Duration(0) // first interval is zero, because we need to start immediately
loop:
	for {
		select {
		case <-time.After(duration): // we do not use Ticker here because it can't start immediately, always need to wait whole interval
			duration = time.Duration(definition.updateIntervalSec) * time.Second // after first start we change interval dutation to required one

			{ // if plugin definition have been changed or deleted, just unload current plugin

				file, err := os.Stat(definition.fileName)
				if err != nil {
					break loop
				} else {
					if file.ModTime() != definition.modificationTime {
						break loop
					}
				}
			}

			break
		case <-current.OnDone():
			break loop
		}
	}
}

// Dispose ...
func (definition *pluginDefinition) Dispose(current context.Context) {
	definition.manager.unregisterPluginDefinition(definition)
}
