package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// AppConfig maps the yaml structure
type AppConfig struct {
	Server struct {
		Port    int    `mapstructure:"port"`
		Message string `mapstructure:"message"`
	} `mapstructure:"server"`
	Features struct {
		Beta bool `mapstructure:"beta"`
	} `mapstructure:"features"`
}

// ConfigManager holds a thread-safe configuration
type ConfigManager struct {
	mu     sync.RWMutex
	config AppConfig
}

func NewConfigManager() *ConfigManager {
	return &ConfigManager{}
}

func (cm *ConfigManager) Update(newConfig AppConfig) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.config = newConfig
	log.Println("Configuration updated successfully")
}

func (cm *ConfigManager) Get() AppConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config
}

func main() {
	cm := NewConfigManager()

	// Viper setup
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config: %s", err)
	}

	// Initial load
	var initialConfig AppConfig
	if err := v.Unmarshal(&initialConfig); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}
	cm.Update(initialConfig)

	// Watch for changes
	v.OnConfigChange(func(e fsnotify.Event) {
		log.Printf("Config file changed: %s", e.Name)
		var newConfig AppConfig
		if err := v.Unmarshal(&newConfig); err != nil {
			log.Printf("Error unmarshaling new config: %v", err)
			return
		}
		cm.Update(newConfig)
	})
	v.WatchConfig()

	// HTTP Handler
	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		cfg := cm.Get()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cfg)
	})

	port := cm.Get().Server.Port // Note: If port changes, restart is usually needed for net.Listen. 
	// Hot reload is mostly useful for runtime flags/messages, not startup bindings.
	
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Server listening on %s (Try editing config.yaml)", addr)
	http.ListenAndServe(addr, nil)
}
