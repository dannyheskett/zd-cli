package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

const (
	configDirName  = ".zd"
	configFileName = "config"
)

// GetConfigPath returns the full path to the config file
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(home, configDirName, configFileName), nil
}

// GetConfigDir returns the full path to the config directory
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(home, configDirName), nil
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	// Create directory with secure permissions (0700 = rwx------)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return nil
}

// Load reads the configuration from the config file
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, ErrConfigNotFound
	}

	// Load INI file
	iniFile, err := ini.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	config := NewConfig()

	// Read core section
	coreSection := iniFile.Section("core")
	if coreSection != nil {
		config.Current = coreSection.Key("current").String()
	}

	// Read instance sections
	for _, section := range iniFile.Sections() {
		// Skip default and core sections
		if section.Name() == ini.DefaultSection || section.Name() == "core" {
			continue
		}

		// Parse instance sections (format: instance "name")
		if strings.HasPrefix(section.Name(), "instance \"") && strings.HasSuffix(section.Name(), "\"") {
			instanceName := strings.TrimPrefix(section.Name(), "instance \"")
			instanceName = strings.TrimSuffix(instanceName, "\"")

			instance := &Instance{
				Name: instanceName,
			}

			if err := section.MapTo(instance); err != nil {
				return nil, fmt.Errorf("failed to parse instance %s: %w", instanceName, err)
			}

			config.Instances[instanceName] = instance
		}
	}

	return config, nil
}

// Save writes the configuration to the config file
func Save(config *Config) error {
	// Ensure config directory exists
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create new INI file
	iniFile := ini.Empty()

	// Write core section
	coreSection, err := iniFile.NewSection("core")
	if err != nil {
		return fmt.Errorf("failed to create core section: %w", err)
	}

	if _, err := coreSection.NewKey("current", config.Current); err != nil {
		return fmt.Errorf("failed to write current instance: %w", err)
	}

	// Write instance sections
	for name, instance := range config.Instances {
		sectionName := fmt.Sprintf("instance \"%s\"", name)
		section, err := iniFile.NewSection(sectionName)
		if err != nil {
			return fmt.Errorf("failed to create section for instance %s: %w", name, err)
		}

		if err := section.ReflectFrom(instance); err != nil {
			return fmt.Errorf("failed to write instance %s: %w", name, err)
		}
	}

	// Save to file with secure permissions (0600 = rw-------)
	if err := iniFile.SaveTo(configPath); err != nil {
		return fmt.Errorf("failed to save config file: %w", err)
	}

	// Ensure file has correct permissions
	if err := os.Chmod(configPath, 0600); err != nil {
		return fmt.Errorf("failed to set config file permissions: %w", err)
	}

	return nil
}

// LoadOrCreate loads the config file or creates a new one if it doesn't exist
func LoadOrCreate() (*Config, error) {
	config, err := Load()
	if err == ErrConfigNotFound {
		return NewConfig(), nil
	}
	return config, err
}
