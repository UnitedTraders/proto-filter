package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// AnnotationConfig holds include/exclude annotation filter lists.
// Supports both the old flat format (annotations: [list]) and the new
// structured format (annotations: {include: [...], exclude: [...]}).
type AnnotationConfig struct {
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
}

// UnmarshalYAML implements custom YAML unmarshaling for AnnotationConfig.
// If the YAML node is a sequence (old flat format), the items are treated
// as an exclude list. If it is a mapping (new structured format), the
// include and exclude sub-keys are parsed normally.
func (a *AnnotationConfig) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.SequenceNode:
		// Old flat format: annotations: ["HasAnyRole", "Internal"]
		var items []string
		if err := value.Decode(&items); err != nil {
			return err
		}
		a.Exclude = items
		return nil
	case yaml.MappingNode:
		// New structured format: annotations: {include: [...], exclude: [...]}
		type plain AnnotationConfig
		var p plain
		if err := value.Decode(&p); err != nil {
			return err
		}
		*a = AnnotationConfig(p)
		return nil
	default:
		return fmt.Errorf("annotations must be a list or a mapping, got %v", value.Kind)
	}
}

// FilterConfig holds include/exclude glob patterns and annotation filters.
type FilterConfig struct {
	Include     []string         `yaml:"include"`
	Exclude     []string         `yaml:"exclude"`
	Annotations AnnotationConfig `yaml:"annotations"`
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

// Validate checks the configuration for invalid combinations.
// Returns an error if both annotations.include and annotations.exclude
// are non-empty.
func (c *FilterConfig) Validate() error {
	if len(c.Annotations.Include) > 0 && len(c.Annotations.Exclude) > 0 {
		return fmt.Errorf("annotations.include and annotations.exclude are mutually exclusive")
	}
	return nil
}

// IsPassThrough returns true if no filter rules are defined.
func (c *FilterConfig) IsPassThrough() bool {
	return len(c.Include) == 0 && len(c.Exclude) == 0 &&
		len(c.Annotations.Include) == 0 && len(c.Annotations.Exclude) == 0
}

// HasAnnotations returns true if annotation-based filtering is configured.
func (c *FilterConfig) HasAnnotations() bool {
	return len(c.Annotations.Include) > 0 || len(c.Annotations.Exclude) > 0
}

// HasAnnotationInclude returns true if annotation include mode is configured.
func (c *FilterConfig) HasAnnotationInclude() bool {
	return len(c.Annotations.Include) > 0
}

// HasAnnotationExclude returns true if annotation exclude mode is configured.
func (c *FilterConfig) HasAnnotationExclude() bool {
	return len(c.Annotations.Exclude) > 0
}
