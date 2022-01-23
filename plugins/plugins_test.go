package plugins_test

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/mcfly722/goPackages/plugins"
)

type Plugin struct {
	*plugins.Plugin
	router *mux.Router
}

// OnLoad ...
func (plugin *Plugin) OnLoad() {
	backSlashed := strings.Replace(plugin.RelativeName, "\\", "/", -1)

	plugin.router.HandleFunc(backSlashed, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, fmt.Sprintf("<html>%v</html>", backSlashed))
	})

	log.Println(fmt.Sprintf("loaded plugin: %v (%v)", plugin.RelativeName, backSlashed))
}

// OnUpdate ...
func (plugin *Plugin) OnUpdate() {
	log.Println(fmt.Sprintf("updated plugin: %v", plugin.RelativeName))
}

// OnUnload ...
func (plugin *Plugin) OnUnload() {
	log.Println(fmt.Sprintf("uloaded plugin: %v", plugin.RelativeName))
}

func Test_AsWebServer(t *testing.T) {

	router := mux.NewRouter()

	pluginsConstructor := func(plugin *plugins.Plugin) plugins.IPlugin {
		return &Plugin{
			Plugin: plugin,
			router: router,
		}
	}

	pluginsManager, err := plugins.NewPluginsManager("", "*", 3, pluginsConstructor)
	if err != nil {
		t.Fatal(err)
	}

	if err := pluginsManager.Start(); err != nil {
		t.Fatal(err)
	}

	go func() {
		if err := http.ListenAndServe("127.0.0.1:8081", router); err != nil {
			t.Fatal(err)
		}
	}()

	time.Sleep(10 * time.Second)
}
