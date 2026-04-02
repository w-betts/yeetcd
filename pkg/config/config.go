package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// YeetcdConfig represents yeetcd.yaml configuration
type YeetcdConfig struct {
	Name       string              `yaml:"name"`
	Language   SourceLanguage      `yaml:"language"`
	BuildImage string              `yaml:"buildImage"`
	BuildCmd   string              `yaml:"buildCmd,omitempty"`
	Artifacts  []ArtifactDefinition `yaml:"artifacts,omitempty"`
}

// Load parses a yeetcd.yaml file
func Load(path string) (*YeetcdConfig, error) {
	if path == "" {
		return nil, errors.New("config path is empty")
	}
	
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}
	
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Parse YAML
	var config YeetcdConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	// Validate required fields
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	
	return &config, nil
}

// Validate checks if the configuration is valid
func (c *YeetcdConfig) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	
	if c.Language == "" {
		return errors.New("language is required")
	}
	
	if c.BuildImage == "" {
		return errors.New("buildImage is required")
	}
	
	return nil
}

// LoadFromBytes parses YAML data directly into YeetcdConfig
func LoadFromBytes(data []byte) (*YeetcdConfig, error) {
	var config YeetcdConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	
	return &config, nil
}
