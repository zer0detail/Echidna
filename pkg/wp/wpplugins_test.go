package wp

import (
	"context"
	"reflect"
	"testing"
)

//var client = &echidnatesting.MockClient{}
//var errChan = make(chan error)
var ctx, _ = context.WithCancel(context.Background())

func TestNewPlugins(t *testing.T) {
	testPlugins, err := NewPlugins(ctx)
	if err != nil {
		t.Errorf("NewPlugins() call error with error:\n%s", err)
	}
	pluginType := reflect.TypeOf(testPlugins)
	if pluginType.String() != "*wp.Plugins" {
		t.Errorf("TestPlugins did not initialize as type *Plugins.\nWant: *Plugins, Got: %s\n", pluginType)
	}
	if testPlugins.URI != pluginAPI || testPlugins.Info.Page != 1 {
		t.Errorf("Test Plugins variable did not initialize with proper information.\nGot: %#v\n", testPlugins)
	}
}

// func TestAddPlugins(t *testing.T) {
// 	plugins, err := NewPlugins(ctx)
// 	if err != nil {
// 		t.Errorf("NewPlugins() call error with error:\n%s", err)
// 	}

// 	dummyBody := []byte(client.DummyBody())
// 	var wg sync.WaitGroup
// 	// we run addPlugins as a goroutine because thats how its run in the
// 	// main program.
// 	beforePlugins := len(plugins.Plugins)
// 	wg.Add(1)
// 	go plugins.addPlugins(ctx, dummyBody, errChan, &wg)
// 	wg.Done()
// 	fmt.Printf("plugins: %v\n", plugins.Plugins)
// 	afterPlugins := len(plugins.Plugins)

// 	if afterPlugins <= beforePlugins {
// 		t.Errorf("Plugin count did not increase after calling .AddPlugins()\nbefore: %d, after: %d\n", beforePlugins, afterPlugins)
// 	}
// }

// func TestRemovePlugin(t *testing.T) {
// 	plugins, err := NewPlugins(ctx)
// 	if err != nil {
// 		t.Errorf("NewPlugins() call error with error:\n%s", err)
// 	}
// 	plugins.client = client
// 	client.SetReply(200, client.DummyBody(), "")
// 	plugins.addPlugins(ctx, errChan)
// 	beforeRemove := len(plugins.Plugins)
// 	plugins.RemovePlugin(1)
// 	afterRemove := len(plugins.Plugins)
// 	if afterRemove != beforeRemove-1 {
// 		t.Errorf("plugins.RemovePlugin(1) did not remove one plugin.\nPrevious Length: %d, After Length: %d\n", beforeRemove, afterRemove)
// 	}

// }

// func TestPluginSetOutPath(t *testing.T) {
// 	plugins, err := NewPlugins(ctx)
// 	if err != nil {
// 		t.Errorf("NewPlugins() call error with error:\n%s", err)
// 	}
// 	plugins.client = client
// 	client.SetReply(200, client.DummyBody(), "")
// 	plugins.addPlugins(ctx, errChan)
// 	plugin := plugins.Plugins[1]
// 	err = plugin.setOutPath()
// 	if err != nil {
// 		t.Errorf("plugin.SetOutPath() failed with unexpected error:\n%s", err)
// 	}
// 	if !(strings.Contains(plugin.OutPath, "current")) {
// 		t.Errorf("plugin.SetOutPath() did not set a path to the current/ folder.\nGot: %s", plugin.OutPath)
// 	}
// }
