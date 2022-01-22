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
	fullPluginsPath          string
	pluginsPath              string
	updatePluginsIntervalSec int
	plugins                  map[string]*plugin

	ready sync.Mutex
}

type pluginer interface {
	OnLoad()
	OnUpdate()
	OnUnload()
}

type plugin struct {
	pluginer
	pluginsManager       *Manager
	relativeName         string
	lastModificationDate time.Time

	ready sync.Mutex
}

func (pluginsManager *Manager) newPlugin(relativeName string, lastModificationDate time.Time) *plugin {
	return &plugin{
		pluginsManager:       pluginsManager,
		relativeName:         relativeName,
		lastModificationDate: lastModificationDate,
	}
}

func (pluginsManager *Manager) loadOrUpdatePlugin(pluginRelativeName string, modificationTime time.Time) {
	newPlugins := []*plugin{}
	pluginsToUpdate := []*plugin{}

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
		plugin.OnLoad()
	}

	for _, plugin := range pluginsToUpdate {
		plugin.OnUpdate()
	}

}

// ToHTML ...
func (pluginsManager *Manager) ToHTML() string {

	tableHeader := "<tr><td align='center'><b>Plugin</b></td><td align='center'><b>Modification Date</b></td><tr>"

	tableContent := ""

	pluginsManager.ready.Lock()

	for _, plugin := range pluginsManager.plugins {
		tableContent += fmt.Sprintf("<tr><td>%v</td><td>%v</td></tr>", plugin.relativeName, plugin.lastModificationDate)
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
	pluginsToUnload := []*plugin{}

	pluginsManager.ready.Lock()

	for _, plugin := range pluginsManager.plugins {
		fullPluginFileName := filepath.Join(pluginsManager.fullPluginsPath, plugin.relativeName)
		_, err := os.Stat(fullPluginFileName)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				pluginsToUnload = append(pluginsToUnload, plugin)
			}
		}
	}

	for _, pluginToUnload := range pluginsToUnload {
		delete(pluginsManager.plugins, pluginToUnload.relativeName)
	}

	pluginsManager.ready.Unlock()

	for _, pluginToUnload := range pluginsToUnload {
		pluginToUnload.OnUnload()
	}
}

// NewManager ...
func NewManager(pluginsPath string, updatePluginsIntervalSec int, pluginer pluginer) (*Manager, error) {

	pluginsManager := &Manager{
		pluginsPath:              pluginsPath,
		updatePluginsIntervalSec: updatePluginsIntervalSec,
		plugins:                  map[string]*plugin{},
	}

	pluginsPathFull, err := filepath.Abs(pluginsManager.pluginsPath)
	if err != nil {
		return nil, err
	}
	pluginsManager.fullPluginsPath = pluginsPathFull

	go func() {

		for {

			err := pluginsManager.loadAndUpdateAllPlugins(pluginsPathFull)
			if err != nil {
				log.Println(err)
			}

			pluginsManager.unloadDeletedPlugins()

			time.Sleep(time.Duration(pluginsManager.updatePluginsIntervalSec) * time.Second)
		}

	}()

	return pluginsManager, nil
}
