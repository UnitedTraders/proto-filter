package config

import (
	"os"
	"path/filepath"
	"testing"
)

// T020: Test YAML config loading
func TestLoadConfig(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "filter.yaml")

	content := `include:
  - "my.package.OrderService"
  - "my.package.common.*"
exclude:
  - "my.package.internal.*"
`
	os.WriteFile(cfgPath, []byte(content), 0o644)

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if len(cfg.Include) != 2 {
		t.Errorf("expected 2 include patterns, got %d", len(cfg.Include))
	}
	if len(cfg.Exclude) != 1 {
		t.Errorf("expected 1 exclude pattern, got %d", len(cfg.Exclude))
	}
	if cfg.Include[0] != "my.package.OrderService" {
		t.Errorf("include[0]: expected my.package.OrderService, got %s", cfg.Include[0])
	}
}

func TestLoadConfigEmpty(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "empty.yaml")
	os.WriteFile(cfgPath, []byte(""), 0o644)

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if !cfg.IsPassThrough() {
		t.Error("empty config should be pass-through")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "bad.yaml")
	os.WriteFile(cfgPath, []byte("include: [not closed"), 0o644)

	_, err := LoadConfig(cfgPath)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestIsPassThrough(t *testing.T) {
	tests := []struct {
		name     string
		cfg      FilterConfig
		expected bool
	}{
		{"empty", FilterConfig{}, true},
		{"include only", FilterConfig{Include: []string{"a.*"}}, false},
		{"exclude only", FilterConfig{Exclude: []string{"b.*"}}, false},
		{"both", FilterConfig{Include: []string{"a.*"}, Exclude: []string{"b.*"}}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.cfg.IsPassThrough() != tc.expected {
				t.Errorf("IsPassThrough: expected %v", tc.expected)
			}
		})
	}
}
