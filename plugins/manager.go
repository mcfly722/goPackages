package plugins

import (
	"io/ioutil"
	"path/filepath"
	"strings"
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
	pluginsPath              string
	filter                   string
	rescanPluginsIntervalSec int
	pluginsConstructor       func(fullPath string) Plugin
	definitions              map[string]*pluginDefinition
	ready                    sync.Mutex
}

// NewPluginsManager ...
func NewPluginsManager(pluginsPath string, filter string, rescanPluginsIntervalSec int, pluginsConstructor func(fullPath string) Plugin) *Manager {
	pluginsManager := &Manager{
		pluginsPath:              pluginsPath,
		filter:                   filter,
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

				fullPluginsPath, err := filepath.Abs(manager.pluginsPath)
				if err != nil {
					current.Log(1, err.Error())
					break
				}

				current.Log(0, "check changes...")

				pluginFiles, err := recursiveFilesSearch(fullPluginsPath, fullPluginsPath, manager.filter)
				if err != nil {
					current.Log(2, err.Error())
					break
				}

				{ // load not existing plugins
					for _, pluginFile := range pluginFiles {
						if !manager.hasAlreadyRegisteredDefinition(pluginFile) {
							definition, err := manager.newPluginDefinition(fullPluginsPath, pluginFile, manager.rescanPluginsIntervalSec)
							if err != nil {
								current.Log(2, err.Error())
							} else {
								manager.registerNewPluginDefinition(definition)
								current.NewContextFor(definition, pluginFile, "definition")
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

func recursiveFilesSearch(rootPluginsPath string, currentFullPath string, filter string) ([]string, error) {
	result := []string{}

	files, err := ioutil.ReadDir(currentFullPath)
	if err != nil {
		return nil, err
	}

	path, err := filepath.Abs(currentFullPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			files, err := recursiveFilesSearch(rootPluginsPath, filepath.Join(path, file.Name()), filter)
			if err != nil {
				return nil, err
			}
			result = append(result, files...)
		} else {
			match, _ := filepath.Match(filter, file.Name())
			if match {
				relativeName := strings.TrimPrefix(filepath.Join(path, file.Name()), rootPluginsPath)

				relativeNameWithoutSlash := relativeName
				if len(relativeNameWithoutSlash) > 0 {
					if relativeNameWithoutSlash[0] == 92 {
						relativeNameWithoutSlash = relativeNameWithoutSlash[1:]
					}
				}

				result = append(result, relativeNameWithoutSlash)
			}
		}
	}
	return result, nil
}
