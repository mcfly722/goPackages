package plugins

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/mcfly722/goPackages/context"
	"github.com/mcfly722/goPackages/logger"
)

// Manager ...
type Manager struct {
	pluginsConstructor       func() IPlugin
	pluginsPath              string
	fullPluginsPath          string
	updatePluginsIntervalSec int
	filter                   string
	plugins                  map[string]*plugin
	logger                   *logger.Logger
}

// NewPluginsManager ...
func NewPluginsManager(pluginsPath string, filter string, updatePluginsIntervalSec int, pluginsConstructor func() IPlugin) (*Manager, error) {

	pluginsManager := &Manager{
		pluginsConstructor:       pluginsConstructor,
		pluginsPath:              pluginsPath,
		updatePluginsIntervalSec: updatePluginsIntervalSec,
		filter:                   filter,
	}

	pluginsPathFull, err := filepath.Abs(pluginsManager.pluginsPath)
	if err != nil {
		return nil, err
	}
	pluginsManager.fullPluginsPath = pluginsPathFull

	return pluginsManager, nil
}

// SetLogger ...
func (manager *Manager) SetLogger(logger *logger.Logger) {
	manager.logger = logger
}

// Go ...
func (manager *Manager) Go(current context.Context) {

	if manager.logger == nil {
		manager.logger = logger.NewLogger(5)
	}

	ticker := time.NewTicker(time.Duration(manager.updatePluginsIntervalSec) * time.Second)

	manager.logger.LogEvent(logger.EventTypeInfo, "pluginsManager", fmt.Sprintf("started for %v (filter=%v)", manager.fullPluginsPath, manager.filter))

	plugins := map[string]*plugin{}

loop:
	for {
		select {
		case <-ticker.C:
			{
				manager.logger.LogEvent(logger.EventTypeInfo, "pluginsManager", "looking for plugins...")

				pluginFiles, err := recursiveFilesSearch(manager.fullPluginsPath, manager.fullPluginsPath, manager.filter)
				if err != nil {
					manager.logger.LogEvent(logger.EventTypeException, "pluginsManager", err.Error())
					break
				}

				for _, pluginFile := range pluginFiles {
					if _, ok := plugins[pluginFile]; !ok {
						// init
						plugin, err := newPlugin(manager.logger, manager.fullPluginsPath, pluginFile, manager.pluginsConstructor)
						if err != nil {
							manager.logger.LogEvent(logger.EventTypeException, "pluginsManager", err.Error())
						}
						plugins[pluginFile] = plugin
						current.NewContextFor(plugin)
					} else {
						// check mod time
					}

					//manager.logger.LogEvent(logger.EventTypeInfo, "pluginsManager", pluginFile)
				}

				{ // unload deleted plugins
					pluginsToTerminate := []string{}

					for pluginFileName := range plugins {
						if !contains(pluginFiles, pluginFileName) {
							pluginsToTerminate = append(pluginsToTerminate, pluginFileName)
						}
					}

					for _, pluginToTerminate := range pluginsToTerminate {
						plugin := plugins[pluginToTerminate]
						delete(plugins, pluginToTerminate)
						go func() {
							plugin.onTerminate <- true
						}()
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
func (manager *Manager) Dispose() {
	manager.logger.LogEvent(logger.EventTypeInfo, "pluginsManager", "disposed")
}

// -------------------------------------------------------------------------------------------------
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
				result = append(result, relativeName)
			}
		}
	}
	return result, nil
}

/*
func (pluginsManager *Manager) unloadDeletedPlugins() {
	pluginsToUnload := []*Plugin{}

	pluginsManager.ready.Lock()

	for _, plugin := range pluginsManager.plugins {
		fullPluginFileName := filepath.Join(pluginsManager.fullPluginsPath, plugin.RelativeName)
		_, err := os.Stat(fullPluginFileName)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				pluginsToUnload = append(pluginsToUnload, plugin)
			}
		}
	}

	for _, pluginToUnload := range pluginsToUnload {
		delete(pluginsManager.plugins, pluginToUnload.RelativeName)
	}

	pluginsManager.ready.Unlock()

	for _, pluginToUnload := range pluginsToUnload {
		pluginToUnload.actions.OnUnload()
	}
}
*/
