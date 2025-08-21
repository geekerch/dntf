package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"

	"notification/internal/domain/shared"
	publicPlugins "notification/pkg/plugins"
)

// Plugin represents a loaded plugin instance
type Plugin interface {
	// GetInfo returns plugin information
	GetInfo() PluginInfo
	
	// GetChannelType returns the channel type definition
	GetChannelType() shared.ChannelTypeDefinition
	
	// Initialize initializes the plugin with configuration
	Initialize(config map[string]interface{}) error
	
	// Cleanup cleans up plugin resources
	Cleanup() error
}

// PluginInfo contains plugin metadata
type PluginInfo struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	LoadedAt    time.Time `json:"loadedAt"`
}

// PluginStatus represents the current status of a plugin
type PluginStatus struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"` // loaded, error, unloaded
	LoadedAt  time.Time `json:"loadedAt"`
	Error     string    `json:"error,omitempty"`
	Info      PluginInfo `json:"info"`
}

// PluginLoader manages loading and unloading of plugins
type PluginLoader interface {
	// LoadPlugin loads a plugin from file path
	LoadPlugin(pluginPath string) error
	
	// LoadPluginFromSource loads a plugin from source code
	LoadPluginFromSource(name, source string) error
	
	// UnloadPlugin unloads a plugin by name
	UnloadPlugin(pluginName string) error
	
	// GetPlugin gets a loaded plugin by name
	GetPlugin(pluginName string) (Plugin, error)
	
	// ListLoadedPlugins returns list of loaded plugin names
	ListLoadedPlugins() []string
	
	// GetPluginStatus gets the status of a plugin
	GetPluginStatus(pluginName string) (*PluginStatus, error)
	
	// GetAllPluginStatuses gets statuses of all plugins
	GetAllPluginStatuses() map[string]*PluginStatus
}

// YaegiPluginLoader implements PluginLoader using Yaegi interpreter
type YaegiPluginLoader struct {
	interpreter *interp.Interpreter
	plugins     map[string]*loadedPlugin
	statuses    map[string]*PluginStatus
	mutex       sync.RWMutex
	registry    shared.ChannelTypeRegistry
}

// loadedPlugin represents a loaded plugin with its context
type loadedPlugin struct {
	plugin   Plugin
	info     PluginInfo
	source   string
	loadedAt time.Time
}

// NewYaegiPluginLoader creates a new Yaegi-based plugin loader
func NewYaegiPluginLoader(registry shared.ChannelTypeRegistry) *YaegiPluginLoader {
	// Set up interpreter options with proper Go path
	options := interp.Options{
		GoPath: ".", // Set current directory as GOPATH
	}
	
	i := interp.New(options)
	
	// Use standard library
	i.Use(stdlib.Symbols)
	
	// Register our domain interfaces and types
	i.Use(map[string]map[string]reflect.Value{
		"notification/internal/domain/shared": {
			"ChannelTypeDefinition": reflect.ValueOf((*shared.ChannelTypeDefinition)(nil)),
			"GetChannelTypeRegistry": reflect.ValueOf(shared.GetChannelTypeRegistry),
		},
		"notification/pkg/plugins": {
			"PluginInfo":             reflect.ValueOf((*publicPlugins.PluginInfo)(nil)),
			"Plugin":                 reflect.ValueOf((*publicPlugins.Plugin)(nil)),
			"ChannelTypeDefinition":  reflect.ValueOf((*publicPlugins.ChannelTypeDefinition)(nil)),
		},
	})
	
	return &YaegiPluginLoader{
		interpreter: i,
		plugins:     make(map[string]*loadedPlugin),
		statuses:    make(map[string]*PluginStatus),
		registry:    registry,
	}
}

// LoadPlugin loads a plugin from file path
func (l *YaegiPluginLoader) LoadPlugin(pluginPath string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	// Read plugin source code
	source, err := os.ReadFile(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to read plugin file %s: %w", pluginPath, err)
	}
	
	// Extract plugin name from file path
	pluginName := filepath.Base(filepath.Dir(pluginPath))
	
	return l.loadPluginFromSourceInternal(pluginName, string(source))
}

