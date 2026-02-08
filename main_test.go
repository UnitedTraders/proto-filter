package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "proto-filter")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build: %v", err)
	}
	return bin
}

func runBinary(t *testing.T, bin string, args ...string) (stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	stderr = stderrBuf.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return stderr, exitErr.ExitCode()
		}
		t.Fatalf("run: %v", err)
	}
	return stderr, 0
}

func testdataDir(t *testing.T, sub string) string {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("testdata", sub))
	if err != nil {
		t.Fatalf("testdata path: %v", err)
	}
	return dir
}

// T035: Missing input directory
func TestMissingInputDirectory(t *testing.T) {
	bin := buildBinary(t)
	stderr, code := runBinary(t, bin,
		"--input", "/nonexistent/path/to/protos",
		"--output", t.TempDir(),
	)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("stderr should contain 'not found', got: %s", stderr)
	}
}

// T036: Same input/output path
func TestSameInputOutputPath(t *testing.T) {
	bin := buildBinary(t)
	dir := testdataDir(t, "simple")
	stderr, code := runBinary(t, bin,
		"--input", dir,
		"--output", dir,
	)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr, "must be different") {
		t.Errorf("stderr should mention 'must be different', got: %s", stderr)
	}
}

// T037: Invalid YAML config
func TestInvalidYAMLConfig(t *testing.T) {
	bin := buildBinary(t)
	tmp := t.TempDir()

	// Create invalid YAML
	badConfig := filepath.Join(tmp, "bad.yaml")
	os.WriteFile(badConfig, []byte("include: [not closed"), 0o644)

	stderr, code := runBinary(t, bin,
		"--input", testdataDir(t, "simple"),
		"--output", filepath.Join(tmp, "out"),
		"--config", badConfig,
	)
	if code != 2 {
		t.Errorf("expected exit code 2 for config error, got %d", code)
	}
	if !strings.Contains(stderr, "error") {
		t.Errorf("stderr should contain 'error', got: %s", stderr)
	}
}

// T038: Verbose output
func TestVerboseOutput(t *testing.T) {
	bin := buildBinary(t)
	outDir := t.TempDir()

	stderr, code := runBinary(t, bin,
		"--input", testdataDir(t, "simple"),
		"--output", outDir,
		"--verbose",
	)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stderr, "processed") {
		t.Errorf("verbose output should contain 'processed', got: %s", stderr)
	}
	if !strings.Contains(stderr, "wrote") {
		t.Errorf("verbose output should contain 'wrote', got: %s", stderr)
	}
}

// Test: missing required flags
func TestMissingRequiredFlags(t *testing.T) {
	bin := buildBinary(t)
	stderr, code := runBinary(t, bin)
	if code != 1 {
		t.Errorf("expected exit code 1 for missing flags, got %d", code)
	}
	if !strings.Contains(stderr, "--input and --output") {
		t.Errorf("stderr should mention required flags, got: %s", stderr)
	}
}

// Test: zero proto files warning
func TestZeroProtoFilesWarning(t *testing.T) {
	bin := buildBinary(t)
	emptyDir := t.TempDir()
	outDir := t.TempDir()

	stderr, code := runBinary(t, bin,
		"--input", emptyDir,
		"--output", outDir,
	)
	if code != 0 {
		t.Errorf("expected exit code 0 for zero files, got %d", code)
	}
	if !strings.Contains(stderr, "warning") {
		t.Errorf("stderr should contain warning, got: %s", stderr)
	}
}

// Test: successful pass-through
func TestSuccessfulPassThrough(t *testing.T) {
	bin := buildBinary(t)
	outDir := t.TempDir()

	stderr, code := runBinary(t, bin,
		"--input", testdataDir(t, "simple"),
		"--output", outDir,
	)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr)
	}

	// Verify output file exists
	outFile := filepath.Join(outDir, "service.proto")
	if _, err := os.Stat(outFile); err != nil {
		t.Errorf("output file missing: %v", err)
	}
}

// Test: annotation-based filtering via CLI
func TestAnnotationFiltering(t *testing.T) {
	bin := buildBinary(t)
	outDir := t.TempDir()
	cfgDir := t.TempDir()

	cfgPath := filepath.Join(cfgDir, "filter.yaml")
	os.WriteFile(cfgPath, []byte("annotations:\n  - \"HasAnyRole\"\n"), 0o644)

	stderr, code := runBinary(t, bin,
		"--input", testdataDir(t, "annotations"),
		"--output", outDir,
		"--config", cfgPath,
		"--verbose",
	)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr)
	}

	// Verify verbose output contains annotation info
	if !strings.Contains(stderr, "removed") {
		t.Errorf("verbose output should contain 'removed', got: %s", stderr)
	}
	if !strings.Contains(stderr, "by annotation") {
		t.Errorf("verbose output should contain 'by annotation', got: %s", stderr)
	}
	if !strings.Contains(stderr, "orphaned") {
		t.Errorf("verbose output should contain 'orphaned', got: %s", stderr)
	}

	// service.proto should exist (has remaining methods)
	serviceOut := filepath.Join(outDir, "service.proto")
	serviceContent, err := os.ReadFile(serviceOut)
	if err != nil {
		t.Fatalf("service.proto should exist: %v", err)
	}
	serviceStr := string(serviceContent)
	if !strings.Contains(serviceStr, "ListOrders") {
		t.Error("service.proto should contain ListOrders (non-annotated)")
	}
	if strings.Contains(serviceStr, "rpc CreateOrder") {
		t.Error("service.proto should NOT contain rpc CreateOrder (annotated)")
	}

	// internal_only.proto should NOT exist (all methods annotated → empty service → empty file)
	internalOut := filepath.Join(outDir, "internal_only.proto")
	if _, err := os.Stat(internalOut); err == nil {
		t.Error("internal_only.proto should NOT be in output (all methods annotated)")
	}
}

