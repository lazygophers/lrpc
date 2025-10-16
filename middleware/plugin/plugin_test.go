package plugin

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestPluginErrors(t *testing.T) {
	t.Run("error messages", func(t *testing.T) {
		assert.Equal(t, "plugin not found", ErrPluginNotFound.Error())
		assert.Equal(t, "plugin already exists", ErrPluginAlreadyExists.Error())
		assert.Equal(t, "plugin is disabled", ErrPluginDisabled.Error())
	})
}

func TestNewBasePlugin(t *testing.T) {
	t.Run("create new base plugin", func(t *testing.T) {
		plugin := NewBasePlugin("test", "1.0.0")

		assert.NotNil(t, plugin)
		assert.Equal(t, "test", plugin.Name())
		assert.Equal(t, "1.0.0", plugin.Version())
		assert.Equal(t, StatusUninitialized, plugin.Status())
	})
}

func TestBasePluginInit(t *testing.T) {
	t.Run("initialize plugin", func(t *testing.T) {
		plugin := NewBasePlugin("test", "1.0.0")

		config := map[string]interface{}{
			"key": "value",
		}

		err := plugin.Init(config)

		require.NoError(t, err)
		assert.Equal(t, StatusInitialized, plugin.Status())
		assert.Equal(t, config, plugin.Config())
	})
}

func TestBasePluginLifecycle(t *testing.T) {
	t.Run("complete lifecycle", func(t *testing.T) {
		plugin := NewBasePlugin("test", "1.0.0")

		// Initial state
		assert.Equal(t, StatusUninitialized, plugin.Status())

		// Init
		err := plugin.Init(nil)
		require.NoError(t, err)
		assert.Equal(t, StatusInitialized, plugin.Status())

		// Start
		err = plugin.Start()
		require.NoError(t, err)
		assert.Equal(t, StatusRunning, plugin.Status())

		// Stop
		err = plugin.Stop()
		require.NoError(t, err)
		assert.Equal(t, StatusStopped, plugin.Status())
	})
}

func TestBasePluginSetStatus(t *testing.T) {
	t.Run("set custom status", func(t *testing.T) {
		plugin := NewBasePlugin("test", "1.0.0")

		plugin.SetStatus(StatusError)

		assert.Equal(t, StatusError, plugin.Status())
	})
}

func TestBasePluginGetConfigAs(t *testing.T) {
	t.Run("get config as map", func(t *testing.T) {
		plugin := NewBasePlugin("test", "1.0.0")

		config := map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		}
		plugin.Init(config)

		var out map[string]interface{}
		err := plugin.GetConfigAs(&out)

		require.NoError(t, err)
		assert.Equal(t, config, out)
	})

	t.Run("return nil for empty config", func(t *testing.T) {
		plugin := NewBasePlugin("test", "1.0.0")

		var out map[string]interface{}
		err := plugin.GetConfigAs(&out)

		assert.NoError(t, err)
	})
}

func TestNewManager(t *testing.T) {
	t.Run("create new manager", func(t *testing.T) {
		manager := NewManager()

		assert.NotNil(t, manager)
		assert.NotNil(t, manager.plugins)
		assert.NotNil(t, manager.enabled)
		assert.Empty(t, manager.List())
	})
}

func TestManagerRegister(t *testing.T) {
	t.Run("register plugin", func(t *testing.T) {
		manager := NewManager()
		plugin := NewBasePlugin("test", "1.0.0")

		err := manager.Register(plugin)

		require.NoError(t, err)
		assert.Contains(t, manager.List(), "test")
		assert.True(t, manager.IsEnabled("test"))
	})

	t.Run("error when plugin already exists", func(t *testing.T) {
		manager := NewManager()
		plugin1 := NewBasePlugin("test", "1.0.0")
		plugin2 := NewBasePlugin("test", "2.0.0")

		manager.Register(plugin1)
		err := manager.Register(plugin2)

		assert.ErrorIs(t, err, ErrPluginAlreadyExists)
	})
}