// LoadPluginFromSource loads a plugin from source code
func (l *YaegiPluginLoader) LoadPluginFromSource(name, source string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	return l.loadPluginFromSourceInternal(name, source)
}

// loadPluginFromSourceInternal is the internal implementation for loading plugins
func (l *YaegiPluginLoader) loadPluginFromSourceInternal(name, source string) error {
	// Check if plugin is already loaded
	if _, exists := l.plugins[name]; exists {
		return fmt.Errorf("plugin %s is already loaded", name)
	}
	
	// Create a new interpreter context for this plugin
	pluginOptions := interp.Options{
		GoPath: ".", // Set current directory as GOPATH
	}
	pluginInterpreter := interp.New(pluginOptions)
	pluginInterpreter.Use(stdlib.Symbols)
	
	// Register our domain interfaces and public plugin API
	pluginInterpreter.Use(map[string]map[string]reflect.Value{
		"notification/internal/domain/shared": {
			"ChannelTypeDefinition": reflect.ValueOf((*shared.ChannelTypeDefinition)(nil)),
			"GetChannelTypeRegistry": reflect.ValueOf(shared.GetChannelTypeRegistry),
		},
		"notification/pkg/plugins": {
			"PluginInfo":             reflect.ValueOf((*publicPlugins.PluginInfo)(nil)),
			"Plugin":                 reflect.ValueOf((*publicPlugins.Plugin)(nil)),
			"ChannelTypeDefinition":  reflect.ValueOf((*publicPlugins.ChannelTypeDefinition)(nil)),
		},
	})
	
	// Evaluate the plugin source code
	_, err := pluginInterpreter.Eval(source)
	if err != nil {
		l.updatePluginStatus(name, "error", fmt.Sprintf("failed to evaluate plugin: %v", err), PluginInfo{})
		return fmt.Errorf("failed to evaluate plugin %s: %w", name, err)
	}
	
	// Get the NewPlugin function
	newPluginFunc, err := pluginInterpreter.Eval("NewPlugin")
	if err != nil {
		l.updatePluginStatus(name, "error", "plugin must export NewPlugin function", PluginInfo{})
		return fmt.Errorf("plugin %s must export NewPlugin function: %w", name, err)
	}
	
	// Call NewPlugin() to create plugin instance
	results := newPluginFunc.Call(nil)
	if len(results) == 0 {
		l.updatePluginStatus(name, "error", "NewPlugin function must return a plugin instance", PluginInfo{})
		return fmt.Errorf("NewPlugin function in plugin %s must return a plugin instance", name)
	}
	
	// Get the plugin value
	pluginValue := results[0]
	
	// Create plugin wrapper that can handle Yaegi values
	plugin := &yaegiPluginWrapper{
		interpreter: pluginInterpreter,
		value:       pluginValue,
		name:        name,
	}
	
	// Skip validation for now - Yaegi's valueInterface makes it difficult
	// We'll validate by actually trying to use the plugin methods
	fmt.Printf("⚠️ Skipping validation for Yaegi plugin %s\n", name)
	
	// Initialize the plugin
	if err := plugin.Initialize(nil); err != nil {
		l.updatePluginStatus(name, "error", fmt.Sprintf("failed to initialize plugin: %v", err), PluginInfo{})
		return fmt.Errorf("failed to initialize plugin %s: %w", name, err)
	}
	
	// Get plugin info
	info := plugin.GetInfo()
	if info.Name == "" {
		info.Name = name
	}
	info.LoadedAt = time.Now()
	
	// Register the channel type
	channelType := plugin.GetChannelType()
	if err := l.registry.RegisterChannelType(channelType); err != nil {
		l.updatePluginStatus(name, "error", fmt.Sprintf("failed to register channel type: %v", err), info)
		return fmt.Errorf("failed to register channel type for plugin %s: %w", name, err)
	}
	
	// Store the loaded plugin
	l.plugins[name] = &loadedPlugin{
		plugin:   plugin,
		info:     info,
		source:   source,
		loadedAt: time.Now(),
	}
	
	// Update status
	l.updatePluginStatus(name, "loaded", "", info)
	
	return nil
}

