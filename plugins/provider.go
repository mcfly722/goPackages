package plugins

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Provider ...
type Provider interface {
	GetPlugins() ([]string, error)
	GetPluginModificationTime(pluginPath string) (time.Time, error)
	GetResource(path string) (*[]byte, error)
}

type fromFilesProvider struct {
	pluginsPath string
	filter      string
}

// NewPluginsFromFilesProvider ...
func NewPluginsFromFilesProvider(pluginsPath string, filter string) Provider {
	return &fromFilesProvider{
		pluginsPath: pluginsPath,
		filter:      filter,
	}
}

func (provider *fromFilesProvider) GetPluginModificationTime(pluginPath string) (time.Time, error) {
	file, err := os.Stat(filepath.Join(provider.pluginsPath, pluginPath))
	if err != nil {
		return time.Time{}, err
	}
	return file.ModTime(), nil
}

// GetPlugins ...
func (provider *fromFilesProvider) GetPlugins() ([]string, error) {

	fullPluginsPath, err := filepath.Abs(provider.pluginsPath)
	if err != nil {
		return nil, err
	}

	return recursiveFilesSearch(fullPluginsPath, fullPluginsPath, provider.filter)
}

func (provider *fromFilesProvider) GetResource(path string) (*[]byte, error) {

	data, err := os.ReadFile(filepath.Join(provider.pluginsPath, path))
	if err != nil {
		return nil, err
	}

	return &data, nil
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
