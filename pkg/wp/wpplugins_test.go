package wp

import (
	"reflect"
	"strings"
	"testing"
)

var errChan = make(chan error)

func TestNewPlugins(t *testing.T) {
	testPlugins, err := NewPlugins()
	if err != nil {
		t.Errorf("NewPlugins() call error with error:\n%s", err)
	}
	pluginType := reflect.TypeOf(testPlugins)
	if pluginType.String() != "*wp.Plugins" {
		t.Errorf("TestPlugins did not initialize as type *Plugins.\nWant: *Plugins, Got: %s\n", pluginType)
	}
	if testPlugins.URI != pluginAPI || testPlugins.Info.Pages < 1 {
		t.Errorf("Test Plugins variable did not initialize with proper information.\nGot: %#v\n", testPlugins)
	}
}

func TestAddPlugins(t *testing.T) {
	plugins, err := NewPlugins()
	if err != nil {
		t.Errorf("NewPlugins() call error with error:\n%s", err)
	}
	beforePlugins := len(plugins.Plugins)
	plugins.AddPlugins(errChan)
	afterPlugins := len(plugins.Plugins)

	if afterPlugins <= beforePlugins {
		t.Errorf("Plugin count did not increase after calling .AddPlugins()\nbefore: %d, after: %d\n", beforePlugins, afterPlugins)
	}
}

func TestRemovePlugin(t *testing.T) {
	plugins, err := NewPlugins()
	if err != nil {
		t.Errorf("NewPlugins() call error with error:\n%s", err)
	}
	plugins.AddPlugins(errChan)
	beforeRemove := len(plugins.Plugins)
	plugins.RemovePlugin(1)
	afterRemove := len(plugins.Plugins)
	if afterRemove != beforeRemove-1 {
		t.Errorf("plugins.RemovePlugin(1) did not remove one plugin.\nPrevious Length: %d, After Length: %d\n", beforeRemove, afterRemove)
	}

}

func TestPluginSetOutPath(t *testing.T) {
	plugins, err := NewPlugins()
	if err != nil {
		t.Errorf("NewPlugins() call error with error:\n%s", err)
	}
	plugins.AddPlugins(errChan)
	plugin := plugins.Plugins[1]
	plugin.setOutPath()
	if !(strings.Contains(plugin.OutPath, "current")) {
		t.Errorf("plugin.SetOutPath() did not set a path to the current/ folder.\nGot: %s", plugin.OutPath)
	}
}
