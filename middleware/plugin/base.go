package plugin

import (
	"sync"

	"github.com/lazygophers/log"
)

// BasePlugin provides a base implementation of Plugin interface
type BasePlugin struct {
	name    string
	version string
	status  PluginStatus
	config  interface{}
	mu      sync.RWMutex
}

// NewBasePlugin creates a new base plugin
func NewBasePlugin(name, version string) *BasePlugin {
	return &BasePlugin{
		name:    name,
		version: version,
		status:  StatusUninitialized,
	}
}

// Name returns the plugin name
func (p *BasePlugin) Name() string {
	return p.name
}

// Version returns the plugin version
func (p *BasePlugin) Version() string {
	return p.version
}

// Init initializes the plugin
func (p *BasePlugin) Init(config interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.config = config
	p.status = StatusInitialized

	log.Infof("Plugin %s@%s initialized", p.name, p.version)
	return nil
}

// Start starts the plugin
func (p *BasePlugin) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.status = StatusRunning

	log.Infof("Plugin %s@%s started", p.name, p.version)
	return nil
}

// Stop stops the plugin
func (p *BasePlugin) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.status = StatusStopped

	log.Infof("Plugin %s@%s stopped", p.name, p.version)
	return nil
}

// Status returns the plugin status
func (p *BasePlugin) Status() PluginStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status
}

// Config returns the plugin configuration
func (p *BasePlugin) Config() interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.config
}

// SetStatus sets the plugin status
func (p *BasePlugin) SetStatus(status PluginStatus) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.status = status
}

// GetConfigAs retrieves config as a specific type
func (p *BasePlugin) GetConfigAs(out interface{}) error {
	p.mu.RLock()
	config := p.config
	p.mu.RUnlock()

	if config == nil {
		return nil
	}

	// Type assertion - in real usage, might want to use reflection or encoding/json
	// For now, this is a simple implementation
	switch v := config.(type) {
	case map[string]interface{}:
		// Config is already a map
		if m, ok := out.(*map[string]interface{}); ok {
			*m = v
			return nil
		}
	}

	return nil
}

// PluginMetadata represents plugin metadata
type PluginMetadata struct {
	Name         string
	Version      string
	Description  string
	Author       string
	License      string
	Homepage     string
	Dependencies []string
	Tags         []string
}

// MetadataPlugin extends Plugin with metadata
type MetadataPlugin interface {
	Plugin
	Metadata() PluginMetadata
}