// UnloadPlugin unloads a plugin by name
func (l *YaegiPluginLoader) UnloadPlugin(pluginName string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	loadedPlugin, exists := l.plugins[pluginName]
	if !exists {
		return fmt.Errorf("plugin %s is not loaded", pluginName)
	}
	
	// Cleanup the plugin
	if err := loadedPlugin.plugin.Cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup plugin %s: %w", pluginName, err)
	}
	
	// Remove from loaded plugins
	delete(l.plugins, pluginName)
	
	// Update status
	l.updatePluginStatus(pluginName, "unloaded", "", loadedPlugin.info)
	
	return nil
}

// GetPlugin gets a loaded plugin by name
func (l *YaegiPluginLoader) GetPlugin(pluginName string) (Plugin, error) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	
	loadedPlugin, exists := l.plugins[pluginName]
	if !exists {
		return nil, fmt.Errorf("plugin %s is not loaded", pluginName)
	}
	
	return loadedPlugin.plugin, nil
}

// ListLoadedPlugins returns list of loaded plugin names
func (l *YaegiPluginLoader) ListLoadedPlugins() []string {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	
	names := make([]string, 0, len(l.plugins))
	for name := range l.plugins {
		names = append(names, name)
	}
	
	return names
}

// GetPluginStatus gets the status of a plugin
func (l *YaegiPluginLoader) GetPluginStatus(pluginName string) (*PluginStatus, error) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	
	status, exists := l.statuses[pluginName]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginName)
	}
	
	return status, nil
}

// GetAllPluginStatuses gets statuses of all plugins
func (l *YaegiPluginLoader) GetAllPluginStatuses() map[string]*PluginStatus {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	
	statuses := make(map[string]*PluginStatus)
	for name, status := range l.statuses {
		statuses[name] = status
	}
	
	return statuses
}

// updatePluginStatus updates the status of a plugin
func (l *YaegiPluginLoader) updatePluginStatus(name, status, errorMsg string, info PluginInfo) {
	l.statuses[name] = &PluginStatus{
		Name:     name,
		Status:   status,
		LoadedAt: time.Now(),
		Error:    errorMsg,
		Info:     info,
	}
}

// LoadPluginsFromDirectory loads all plugins from a directory
func (l *YaegiPluginLoader) LoadPluginsFromDirectory(pluginDir string) error {
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return fmt.Errorf("plugin directory %s does not exist", pluginDir)
	}
	
	return filepath.Walk(pluginDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Look for plugin.go files
		if info.Name() == "plugin.go" {
			if err := l.LoadPlugin(path); err != nil {
				// Log error but continue loading other plugins
				fmt.Printf("Failed to load plugin from %s: %v\n", path, err)
			}
		}
		
		return nil
	})
}

// yaegiPluginWrapper wraps a Yaegi plugin value and implements Plugin interface
type yaegiPluginWrapper struct {
	interpreter *interp.Interpreter
	value       reflect.Value
	name        string
}

func (ypw *yaegiPluginWrapper) validate() error {
	// Check if the value has the required methods by trying to call them
	requiredMethods := []string{"GetInfo", "GetChannelType", "Initialize", "Cleanup"}
	
	for _, methodName := range requiredMethods {
		if !ypw.hasMethod(methodName) {
			return fmt.Errorf("missing required method: %s", methodName)
		}
	}
	
	return nil
}

// hasMethod checks if a method exists by trying to call it via the interpreter
func (ypw *yaegiPluginWrapper) hasMethod(methodName string) bool {
	// Debug: print method checking
	fmt.Printf("🔍 Checking method: %s\n", methodName)
	
	// Try to get a fresh plugin instance and check if it has the method
	newPluginFunc, err := ypw.interpreter.Eval("NewPlugin")
	if err != nil {
		fmt.Printf("   ❌ Failed to get NewPlugin function: %v\n", err)
		return false
	}
	
	// Call NewPlugin() to get plugin instance
	results := newPluginFunc.Call(nil)
	if len(results) == 0 {
		fmt.Printf("   ❌ NewPlugin returned no results\n")
		return false
	}
	
	pluginInstance := results[0]
	fmt.Printf("   Plugin instance type: %v\n", pluginInstance.Type())
	
	// Check if method exists on this instance
	method := pluginInstance.MethodByName(methodName)
	if method.IsValid() {
		fmt.Printf("   ✅ Method %s found\n", methodName)
		return true
	}
	
	fmt.Printf("   ❌ Method %s not found\n", methodName)
	return false
}