func TestManagerUnregister(t *testing.T) {
	t.Run("unregister plugin", func(t *testing.T) {
		manager := NewManager()
		plugin := NewBasePlugin("test", "1.0.0")

		manager.Register(plugin)
		err := manager.Unregister("test")

		require.NoError(t, err)
		assert.NotContains(t, manager.List(), "test")
	})

	t.Run("stop plugin before unregister", func(t *testing.T) {
		manager := NewManager()
		plugin := NewBasePlugin("test", "1.0.0")

		manager.Register(plugin)
		plugin.Start()

		err := manager.Unregister("test")

		require.NoError(t, err)
		assert.Equal(t, StatusStopped, plugin.Status())
	})

	t.Run("error when plugin not found", func(t *testing.T) {
		manager := NewManager()

		err := manager.Unregister("nonexistent")

		assert.ErrorIs(t, err, ErrPluginNotFound)
	})
}

func TestManagerGet(t *testing.T) {
	t.Run("get plugin by name", func(t *testing.T) {
		manager := NewManager()
		plugin := NewBasePlugin("test", "1.0.0")

		manager.Register(plugin)

		retrieved, err := manager.Get("test")

		require.NoError(t, err)
		assert.Equal(t, plugin, retrieved)
	})

	t.Run("error when plugin not found", func(t *testing.T) {
		manager := NewManager()

		_, err := manager.Get("nonexistent")

		assert.ErrorIs(t, err, ErrPluginNotFound)
	})
}

func TestManagerEnableDisable(t *testing.T) {
	t.Run("disable and enable plugin", func(t *testing.T) {
		manager := NewManager()
		plugin := NewBasePlugin("test", "1.0.0")

		manager.Register(plugin)
		assert.True(t, manager.IsEnabled("test"))

		err := manager.Disable("test")
		require.NoError(t, err)
		assert.False(t, manager.IsEnabled("test"))

		err = manager.Enable("test")
		require.NoError(t, err)
		assert.True(t, manager.IsEnabled("test"))
	})

	t.Run("error when disabling nonexistent plugin", func(t *testing.T) {
		manager := NewManager()

		err := manager.Disable("nonexistent")

		assert.ErrorIs(t, err, ErrPluginNotFound)
	})

	t.Run("error when enabling nonexistent plugin", func(t *testing.T) {
		manager := NewManager()

		err := manager.Enable("nonexistent")

		assert.ErrorIs(t, err, ErrPluginNotFound)
	})
}

func TestManagerList(t *testing.T) {
	t.Run("list all plugins", func(t *testing.T) {
		manager := NewManager()
		manager.Register(NewBasePlugin("plugin1", "1.0.0"))
		manager.Register(NewBasePlugin("plugin2", "2.0.0"))
		manager.Register(NewBasePlugin("plugin3", "3.0.0"))

		list := manager.List()

		assert.Len(t, list, 3)
		assert.Contains(t, list, "plugin1")
		assert.Contains(t, list, "plugin2")
		assert.Contains(t, list, "plugin3")
	})
}

func TestManagerInitAll(t *testing.T) {
	t.Run("initialize all enabled plugins", func(t *testing.T) {
		manager := NewManager()
		plugin1 := NewBasePlugin("plugin1", "1.0.0")
		plugin2 := NewBasePlugin("plugin2", "2.0.0")

		manager.Register(plugin1)
		manager.Register(plugin2)

		configs := map[string]interface{}{
			"plugin1": map[string]interface{}{"key1": "value1"},
			"plugin2": map[string]interface{}{"key2": "value2"},
		}

		err := manager.InitAll(configs)

		require.NoError(t, err)
		assert.Equal(t, StatusInitialized, plugin1.Status())
		assert.Equal(t, StatusInitialized, plugin2.Status())
	})

	t.Run("skip disabled plugins", func(t *testing.T) {
		manager := NewManager()
		plugin1 := NewBasePlugin("plugin1", "1.0.0")
		plugin2 := NewBasePlugin("plugin2", "2.0.0")

		manager.Register(plugin1)
		manager.Register(plugin2)
		manager.Disable("plugin2")

		err := manager.InitAll(nil)

		require.NoError(t, err)
		assert.Equal(t, StatusInitialized, plugin1.Status())
		assert.Equal(t, StatusUninitialized, plugin2.Status())
	})
}

