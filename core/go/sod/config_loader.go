package sod

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// fileConfigLoader implements ConfigLoader for YAML files
type fileConfigLoader struct {
	configPath string
	config     *Config
	mu         sync.RWMutex
	watchers   []func(*Config)
}

// NewFileConfigLoader creates a config loader from a YAML file
func NewFileConfigLoader(configPath string) ConfigLoader {
	return &fileConfigLoader{
		configPath: configPath,
		watchers:   make([]func(*Config), 0),
	}
}

// Load reads SOD configuration from YAML file
func (l *fileConfigLoader) Load() (*Config, error) {
	data, err := os.ReadFile(l.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Validate configuration
	if err := l.validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid SOD config: %w", err)
	}

	l.mu.Lock()
	l.config = &config
	l.mu.Unlock()

	return &config, nil
}

// Reload refreshes the configuration
func (l *fileConfigLoader) Reload() error {
	config, err := l.Load()
	if err != nil {
		return err
	}

	// Notify watchers
	l.mu.RLock()
	watchers := l.watchers
	l.mu.RUnlock()

	for _, watcher := range watchers {
		watcher(config)
	}

	return nil
}

// Watch monitors configuration changes
func (l *fileConfigLoader) Watch(ctx context.Context, callback func(*Config)) error {
	l.mu.Lock()
	l.watchers = append(l.watchers, callback)
	l.mu.Unlock()

	// Simple polling-based watch (could use fsnotify for production)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	var lastModTime time.Time
	if stat, err := os.Stat(l.configPath); err == nil {
		lastModTime = stat.ModTime()
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			stat, err := os.Stat(l.configPath)
			if err != nil {
				continue
			}

			if stat.ModTime().After(lastModTime) {
				lastModTime = stat.ModTime()
				if err := l.Reload(); err != nil {
					// Log error but continue watching
					continue
				}
			}
		}
	}
}

// validateConfig performs basic validation
func (l *fileConfigLoader) validateConfig(config *Config) error {
	if config.ServiceName == "" {
		return fmt.Errorf("service_name is required")
	}

	if config.Environment == "" {
		return fmt.Errorf("environment is required")
	}

	for code, errCfg := range config.Errors {
		if errCfg.BaseSeverity < 1 || errCfg.BaseSeverity > 10 {
			return fmt.Errorf("error %s: base_severity must be 1-10", code)
		}
		if errCfg.BaseOccurrence < 1 || errCfg.BaseOccurrence > 10 {
			return fmt.Errorf("error %s: base_occurrence must be 1-10", code)
		}
		if errCfg.BaseDetect < 1 || errCfg.BaseDetect > 10 {
			return fmt.Errorf("error %s: base_detect must be 1-10", code)
		}
	}

	return nil
}

// envConfigLoader implements ConfigLoader from environment variables (for cloud-native)
type envConfigLoader struct {
	serviceName string
	environment string
}

// NewEnvConfigLoader creates a config loader from environment variables
func NewEnvConfigLoader(serviceName, environment string) ConfigLoader {
	return &envConfigLoader{
		serviceName: serviceName,
		environment: environment,
	}
}

// Load reads configuration from environment variables
func (l *envConfigLoader) Load() (*Config, error) {
	// This is a placeholder for cloud-native configuration
	// In production, this would read from env vars, consul, etcd, etc.

	configPath := os.Getenv("SOD_CONFIG_PATH")
	if configPath == "" {
		// Default to service-specific config
		configPath = filepath.Join("config", fmt.Sprintf("sod_%s.yaml", l.serviceName))
	}

	loader := NewFileConfigLoader(configPath)
	return loader.Load()
}

// Reload refreshes the configuration
func (l *envConfigLoader) Reload() error {
	_, err := l.Load()
	return err
}

// Watch is not implemented for env loader (use file loader for watching)
func (l *envConfigLoader) Watch(ctx context.Context, callback func(*Config)) error {
	return fmt.Errorf("watch not supported for environment config loader")
}
