package plugins

import (
	"sync"
	"time"

	"github.com/mcfly722/goPackages/context"
)

// Plugin ...
type Plugin interface {
	context.ContextedInstance
}

// Manager ...
type Manager struct {
	provider                 Provider
	rescanPluginsIntervalSec int
	pluginsConstructor       func(fullPath string) Plugin
	definitions              map[string]*pluginDefinition
	ready                    sync.Mutex
}

// NewPluginsManager ...
func NewPluginsManager(provider Provider, rescanPluginsIntervalSec int, pluginsConstructor func(fullPath string) Plugin) *Manager {
	pluginsManager := &Manager{
		provider:                 provider,
		rescanPluginsIntervalSec: rescanPluginsIntervalSec,
		pluginsConstructor:       pluginsConstructor,
		definitions:              make(map[string]*pluginDefinition),
	}

	return pluginsManager
}

func (manager *Manager) registerNewPluginDefinition(definition *pluginDefinition) {
	manager.ready.Lock()
	defer manager.ready.Unlock()
	manager.definitions[definition.getID()] = definition
}

func (manager *Manager) unregisterPluginDefinition(definition *pluginDefinition) {
	manager.ready.Lock()
	defer manager.ready.Unlock()
	delete(manager.definitions, definition.getID())
}

func (manager *Manager) hasAlreadyRegisteredDefinition(definitionID string) bool {
	manager.ready.Lock()
	defer manager.ready.Unlock()
	_, ok := manager.definitions[definitionID]
	return ok
}

// Go ...
func (manager *Manager) Go(current context.Context) {
	duration := time.Duration(0) // first interval is zero, because we need to start immediately
loop:
	for {
		select {
		case <-time.After(duration): // we do not use Ticker here because it can't start immediately, always need to wait interval

			{ // rescan for not loaded yet plugins
				duration = time.Duration(manager.rescanPluginsIntervalSec) * time.Second // after first start we change interval dutation to seconds

				current.Log(0, "check changes...")

				plugins, err := manager.provider.GetPlugins()
				if err != nil {
					current.Log(2, err.Error())
					break
				}

				{ // load not existing plugins
					for _, plugin := range plugins {
						if !manager.hasAlreadyRegisteredDefinition(plugin) {

							path, err := manager.provider.GetCurrentPath()
							if err != nil {
								current.Log(2, err.Error())
							} else {
								definition, err := manager.newPluginDefinition(path, plugin, manager.rescanPluginsIntervalSec)
								if err != nil {
									current.Log(2, err.Error())
								} else {
									manager.registerNewPluginDefinition(definition)
									current.NewContextFor(definition, plugin, "definition")
								}
							}
						}
					}
				}

				break
			}
		case <-current.OnDone():
			break loop
		}
	}
}

// Dispose ...
func (manager *Manager) Dispose(current context.Context) {}

func contains(elems []string, v string) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}
