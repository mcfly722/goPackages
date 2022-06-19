package plugins

import (
	"path/filepath"
	"time"

	"github.com/mcfly722/goPackages/context"
)

// PluginDefinition ...
type PluginDefinition interface {
	Name() string
	Outdated() bool
	GetBody() string
	GetResource(path string) (*[]byte, error)
}

type pluginDefinition struct {
	id               string
	modificationTime time.Time
	manager          *manager
	context          context.Context
	body             string
}

func (pluginDefinition *pluginDefinition) Name() string {
	return pluginDefinition.id
}

func (pluginDefinition *pluginDefinition) GetBody() string {
	return pluginDefinition.body
}

func (pluginDefinition *pluginDefinition) Outdated() bool {
	return pluginDefinition.manager.definitionIsOutdated(pluginDefinition)
}

func (pluginDefinition *pluginDefinition) GetResource(path string) (*[]byte, error) {
	return pluginDefinition.manager.getResource(filepath.Join(filepath.Dir(pluginDefinition.id), path))
}
