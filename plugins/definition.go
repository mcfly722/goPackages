package plugins

import (
	"time"

	"github.com/mcfly722/goPackages/context"
)

// PluginDefinition ...
type PluginDefinition interface {
	Name() string
	Outdated() bool
}

type pluginDefinition struct {
	id               string
	modificationTime time.Time
	manager          *manager
	context          context.Context
}

func (pluginDefinition *pluginDefinition) Name() string {
	return pluginDefinition.id
}

func (pluginDefinition *pluginDefinition) Outdated() bool {
	return pluginDefinition.manager.definitionIsOutdated(pluginDefinition)
}