// Test: backward compatibility (no annotations key in config)
func TestBackwardCompatibilityNoAnnotations(t *testing.T) {
	bin := buildBinary(t)
	outDir := t.TempDir()
	cfgDir := t.TempDir()

	// Config with only include, no annotations key
	cfgPath := filepath.Join(cfgDir, "filter.yaml")
	os.WriteFile(cfgPath, []byte("include:\n  - \"filter.OrderService\"\n"), 0o644)

	stderr, code := runBinary(t, bin,
		"--input", testdataDir(t, "filter"),
		"--output", outDir,
		"--config", cfgPath,
		"--verbose",
	)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr)
	}

	// Verify no annotation output in verbose (no annotations configured)
	if strings.Contains(stderr, "by annotation") {
		t.Errorf("verbose should NOT mention annotations when none configured, got: %s", stderr)
	}

	// orders.proto should exist
	if _, err := os.Stat(filepath.Join(outDir, "orders.proto")); err != nil {
		t.Error("orders.proto should be in output")
	}
	// users.proto should NOT exist
	if _, err := os.Stat(filepath.Join(outDir, "users.proto")); err == nil {
		t.Error("users.proto should NOT be in output")
	}
}

// Test: comment conversion via CLI (pass-through mode)
func TestCommentConversionCLI(t *testing.T) {
	bin := buildBinary(t)
	outDir := t.TempDir()

	stderr, code := runBinary(t, bin,
		"--input", testdataDir(t, "comments"),
		"--output", outDir,
	)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", code, stderr)
	}

	// Compare each output file against golden file
	goldenDir := filepath.Join(testdataDir(t, "comments"), "expected")
	for _, name := range []string{"commented.proto", "multiline.proto", "block_comments.proto"} {
		actual, err := os.ReadFile(filepath.Join(outDir, name))
		if err != nil {
			t.Fatalf("read output %s: %v", name, err)
		}
		expected, err := os.ReadFile(filepath.Join(goldenDir, name))
		if err != nil {
			t.Fatalf("read golden %s: %v", name, err)
		}
		if string(actual) != string(expected) {
			t.Errorf("%s: output does not match golden file", name)
		}
	}
}

// Test: comment conversion works alongside filtering
func TestCommentConversionWithFiltering(t *testing.T) {
	bin := buildBinary(t)
	outDir := t.TempDir()
	cfgDir := t.TempDir()

	// Config that includes everything (pass-through filtering)
	cfgPath := filepath.Join(cfgDir, "filter.yaml")
	os.WriteFile(cfgPath, []byte("include:\n  - \"comments.*\"\n"), 0o644)

	stderr, code := runBinary(t, bin,
		"--input", testdataDir(t, "comments"),
		"--output", outDir,
		"--config", cfgPath,
	)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", code, stderr)
	}

	// Verify block comments are converted in filtered output
	content, err := os.ReadFile(filepath.Join(outDir, "multiline.proto"))
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	out := string(content)

	// Should NOT contain block comment markers
	if strings.Contains(out, "/*") {
		t.Error("output should not contain block comment markers after conversion")
	}
	// Should contain converted single-line comments
	if !strings.Contains(out, "// PaymentStatus tracks") {
		t.Error("output should contain converted single-line comment")
	}
}

// Test: successful filtering via CLI
func TestSuccessfulFiltering(t *testing.T) {
	bin := buildBinary(t)
	outDir := t.TempDir()
	cfgDir := t.TempDir()

	cfgPath := filepath.Join(cfgDir, "filter.yaml")
	os.WriteFile(cfgPath, []byte("include:\n  - \"filter.OrderService\"\n"), 0o644)

	stderr, code := runBinary(t, bin,
		"--input", testdataDir(t, "filter"),
		"--output", outDir,
		"--config", cfgPath,
		"--verbose",
	)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr)
	}

	// orders.proto should exist
	if _, err := os.Stat(filepath.Join(outDir, "orders.proto")); err != nil {
		t.Error("orders.proto should be in output")
	}

	// users.proto should NOT exist
	if _, err := os.Stat(filepath.Join(outDir, "users.proto")); err == nil {
		t.Error("users.proto should NOT be in output")
	}

	// Verify verbose output
	if !strings.Contains(stderr, "included") {
		t.Errorf("verbose should report included count, got: %s", stderr)
	}
}
