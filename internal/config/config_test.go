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
		{"annotations exclude only", FilterConfig{Annotations: AnnotationConfig{Exclude: []string{"HasAnyRole"}}}, false},
		{"annotations include only", FilterConfig{Annotations: AnnotationConfig{Include: []string{"Public"}}}, false},
		{"include and annotations", FilterConfig{Include: []string{"a.*"}, Annotations: AnnotationConfig{Exclude: []string{"Internal"}}}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.cfg.IsPassThrough() != tc.expected {
				t.Errorf("IsPassThrough: expected %v", tc.expected)
			}
		})
	}
}

func TestLoadConfigWithAnnotations(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "filter.yaml")

	content := `annotations:
  - "HasAnyRole"
  - "Internal"
`
	os.WriteFile(cfgPath, []byte(content), 0o644)

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if len(cfg.Annotations.Exclude) != 2 {
		t.Errorf("expected 2 annotations in exclude, got %d", len(cfg.Annotations.Exclude))
	}
	if cfg.Annotations.Exclude[0] != "HasAnyRole" {
		t.Errorf("annotations.exclude[0]: expected HasAnyRole, got %s", cfg.Annotations.Exclude[0])
	}
	if !cfg.HasAnnotations() {
		t.Error("HasAnnotations should return true")
	}
	if cfg.IsPassThrough() {
		t.Error("config with annotations should not be pass-through")
	}
}

func TestLoadConfigAnnotationsWithIncludeExclude(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "filter.yaml")

	content := `include:
  - "my.package.*"
annotations:
  - "HasAnyRole"
`
	os.WriteFile(cfgPath, []byte(content), 0o644)

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if len(cfg.Include) != 1 {
		t.Errorf("expected 1 include, got %d", len(cfg.Include))
	}
	if len(cfg.Annotations.Exclude) != 1 {
		t.Errorf("expected 1 annotation in exclude, got %d", len(cfg.Annotations.Exclude))
	}
}

// T018: Test structured exclude config format
func TestLoadConfigStructuredExclude(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "filter.yaml")

	content := "annotations:\n  exclude:\n    - \"HasAnyRole\"\n"
	os.WriteFile(cfgPath, []byte(content), 0o644)

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if len(cfg.Annotations.Exclude) != 1 || cfg.Annotations.Exclude[0] != "HasAnyRole" {
		t.Errorf("expected Exclude=[HasAnyRole], got %v", cfg.Annotations.Exclude)
	}
	if len(cfg.Annotations.Include) != 0 {
		t.Errorf("expected empty Include, got %v", cfg.Annotations.Include)
	}
	if !cfg.HasAnnotationExclude() {
		t.Error("HasAnnotationExclude should return true")
	}
	if cfg.HasAnnotationInclude() {
		t.Error("HasAnnotationInclude should return false")
	}
}

// T019: Test structured include config format
func TestLoadConfigStructuredInclude(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "filter.yaml")

	content := "annotations:\n  include:\n    - \"Public\"\n"
	os.WriteFile(cfgPath, []byte(content), 0o644)

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if len(cfg.Annotations.Include) != 1 || cfg.Annotations.Include[0] != "Public" {
		t.Errorf("expected Include=[Public], got %v", cfg.Annotations.Include)
	}
	if len(cfg.Annotations.Exclude) != 0 {
		t.Errorf("expected empty Exclude, got %v", cfg.Annotations.Exclude)
	}
	if !cfg.HasAnnotationInclude() {
		t.Error("HasAnnotationInclude should return true")
	}
	if cfg.HasAnnotationExclude() {
		t.Error("HasAnnotationExclude should return false")
	}
}

// T020: Test flat annotations backward compatibility
func TestLoadConfigFlatAnnotationsBackwardCompat(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "filter.yaml")

	content := "annotations:\n  - \"HasAnyRole\"\n"
	os.WriteFile(cfgPath, []byte(content), 0o644)

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if len(cfg.Annotations.Exclude) != 1 || cfg.Annotations.Exclude[0] != "HasAnyRole" {
		t.Errorf("expected Exclude=[HasAnyRole] from flat format, got %v", cfg.Annotations.Exclude)
	}
	if len(cfg.Annotations.Include) != 0 {
		t.Errorf("expected empty Include from flat format, got %v", cfg.Annotations.Include)
	}
}

// T022 (updated by 011): Combined include+exclude is now allowed
func TestValidateCombinedIncludeExcludePass(t *testing.T) {
	cfg := FilterConfig{
		Annotations: AnnotationConfig{
			Include: []string{"Public"},
			Exclude: []string{"Internal"},
		},
	}
	err := cfg.Validate()
	if err != nil {
		t.Fatalf("expected no error for combined include+exclude, got: %v", err)
	}
}

// T023: Test validation passes with empty exclude list
func TestValidateEmptyListsPass(t *testing.T) {
	cfg := FilterConfig{
		Annotations: AnnotationConfig{
			Include: []string{"Public"},
			Exclude: []string{},
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no error with empty exclude, got: %v", err)
	}
}

// T024: Test validation passes with no annotations
func TestValidateNoAnnotationsPass(t *testing.T) {
	cfg := FilterConfig{}
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no error with no annotations, got: %v", err)
	}
}

// T023 (008): Test config loading with substitutions
func TestLoadConfigWithSubstitutions(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "filter.yaml")

	content := `substitutions:
  HasAnyRole: "Auth"
  Internal: ""
strict_substitutions: true
`
	os.WriteFile(cfgPath, []byte(content), 0o644)

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if len(cfg.Substitutions) != 2 {
		t.Errorf("expected 2 substitutions, got %d", len(cfg.Substitutions))
	}
	if cfg.Substitutions["HasAnyRole"] != "Auth" {
		t.Errorf("HasAnyRole: expected 'Auth', got %q", cfg.Substitutions["HasAnyRole"])
	}
	if cfg.Substitutions["Internal"] != "" {
		t.Errorf("Internal: expected empty string, got %q", cfg.Substitutions["Internal"])
	}
	if !cfg.StrictSubstitutions {
		t.Error("StrictSubstitutions should be true")
	}
	if !cfg.HasSubstitutions() {
		t.Error("HasSubstitutions should return true")
	}
}

// T024 (008): Test config loading without substitutions
func TestLoadConfigNoSubstitutions(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "filter.yaml")

	content := `include:
  - "my.package.*"
`
	os.WriteFile(cfgPath, []byte(content), 0o644)

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Substitutions != nil && len(cfg.Substitutions) > 0 {
		t.Errorf("expected nil/empty substitutions, got %v", cfg.Substitutions)
	}
	if cfg.StrictSubstitutions {
		t.Error("StrictSubstitutions should be false")
	}
	if cfg.HasSubstitutions() {
		t.Error("HasSubstitutions should return false")
	}
}

// Test IsPassThrough is not affected by substitutions
func TestIsPassThroughNotAffectedBySubstitutions(t *testing.T) {
	cfg := FilterConfig{
		Substitutions: map[string]string{"HasAnyRole": "Auth"},
	}
	if !cfg.IsPassThrough() {
		t.Error("substitution-only config should be pass-through (writes all files)")
	}
}
