package plugins

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/mcfly722/goPackages/context"
)

// Constructor ...
type Constructor func(definition PluginDefinition) context.ContextedInstance

type manager struct {
	provider          Provider
	rescanIntervalSec int
	constructor       Constructor
	definitions       map[string]*pluginDefinition
	ready             sync.Mutex
}

// NewPluginsManager ...1
func NewPluginsManager(provider Provider, rescanIntervalSec int, constructor Constructor) context.ContextedInstance {
	return &manager{
		provider:          provider,
		rescanIntervalSec: rescanIntervalSec,
		constructor:       constructor,
		definitions:       make(map[string]*pluginDefinition),
	}
}

func (manager *manager) definitionIsOutdated(definition *pluginDefinition) bool {
	manager.ready.Lock()
	defer manager.ready.Unlock()
	if cachedDefinition, ok := manager.definitions[definition.id]; ok {
		if cachedDefinition.modificationTime == definition.modificationTime {
			return false
		}
	}
	return true
}

func (manager *manager) getRegisteredDefinition(definitionID string) (*pluginDefinition, bool) {
	manager.ready.Lock()
	defer manager.ready.Unlock()
	definition, found := manager.definitions[definitionID]
	return definition, found
}

func (manager *manager) registerNewPluginDefinition(definition *pluginDefinition) {
	manager.ready.Lock()
	defer manager.ready.Unlock()
	manager.definitions[definition.id] = definition
}

func (manager *manager) unregisterPluginDefinition(definitionID string) *pluginDefinition {
	manager.ready.Lock()
	defer manager.ready.Unlock()
	if definition, found := manager.definitions[definitionID]; found {
		delete(manager.definitions, definitionID)
		return definition
	}
	return nil
}

// Go ...
func (manager *manager) Go(current context.Context) {
	duration := time.Duration(0) // first interval is zero, because we need to start immediately
loop:
	for {
		select {
		case <-time.After(duration): // we do not use Ticker here because it can't start immediately, always need to wait interval
			{ // rescan for not loaded yet plugins
				duration = time.Duration(manager.rescanIntervalSec) * time.Second // after first start we change interval dutation to seconds

				current.Log(0, "check changes...")

				plugins, err := manager.provider.GetPlugins()
				if err != nil {
					current.Log(2, err.Error())
					break
				}

				pluginsModificationTimes := make(map[string]time.Time)
				{ // collect plugins modification Times
					for _, plugin := range plugins {
						modificationTime, err := manager.provider.GetPluginModificationTime(plugin)
						if err != nil {
							break
						}
						pluginsModificationTimes[plugin] = modificationTime
					}
				}

				{ // delete not existing or outdated definitions
					manager.ready.Lock()
					definitionsForDeleting := []string{}

					for plugin, definition := range manager.definitions {
						if modificationTime, found := pluginsModificationTimes[plugin]; !found {
							definitionsForDeleting = append(definitionsForDeleting, plugin)
						} else {
							if definition.modificationTime != modificationTime {
								definitionsForDeleting = append(definitionsForDeleting, plugin)
							}
						}
					}
					manager.ready.Unlock()

					for _, definitionForDeleting := range definitionsForDeleting {
						unregisteredDefinition := manager.unregisterPluginDefinition(definitionForDeleting)
						unregisteredDefinition.context.Wait()
					}

				}

				{ // load new definitions
					for _, plugin := range plugins {
						if _, found := manager.getRegisteredDefinition(plugin); !found {

							body, err := manager.provider.GetResource(plugin)
							if err != nil {
								current.Log(1, err.Error())
							} else {

								definition := &pluginDefinition{
									id:               plugin,
									modificationTime: pluginsModificationTimes[plugin],
									manager:          manager,
									body:             string(*body),
								}

								manager.registerNewPluginDefinition(definition)
								pluginInstance := manager.constructor(definition)

								definition.context = current.NewContextFor(pluginInstance, fmt.Sprintf("%v(%v)", definition.Name(), rand.Intn(9999999)), "definition")
							}
						}
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
func (manager *manager) Dispose(current context.Context) {}

func contains(elems []string, v string) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}