// callMethod calls a method on the plugin via the interpreter
func (ypw *yaegiPluginWrapper) callMethod(methodName string, args ...interface{}) (interface{}, error) {
	// Use a different approach - execute the method call as an expression
	var expr string
	
	if len(args) == 0 {
		// Simple method call with no arguments
		expr = fmt.Sprintf("func() interface{} { p := NewPlugin(); return p.%s() }()", methodName)
	} else {
		// For methods with arguments, we need a more complex approach
		// For now, let's handle the Initialize case specifically
		if methodName == "Initialize" {
			expr = fmt.Sprintf("func() error { p := NewPlugin(); return p.Initialize(nil) }()")
		} else if methodName == "Cleanup" {
			expr = fmt.Sprintf("func() error { p := NewPlugin(); return p.Cleanup() }()")
		} else {
			expr = fmt.Sprintf("func() interface{} { p := NewPlugin(); return p.%s() }()", methodName)
		}
	}
	
	// Execute the expression
	result, err := ypw.interpreter.Eval(expr)
	if err != nil {
		return nil, fmt.Errorf("failed to execute method %s: %w", methodName, err)
	}
	
	return result.Interface(), nil
}

func (ypw *yaegiPluginWrapper) GetInfo() PluginInfo {
	result, err := ypw.callMethod("GetInfo")
	if err != nil {
		// Return default info if call fails
		return PluginInfo{
			Name:        ypw.name,
			Version:     "1.0.0",
			Description: "Plugin loaded via Yaegi",
			Author:      "Unknown",
			LoadedAt:    time.Now(),
		}
	}
	
	// Try to extract PluginInfo from the result
	if result != nil {
		resultValue := reflect.ValueOf(result)
		if resultValue.Kind() == reflect.Struct {
			info := PluginInfo{}
			
			// Extract fields using reflection
			if nameField := resultValue.FieldByName("Name"); nameField.IsValid() && nameField.Kind() == reflect.String {
				info.Name = nameField.String()
			}
			if versionField := resultValue.FieldByName("Version"); versionField.IsValid() && versionField.Kind() == reflect.String {
				info.Version = versionField.String()
			}
			if descField := resultValue.FieldByName("Description"); descField.IsValid() && descField.Kind() == reflect.String {
				info.Description = descField.String()
			}
			if authorField := resultValue.FieldByName("Author"); authorField.IsValid() && authorField.Kind() == reflect.String {
				info.Author = authorField.String()
			}
			if loadedAtField := resultValue.FieldByName("LoadedAt"); loadedAtField.IsValid() {
				if t, ok := loadedAtField.Interface().(time.Time); ok {
					info.LoadedAt = t
				}
			}
			
			return info
		}
	}
	
	// Return default info if extraction fails
	return PluginInfo{
		Name:        ypw.name,
		Version:     "1.0.0",
		Description: "Plugin loaded via Yaegi",
		Author:      "Unknown",
		LoadedAt:    time.Now(),
	}
}

func (ypw *yaegiPluginWrapper) GetChannelType() shared.ChannelTypeDefinition {
	result, err := ypw.callMethod("GetChannelType")
	if err != nil {
		return nil
	}
	
	if result != nil {
		// Return a wrapper for the channel type
		return &yaegiChannelTypeWrapper{
			interpreter: ypw.interpreter,
			value:       reflect.ValueOf(result),
		}
	}
	
	return nil
}

func (ypw *yaegiPluginWrapper) Initialize(config map[string]interface{}) error {
	result, err := ypw.callMethod("Initialize", config)
	if err != nil {
		return err
	}
	
	if result != nil {
		if err, ok := result.(error); ok {
			return err
		}
	}
	
	return nil
}

func (ypw *yaegiPluginWrapper) Cleanup() error {
	result, err := ypw.callMethod("Cleanup")
	if err != nil {
		return err
	}
	
	if result != nil {
		if err, ok := result.(error); ok {
			return err
		}
	}
	
	return nil
}

