package plugins

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mcfly722/goPackages/context"
)

// Plugin ...
type Plugin interface {
	context.ContextedInstance
	Terminate()
}

type pluginConfig struct {
	plugin           Plugin
	context          context.Context
	modificationTime time.Time
}

// Manager ...
type Manager struct {
	pluginsPath              string
	filter                   string
	updatePluginsIntervalSec int
	pluginsConstructor       func(fullPath string) Plugin
	pluginsConfigurations    map[string]*pluginConfig
}

func newPluginConfiguration(fullPluginFileName string, pluginsConstructor func(fullPath string) Plugin) (*pluginConfig, error) {
	file, err := os.Stat(fullPluginFileName)
	if err != nil {
		return nil, err
	}

	return &pluginConfig{
		plugin:           pluginsConstructor(fullPluginFileName),
		modificationTime: file.ModTime(),
	}, nil
}

// NewPluginsManager ...
func NewPluginsManager(pluginsPath string, filter string, updatePluginsIntervalSec int, pluginsConstructor func(fullPath string) Plugin) *Manager {
	pluginsManager := &Manager{
		pluginsPath:              pluginsPath,
		filter:                   filter,
		updatePluginsIntervalSec: updatePluginsIntervalSec,
		pluginsConstructor:       pluginsConstructor,
		pluginsConfigurations:    make(map[string]*pluginConfig),
	}

	return pluginsManager
}

// Go ...
func (manager *Manager) Go(current context.Context) {
	duration := time.Duration(0) // first interval is zero, because we need to start immediately
loop:
	for {
		select {
		case <-time.After(duration): // we do not use Ticker here because it can't start immediately, always need to wait interval

			{ // load and update changed plugins
				duration = time.Duration(manager.updatePluginsIntervalSec) * time.Second // after first start we change interval dutation to seconds

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

				pluginsToTerminate := []string{}

				{ // delete updated plugins
					for _, pluginFile := range pluginFiles {
						if config, ok := manager.pluginsConfigurations[pluginFile]; ok {
							// plugin file already loaded, we need check file modification date and if it different, just delete it from list
							file, err := os.Stat(pluginFile)
							if err != nil {
								current.Log(3, err.Error())
							} else {
								if file.ModTime() != config.modificationTime {
									pluginsToTerminate = append(pluginsToTerminate, pluginFile)
								}
							}
						}
					}
				}

				{ // unload deleted plugins
					for pluginFileName := range manager.pluginsConfigurations {
						if !contains(pluginFiles, pluginFileName) {
							pluginsToTerminate = append(pluginsToTerminate, pluginFileName)
						}
					}
					for _, pluginToTerminate := range pluginsToTerminate {
						config := manager.pluginsConfigurations[pluginToTerminate]
						{ // send termination and wait till plugin would be terminated
							go func() {
								config.plugin.Terminate()
							}()
							config.context.Wait()
						}
						delete(manager.pluginsConfigurations, pluginToTerminate)
					}
				}

				{ // load not existing plugins
					for _, pluginFile := range pluginFiles {
						if _, ok := manager.pluginsConfigurations[pluginFile]; !ok {

							config, err := newPluginConfiguration(pluginFile, manager.pluginsConstructor)
							if err != nil {
								current.Log(2, err.Error())
							} else {
								config.context = current.NewContextFor(config.plugin, fmt.Sprintf("%v[%v]", pluginFile, rand.Intn(99999999)), "plugin")
								manager.pluginsConfigurations[pluginFile] = config
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
