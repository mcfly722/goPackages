package plugins

import (
	"io/ioutil"
	"path/filepath"
	"strings"
)

// Provider ...
type Provider interface {
	GetPlugins() ([]string, error)
	GetCurrentPath() (string, error)
}

// FromFilesProvider ...
type FromFilesProvider struct {
	pluginsPath string
	filter      string
}

// NewPluginsFromFilesProvider ...
func NewPluginsFromFilesProvider(pluginsPath string, filter string) Provider {

	return &FromFilesProvider{
		pluginsPath: pluginsPath,
		filter:      filter,
	}
}

// GetPlugins ...
func (provider *FromFilesProvider) GetPlugins() ([]string, error) {

	fullPluginsPath, err := provider.GetCurrentPath()
	if err != nil {
		return nil, err
	}

	return recursiveFilesSearch(fullPluginsPath, fullPluginsPath, provider.filter)
}

// GetCurrentPath ...
func (provider *FromFilesProvider) GetCurrentPath() (string, error) {
	return filepath.Abs(provider.pluginsPath)
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