func TestManagerStartAll(t *testing.T) {
	t.Run("start all enabled plugins", func(t *testing.T) {
		manager := NewManager()
		plugin1 := NewBasePlugin("plugin1", "1.0.0")
		plugin2 := NewBasePlugin("plugin2", "2.0.0")

		manager.Register(plugin1)
		manager.Register(plugin2)

		err := manager.StartAll()

		require.NoError(t, err)
		assert.Equal(t, StatusRunning, plugin1.Status())
		assert.Equal(t, StatusRunning, plugin2.Status())
	})

	t.Run("skip disabled plugins", func(t *testing.T) {
		manager := NewManager()
		plugin1 := NewBasePlugin("plugin1", "1.0.0")
		plugin2 := NewBasePlugin("plugin2", "2.0.0")

		manager.Register(plugin1)
		manager.Register(plugin2)
		manager.Disable("plugin2")

		err := manager.StartAll()

		require.NoError(t, err)
		assert.Equal(t, StatusRunning, plugin1.Status())
		assert.Equal(t, StatusUninitialized, plugin2.Status())
	})
}

func TestManagerStopAll(t *testing.T) {
	t.Run("stop all running plugins", func(t *testing.T) {
		manager := NewManager()
		plugin1 := NewBasePlugin("plugin1", "1.0.0")
		plugin2 := NewBasePlugin("plugin2", "2.0.0")

		manager.Register(plugin1)
		manager.Register(plugin2)
		manager.StartAll()

		err := manager.StopAll()

		require.NoError(t, err)
		assert.Equal(t, StatusStopped, plugin1.Status())
		assert.Equal(t, StatusStopped, plugin2.Status())
	})
}

// Mock implementations for testing plugin types

type mockMiddlewarePlugin struct {
	*BasePlugin
}

func (m *mockMiddlewarePlugin) Handler() func(ctx *fasthttp.RequestCtx, next func()) {
	return func(ctx *fasthttp.RequestCtx, next func()) {
		next()
	}
}

type mockServicePlugin struct {
	*BasePlugin
	running bool
	mu      sync.Mutex
}

func (m *mockServicePlugin) Run() error {
	m.mu.Lock()
	m.running = true
	m.mu.Unlock()

	// Simulate background service
	time.Sleep(10 * time.Millisecond)
	return nil
}

type mockHookPlugin struct {
	*BasePlugin
	onRequestCalled  bool
	onResponseCalled bool
	onErrorCalled    bool
	mu               sync.Mutex
}

func (m *mockHookPlugin) OnRequest(ctx *fasthttp.RequestCtx) {
	m.mu.Lock()
	m.onRequestCalled = true
	m.mu.Unlock()
}

func (m *mockHookPlugin) OnResponse(ctx *fasthttp.RequestCtx) {
	m.mu.Lock()
	m.onResponseCalled = true
	m.mu.Unlock()
}

func (m *mockHookPlugin) OnError(ctx *fasthttp.RequestCtx, err error) {
	m.mu.Lock()
	m.onErrorCalled = true
	m.mu.Unlock()
}

func TestManagerPluginTypes(t *testing.T) {
	t.Run("register middleware plugin", func(t *testing.T) {
		manager := NewManager()
		plugin := &mockMiddlewarePlugin{
			BasePlugin: NewBasePlugin("middleware", "1.0.0"),
		}

		err := manager.Register(plugin)

		require.NoError(t, err)

		middleware := manager.GetMiddleware()
		assert.Len(t, middleware, 1)
		assert.Equal(t, "middleware", middleware[0].Name())
	})

	t.Run("register service plugin", func(t *testing.T) {
		manager := NewManager()
		plugin := &mockServicePlugin{
			BasePlugin: NewBasePlugin("service", "1.0.0"),
		}

		err := manager.Register(plugin)

		require.NoError(t, err)

		services := manager.GetServices()
		assert.Len(t, services, 1)
		assert.Equal(t, "service", services[0].Name())
	})

	t.Run("register hook plugin", func(t *testing.T) {
		manager := NewManager()
		plugin := &mockHookPlugin{
			BasePlugin: NewBasePlugin("hook", "1.0.0"),
		}

		err := manager.Register(plugin)

		require.NoError(t, err)

		hooks := manager.GetHooks()
		assert.Len(t, hooks, 1)
		assert.Equal(t, "hook", hooks[0].Name())
	})
}

