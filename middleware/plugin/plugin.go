package plugin

import (
	"errors"
	"sync"

	"github.com/lazygophers/log"
	"github.com/valyala/fasthttp"
)

var (
	ErrPluginNotFound      = errors.New("plugin not found")
	ErrPluginAlreadyExists = errors.New("plugin already exists")
	ErrPluginDisabled      = errors.New("plugin is disabled")
)

// Plugin represents a plugin interface
type Plugin interface {
	// Name returns the plugin name
	Name() string

	// Version returns the plugin version
	Version() string

	// Init initializes the plugin
	Init(config interface{}) error

	// Start starts the plugin
	Start() error

	// Stop stops the plugin
	Stop() error

	// Status returns the plugin status
	Status() PluginStatus

	// Config returns the plugin configuration
	Config() interface{}
}

// PluginStatus represents plugin status
type PluginStatus string

const (
	StatusUninitialized PluginStatus = "uninitialized"
	StatusInitialized   PluginStatus = "initialized"
	StatusRunning       PluginStatus = "running"
	StatusStopped       PluginStatus = "stopped"
	StatusError         PluginStatus = "error"
)

// MiddlewarePlugin represents a middleware plugin
type MiddlewarePlugin interface {
	Plugin

	// Handler returns the middleware handler function
	Handler() func(ctx *fasthttp.RequestCtx, next func())
}

// ServicePlugin represents a service plugin (background services)
type ServicePlugin interface {
	Plugin

	// Run runs the service in background
	Run() error
}

// HookPlugin represents a hook plugin (lifecycle hooks)
type HookPlugin interface {
	Plugin

	// OnRequest is called before request is handled
	OnRequest(ctx *fasthttp.RequestCtx)

	// OnResponse is called after response is sent
	OnResponse(ctx *fasthttp.RequestCtx)

	// OnError is called when an error occurs
	OnError(ctx *fasthttp.RequestCtx, err error)
}

// PluginInfo represents plugin metadata
type PluginInfo struct {
	Name         string
	Version      string
	Description  string
	Author       string
	License      string
	Homepage     string
	Dependencies []string
}

// Manager manages plugins
type Manager struct {
	plugins    map[string]Plugin
	mu         sync.RWMutex
	middleware []MiddlewarePlugin
	services   []ServicePlugin
	hooks      []HookPlugin
	enabled    map[string]bool
}

// NewManager creates a new plugin manager
func NewManager() *Manager {
	return &Manager{
		plugins:    make(map[string]Plugin),
		middleware: make([]MiddlewarePlugin, 0),
		services:   make([]ServicePlugin, 0),
		hooks:      make([]HookPlugin, 0),
		enabled:    make(map[string]bool),
	}
}

// Register registers a plugin
func (m *Manager) Register(plugin Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := plugin.Name()
	if _, exists := m.plugins[name]; exists {
		return ErrPluginAlreadyExists
	}

	m.plugins[name] = plugin
	m.enabled[name] = true

	// Categorize plugin
	if mp, ok := plugin.(MiddlewarePlugin); ok {
		m.middleware = append(m.middleware, mp)
	}
	if sp, ok := plugin.(ServicePlugin); ok {
		m.services = append(m.services, sp)
	}
	if hp, ok := plugin.(HookPlugin); ok {
		m.hooks = append(m.hooks, hp)
	}

	log.Infof("Plugin registered: %s@%s", plugin.Name(), plugin.Version())
	return nil
}

// Unregister unregisters a plugin
func (m *Manager) Unregister(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[name]
	if !exists {
		return ErrPluginNotFound
	}

	// Stop plugin if running
	if plugin.Status() == StatusRunning {
		err := plugin.Stop()
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
	}

	// Remove from categories
	if _, ok := plugin.(MiddlewarePlugin); ok {
		m.middleware = removeMiddlewarePlugin(m.middleware, name)
	}
	if _, ok := plugin.(ServicePlugin); ok {
		m.services = removeServicePlugin(m.services, name)
	}
	if _, ok := plugin.(HookPlugin); ok {
		m.hooks = removeHookPlugin(m.hooks, name)
	}

	delete(m.plugins, name)
	delete(m.enabled, name)

	log.Infof("Plugin unregistered: %s", name)
	return nil
}

// Get gets a plugin by name
func (m *Manager) Get(name string) (Plugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, exists := m.plugins[name]
	if !exists {
		return nil, ErrPluginNotFound
	}

	return plugin, nil
}

// Enable enables a plugin
func (m *Manager) Enable(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[name]; !exists {
		return ErrPluginNotFound
	}

	m.enabled[name] = true
	log.Infof("Plugin enabled: %s", name)
	return nil
}

