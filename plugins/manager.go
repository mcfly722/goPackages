package plugins

import (
	"fmt"
	"io/ioutil"
	"os"
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

	manager.logger.LogEvent(logger.EventTypeInfo, "pluginsManager", fmt.Sprintf("started for %v (filter=%v)", manager.fullPluginsPath, manager.filter))

	plugins := map[string]*plugin{}

	duration := time.Duration(0) // first interval is zero, because we need to start immediately
loop:
	for {
		select {
		case <-time.After(duration): // we do not use Ticker here because it can't start immediately, always need to wait interval

			{ // load and update changed plugins
				duration = time.Duration(manager.updatePluginsIntervalSec) * time.Second // after first start we change interval dutation to seconds
				manager.logger.LogEvent(logger.EventTypeInfo, "pluginsManager", "looking for plugins changes...")

				pluginFiles, err := recursiveFilesSearch(manager.fullPluginsPath, manager.fullPluginsPath, manager.filter)
				if err != nil {
					manager.logger.LogEvent(logger.EventTypeException, "pluginsManager", err.Error())
					break
				}

				for _, pluginFile := range pluginFiles {
					if alreadyLoadedPlugin, ok := plugins[pluginFile]; !ok {
						// plugin file not loaded yet, we need load it
						plugin, err := newPlugin(manager.logger, manager.fullPluginsPath, pluginFile, manager.pluginsConstructor)
						if err != nil {
							manager.logger.LogEvent(logger.EventTypeException, "pluginsManager", err.Error())
						}
						plugins[pluginFile] = plugin
						current.NewContextFor(plugin)
					} else {
						// plugin file already loaded, we need check file modification date
						fullPluginFileName := fmt.Sprintf("%v%v", manager.fullPluginsPath, pluginFile)

						file, err := os.Stat(fullPluginFileName)
						if err != nil {
							manager.logger.LogEvent(logger.EventTypeException, "pluginsManager", err.Error())
						} else {

							if file.ModTime() != alreadyLoadedPlugin.modification {
								// plugin file was modified
								bodyBytes, err := ioutil.ReadFile(fullPluginFileName)
								if err != nil {
									manager.logger.LogEvent(logger.EventTypeException, "pluginsManager", err.Error())
								} else {
									// update plugin
									body := string(bodyBytes[:])
									alreadyLoadedPlugin.modification = file.ModTime()
									go func() {
										alreadyLoadedPlugin.onUpdate <- body
									}()
								}
							}
						}
					}

				}

				{ // check all loaded plugins on UpdateRequired
					for _, plugin := range plugins {
						if plugin.self.UpdateRequired() {
							plugin.self.OnUpdate(plugin.relativeName, plugin.body)
						}
					}
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