func TestManagerHookCalls(t *testing.T) {
	t.Run("call request hooks", func(t *testing.T) {
		manager := NewManager()
		plugin := &mockHookPlugin{
			BasePlugin: NewBasePlugin("hook", "1.0.0"),
		}

		manager.Register(plugin)

		ctx := &fasthttp.RequestCtx{}
		manager.CallRequestHooks(ctx)

		assert.True(t, plugin.onRequestCalled)
	})

	t.Run("call response hooks", func(t *testing.T) {
		manager := NewManager()
		plugin := &mockHookPlugin{
			BasePlugin: NewBasePlugin("hook", "1.0.0"),
		}

		manager.Register(plugin)

		ctx := &fasthttp.RequestCtx{}
		manager.CallResponseHooks(ctx)

		assert.True(t, plugin.onResponseCalled)
	})

	t.Run("call error hooks", func(t *testing.T) {
		manager := NewManager()
		plugin := &mockHookPlugin{
			BasePlugin: NewBasePlugin("hook", "1.0.0"),
		}

		manager.Register(plugin)

		ctx := &fasthttp.RequestCtx{}
		manager.CallErrorHooks(ctx, errors.New("test error"))

		assert.True(t, plugin.onErrorCalled)
	})

	t.Run("skip disabled hook plugins", func(t *testing.T) {
		manager := NewManager()
		plugin := &mockHookPlugin{
			BasePlugin: NewBasePlugin("hook", "1.0.0"),
		}

		manager.Register(plugin)
		manager.Disable("hook")

		ctx := &fasthttp.RequestCtx{}
		manager.CallRequestHooks(ctx)

		assert.False(t, plugin.onRequestCalled)
	})
}

func TestManagerUnregisterPluginTypes(t *testing.T) {
	t.Run("unregister middleware plugin", func(t *testing.T) {
		manager := NewManager()
		plugin := &mockMiddlewarePlugin{
			BasePlugin: NewBasePlugin("middleware", "1.0.0"),
		}

		manager.Register(plugin)
		assert.Len(t, manager.GetMiddleware(), 1)

		manager.Unregister("middleware")
		assert.Empty(t, manager.GetMiddleware())
	})

	t.Run("unregister service plugin", func(t *testing.T) {
		manager := NewManager()
		plugin := &mockServicePlugin{
			BasePlugin: NewBasePlugin("service", "1.0.0"),
		}

		manager.Register(plugin)
		assert.Len(t, manager.GetServices(), 1)

		manager.Unregister("service")
		assert.Empty(t, manager.GetServices())
	})

	t.Run("unregister hook plugin", func(t *testing.T) {
		manager := NewManager()
		plugin := &mockHookPlugin{
			BasePlugin: NewBasePlugin("hook", "1.0.0"),
		}

		manager.Register(plugin)
		assert.Len(t, manager.GetHooks(), 1)

		manager.Unregister("hook")
		assert.Empty(t, manager.GetHooks())
	})
}

func TestConcurrentManagerAccess(t *testing.T) {
	t.Run("concurrent register and unregister", func(t *testing.T) {
		manager := NewManager()
		var wg sync.WaitGroup

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				name := string(rune('a' + id))
				plugin := NewBasePlugin(name, "1.0.0")

				manager.Register(plugin)
				time.Sleep(1 * time.Millisecond)
				manager.Unregister(name)
			}(i)
		}

		wg.Wait()

		// Should not panic
		list := manager.List()
		assert.GreaterOrEqual(t, len(list), 0)
	})
}

func TestPluginStatus(t *testing.T) {
	t.Run("plugin status constants", func(t *testing.T) {
		assert.Equal(t, PluginStatus("uninitialized"), StatusUninitialized)
		assert.Equal(t, PluginStatus("initialized"), StatusInitialized)
		assert.Equal(t, PluginStatus("running"), StatusRunning)
		assert.Equal(t, PluginStatus("stopped"), StatusStopped)
		assert.Equal(t, PluginStatus("error"), StatusError)
	})
}

func BenchmarkManager(b *testing.B) {
	b.Run("Register", func(b *testing.B) {
		manager := NewManager()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			plugin := NewBasePlugin("test", "1.0.0")
			manager.Register(plugin)
		}
	})

	b.Run("Get", func(b *testing.B) {
		manager := NewManager()
		plugin := NewBasePlugin("test", "1.0.0")
		manager.Register(plugin)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = manager.Get("test")
		}
	})

	b.Run("List", func(b *testing.B) {
		manager := NewManager()
		for i := 0; i < 10; i++ {
			plugin := NewBasePlugin(string(rune('a'+i)), "1.0.0")
			manager.Register(plugin)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = manager.List()
		}
	})
}
