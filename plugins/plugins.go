package plugins

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Manager ...
type Manager struct {
	pluginsConstructor       func(*Plugin) IPlugin
	fullPluginsPath          string
	pluginsPath              string
	updatePluginsIntervalSec int
	plugins                  map[string]*Plugin
	started                  bool

	ready sync.Mutex
}

// Plugin ...
type Plugin struct {
	PluginsManager       *Manager
	RelativeName         string
	lastModificationDate time.Time

	actions IPlugin

	ready sync.Mutex
}

// IPlugin ...
type IPlugin interface {
	OnLoad()
	OnUpdate()
	OnUnload()
}

func (pluginsManager *Manager) newPlugin(relativeName string, lastModificationDate time.Time) *Plugin {
	plugin := &Plugin{
		PluginsManager:       pluginsManager,
		RelativeName:         relativeName,
		lastModificationDate: lastModificationDate,
	}

	plugin.actions = pluginsManager.pluginsConstructor(plugin)

	return plugin
}

func (pluginsManager *Manager) loadOrUpdatePlugin(pluginRelativeName string, modificationTime time.Time) {
	newPlugins := []*Plugin{}
	pluginsToUpdate := []*Plugin{}

	pluginsManager.ready.Lock()

	plugin, exist := pluginsManager.plugins[pluginRelativeName]
	if !exist {

		// Add new plugin
		plugin := pluginsManager.newPlugin(pluginRelativeName, modificationTime)
		pluginsManager.plugins[pluginRelativeName] = plugin
		newPlugins = append(newPlugins, plugin)

	} else {

		// Update existing plugin (if required)
		if modificationTime.After(plugin.lastModificationDate) {
			plugin.lastModificationDate = modificationTime
			pluginsToUpdate = append(pluginsToUpdate, plugin)
		}

	}

	pluginsManager.ready.Unlock()

	for _, plugin := range newPlugins {
		plugin.actions.OnLoad()
	}

	for _, plugin := range pluginsToUpdate {
		plugin.actions.OnUpdate()
	}

}

// ToHTML ...
func (pluginsManager *Manager) ToHTML() string {

	tableHeader := "<tr><td align='center'><b>Plugin</b></td><td align='center'><b>Modification Date</b></td><tr>"

	tableContent := ""

	pluginsManager.ready.Lock()

	for _, plugin := range pluginsManager.plugins {
		tableContent += fmt.Sprintf("<tr><td>%v</td><td>%v</td></tr>", plugin.RelativeName, plugin.lastModificationDate)
	}

	pluginsManager.ready.Unlock()

	return fmt.Sprintf("<table border=1px cellpadding='10' cellspacing='0'>%v%v</table>", tableHeader, tableContent)
}

func (pluginsManager *Manager) loadAndUpdateAllPlugins(currentFullPath string) error {

	files, err := ioutil.ReadDir(currentFullPath)
	if err != nil {
		return err
	}

	path, err := filepath.Abs(currentFullPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			pluginsManager.loadAndUpdateAllPlugins(filepath.Join(path, file.Name()))
		} else {

			relativeName := strings.TrimPrefix(filepath.Join(path, file.Name()), pluginsManager.fullPluginsPath)
			pluginsManager.loadOrUpdatePlugin(relativeName, file.ModTime())
		}
	}
	return nil
}

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

// NewPluginsManager ...
func NewPluginsManager(pluginsPath string, updatePluginsIntervalSec int, pluginsConstructor func(*Plugin) IPlugin) (*Manager, error) {

	pluginsManager := &Manager{
		pluginsConstructor:       pluginsConstructor,
		pluginsPath:              pluginsPath,
		updatePluginsIntervalSec: updatePluginsIntervalSec,
		plugins:                  map[string]*Plugin{},
		started:                  false,
	}

	pluginsPathFull, err := filepath.Abs(pluginsManager.pluginsPath)
	if err != nil {
		return nil, err
	}
	pluginsManager.fullPluginsPath = pluginsPathFull

	return pluginsManager, nil
}

// Start ...
func (pluginsManager *Manager) Start() error {
	if pluginsManager.started == true {
		return errors.New("Plugins Manager is already started")
	}

	go func() {

		for {
			err := pluginsManager.loadAndUpdateAllPlugins(pluginsManager.fullPluginsPath)
			if err != nil {
				log.Println(err)
			}

			pluginsManager.unloadDeletedPlugins()

			time.Sleep(time.Duration(pluginsManager.updatePluginsIntervalSec) * time.Second)
		}
	}()

	return nil
}
