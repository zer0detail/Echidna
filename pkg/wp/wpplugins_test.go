package wp

import (
	"reflect"
	"strings"
	"testing"
)

var plugins = NewPlugins()

func TestNewPlugins(t *testing.T) {
	testPlugins := NewPlugins()
	pluginType := reflect.TypeOf(testPlugins)
	if pluginType.String() != "*wp.Plugins" {
		t.Errorf("TestPlugins did not initialize as type *Plugins.\nWant: *Plugins, Got: %s\n", pluginType)
	}
	if testPlugins.URI != pluginAPI || testPlugins.Info.Pages < 1 {
		t.Errorf("Test Plugins variable did not initialize with proper information.\nGot: %#v\n", testPlugins)
	}
}

func TestAddPlugins(t *testing.T) {
	beforePlugins := len(plugins.Plugins)
	plugins.AddPlugins()
	afterPlugins := len(plugins.Plugins)

	if afterPlugins <= beforePlugins {
		t.Errorf("Plugin count did not increase after calling .AddPlugins()\nbefore: %d, after: %d\n", beforePlugins, afterPlugins)
	}
}

func TestRemovePlugin(t *testing.T) {
	plugins.AddPlugins()
	beforeRemove := len(plugins.Plugins)
	plugins.RemovePlugin(1)
	afterRemove := len(plugins.Plugins)
	if afterRemove != beforeRemove-1 {
		t.Errorf("plugins.RemovePlugin(1) did not remove one plugin.\nPrevious Length: %d, After Length: %d\n", beforeRemove, afterRemove)
	}

}

func TestPluginSetOutPath(t *testing.T) {
	plugins.AddPlugins()
	plugin := plugins.Plugins[1]
	plugin.setOutPath()
	if !(strings.Contains(plugin.OutPath, "current")) {
		t.Errorf("plugin.SetOutPath() did not set a path to the current/ folder.\nGot: %s", plugin.OutPath)
	}
}