// yaegiChannelTypeWrapper wraps a Yaegi channel type value
type yaegiChannelTypeWrapper struct {
	interpreter *interp.Interpreter
	value       reflect.Value
}

// callChannelMethod calls a method on the channel type via the interpreter
func (yctw *yaegiChannelTypeWrapper) callChannelMethod(methodName string, args ...interface{}) (interface{}, error) {
	// Use the same expression-based approach as plugin methods
	var expr string
	
	if len(args) == 0 {
		// Simple method call with no arguments
		expr = fmt.Sprintf("func() interface{} { p := NewPlugin(); ct := p.GetChannelType(); return ct.%s() }()", methodName)
	} else {
		// For methods with arguments
		if methodName == "ValidateConfig" {
			expr = fmt.Sprintf("func() error { p := NewPlugin(); ct := p.GetChannelType(); return ct.ValidateConfig(nil) }()")
		} else if methodName == "CreateMessageSender" {
			expr = fmt.Sprintf("func() interface{} { p := NewPlugin(); ct := p.GetChannelType(); sender, _ := ct.CreateMessageSender(30000000000); return sender }()")
		} else {
			expr = fmt.Sprintf("func() interface{} { p := NewPlugin(); ct := p.GetChannelType(); return ct.%s() }()", methodName)
		}
	}
	
	// Execute the expression
	result, err := yctw.interpreter.Eval(expr)
	if err != nil {
		return nil, fmt.Errorf("failed to execute channel method %s: %w", methodName, err)
	}
	
	return result.Interface(), nil
}

func (yctw *yaegiChannelTypeWrapper) GetName() string {
	result, err := yctw.callChannelMethod("GetName")
	if err != nil {
		return "unknown"
	}
	
	if name, ok := result.(string); ok {
		return name
	}
	
	return "unknown"
}

func (yctw *yaegiChannelTypeWrapper) GetDisplayName() string {
	result, err := yctw.callChannelMethod("GetDisplayName")
	if err != nil {
		return "Unknown Channel"
	}
	
	if name, ok := result.(string); ok {
		return name
	}
	
	return "Unknown Channel"
}

func (yctw *yaegiChannelTypeWrapper) GetDescription() string {
	result, err := yctw.callChannelMethod("GetDescription")
	if err != nil {
		return "Channel type loaded via plugin"
	}
	
	if desc, ok := result.(string); ok {
		return desc
	}
	
	return "Channel type loaded via plugin"
}

func (yctw *yaegiChannelTypeWrapper) ValidateConfig(config map[string]interface{}) error {
	result, err := yctw.callChannelMethod("ValidateConfig", config)
	if err != nil {
		return err
	}
	
	if result != nil {
		if err, ok := result.(error); ok {
			return err
		}
	}
	
	return nil
}

func (yctw *yaegiChannelTypeWrapper) GetConfigSchema() map[string]interface{} {
	result, err := yctw.callChannelMethod("GetConfigSchema")
	if err != nil {
		return map[string]interface{}{}
	}
	
	if schema, ok := result.(map[string]interface{}); ok {
		return schema
	}
	
	return map[string]interface{}{}
}

func (yctw *yaegiChannelTypeWrapper) CreateMessageSender(timeout time.Duration) (interface{}, error) {
	result, err := yctw.callChannelMethod("CreateMessageSender", timeout)
	if err != nil {
		return nil, err
	}
	
	// For CreateMessageSender, we expect a tuple (sender, error)
	// But since we're using simplified method calling, we'll return the result as-is
	return result, nil
}

// Global plugin loader instance
var globalPluginLoader PluginLoader
var pluginLoaderOnce sync.Once

// GetPluginLoader returns the global plugin loader instance
func GetPluginLoader() PluginLoader {
	pluginLoaderOnce.Do(func() {
		globalPluginLoader = NewYaegiPluginLoader(shared.GetChannelTypeRegistry())
	})
	return globalPluginLoader
}

// SetPluginLoader sets the global plugin loader (for testing)
func SetPluginLoader(loader PluginLoader) {
	globalPluginLoader = loader
}