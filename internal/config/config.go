package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// FilterConfig holds include/exclude glob patterns and annotation filters.
type FilterConfig struct {
	Include     []string `yaml:"include"`
	Exclude     []string `yaml:"exclude"`
	Annotations []string `yaml:"annotations"`
}

// LoadConfig reads and parses a YAML filter configuration file.
func LoadConfig(path string) (*FilterConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg FilterConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config YAML: %w", err)
	}

	return &cfg, nil
}

// IsPassThrough returns true if no filter rules are defined.
func (c *FilterConfig) IsPassThrough() bool {
	return len(c.Include) == 0 && len(c.Exclude) == 0 && len(c.Annotations) == 0
}

// HasAnnotations returns true if annotation-based filtering is configured.
func (c *FilterConfig) HasAnnotations() bool {
	return len(c.Annotations) > 0
}