// Disable disables a plugin
func (m *Manager) Disable(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[name]; !exists {
		return ErrPluginNotFound
	}

	m.enabled[name] = false
	log.Infof("Plugin disabled: %s", name)
	return nil
}

// IsEnabled checks if a plugin is enabled
func (m *Manager) IsEnabled(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled[name]
}

// List lists all registered plugins
func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.plugins))
	for name := range m.plugins {
		names = append(names, name)
	}
	return names
}

// InitAll initializes all enabled plugins
func (m *Manager) InitAll(configs map[string]interface{}) error {
	m.mu.RLock()
	plugins := make([]Plugin, 0, len(m.plugins))
	for _, plugin := range m.plugins {
		if m.enabled[plugin.Name()] {
			plugins = append(plugins, plugin)
		}
	}
	m.mu.RUnlock()

	for _, plugin := range plugins {
		config := configs[plugin.Name()]
		err := plugin.Init(config)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
	}

	return nil
}

// StartAll starts all enabled plugins
func (m *Manager) StartAll() error {
	m.mu.RLock()
	plugins := make([]Plugin, 0, len(m.plugins))
	for _, plugin := range m.plugins {
		if m.enabled[plugin.Name()] {
			plugins = append(plugins, plugin)
		}
	}
	m.mu.RUnlock()

	for _, plugin := range plugins {
		err := plugin.Start()
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
	}

	// Start service plugins in background
	for _, service := range m.services {
		if m.IsEnabled(service.Name()) {
			go func(s ServicePlugin) {
				err := s.Run()
				if err != nil {
					log.Errorf("err:%v", err)
				}
			}(service)
		}
	}

	return nil
}

// StopAll stops all running plugins
func (m *Manager) StopAll() error {
	m.mu.RLock()
	plugins := make([]Plugin, 0, len(m.plugins))
	for _, plugin := range m.plugins {
		plugins = append(plugins, plugin)
	}
	m.mu.RUnlock()

	for _, plugin := range plugins {
		if plugin.Status() == StatusRunning {
			err := plugin.Stop()
			if err != nil {
				log.Errorf("err:%v", err)
				// Continue stopping other plugins
			}
		}
	}

	return nil
}

// GetMiddleware returns all enabled middleware plugins
func (m *Manager) GetMiddleware() []MiddlewarePlugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	enabled := make([]MiddlewarePlugin, 0)
	for _, plugin := range m.middleware {
		if m.enabled[plugin.Name()] {
			enabled = append(enabled, plugin)
		}
	}
	return enabled
}

// GetServices returns all enabled service plugins
func (m *Manager) GetServices() []ServicePlugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	enabled := make([]ServicePlugin, 0)
	for _, plugin := range m.services {
		if m.enabled[plugin.Name()] {
			enabled = append(enabled, plugin)
		}
	}
	return enabled
}

// GetHooks returns all enabled hook plugins
func (m *Manager) GetHooks() []HookPlugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	enabled := make([]HookPlugin, 0)
	for _, plugin := range m.hooks {
		if m.enabled[plugin.Name()] {
			enabled = append(enabled, plugin)
		}
	}
	return enabled
}

// CallRequestHooks calls OnRequest for all hook plugins
func (m *Manager) CallRequestHooks(ctx *fasthttp.RequestCtx) {
	hooks := m.GetHooks()
	for _, hook := range hooks {
		hook.OnRequest(ctx)
	}
}

// CallResponseHooks calls OnResponse for all hook plugins
func (m *Manager) CallResponseHooks(ctx *fasthttp.RequestCtx) {
	hooks := m.GetHooks()
	for _, hook := range hooks {
		hook.OnResponse(ctx)
	}
}

// CallErrorHooks calls OnError for all hook plugins
func (m *Manager) CallErrorHooks(ctx *fasthttp.RequestCtx, err error) {
	hooks := m.GetHooks()
	for _, hook := range hooks {
		hook.OnError(ctx, err)
	}
}

// Helper functions
func removeMiddlewarePlugin(plugins []MiddlewarePlugin, name string) []MiddlewarePlugin {
	for i, plugin := range plugins {
		if plugin.Name() == name {
			return append(plugins[:i], plugins[i+1:]...)
		}
	}
	return plugins
}

func removeServicePlugin(plugins []ServicePlugin, name string) []ServicePlugin {
	for i, plugin := range plugins {
		if plugin.Name() == name {
			return append(plugins[:i], plugins[i+1:]...)
		}
	}
	return plugins
}

func removeHookPlugin(plugins []HookPlugin, name string) []HookPlugin {
	for i, plugin := range plugins {
		if plugin.Name() == name {
			return append(plugins[:i], plugins[i+1:]...)
		}
	}
	return plugins
}
