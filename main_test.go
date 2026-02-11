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

// T015: Test service-level annotation filtering via CLI
func TestServiceAnnotationFilteringCLI(t *testing.T) {
	bin := buildBinary(t)
	outDir := t.TempDir()
	cfgDir := t.TempDir()

	cfgPath := filepath.Join(cfgDir, "filter.yaml")
	os.WriteFile(cfgPath, []byte("annotations:\n  - \"Internal\"\n"), 0o644)

	stderr, code := runBinary(t, bin,
		"--input", testdataDir(t, "annotations"),
		"--output", outDir,
		"--config", cfgPath,
	)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", code, stderr)
	}

	// service_annotated.proto should exist with only OrderService
	serviceOut := filepath.Join(outDir, "service_annotated.proto")
	content, err := os.ReadFile(serviceOut)
	if err != nil {
		t.Fatalf("service_annotated.proto should exist: %v", err)
	}
	out := string(content)
	if !strings.Contains(out, "OrderService") {
		t.Error("output should contain OrderService (not annotated)")
	}
	if strings.Contains(out, "AdminService") {
		t.Error("output should NOT contain AdminService (annotated @Internal)")
	}
	if strings.Contains(out, "ResetCache") {
		t.Error("output should NOT contain ResetCache methods (AdminService removed)")
	}
	if strings.Contains(out, "MetricsRequest") {
		t.Error("output should NOT contain MetricsRequest (orphaned after AdminService removal)")
	}
}

// T016: Test mixed service+method annotation filtering via CLI
func TestMixedServiceMethodAnnotationFilteringCLI(t *testing.T) {
	bin := buildBinary(t)
	outDir := t.TempDir()
	cfgDir := t.TempDir()

	cfgPath := filepath.Join(cfgDir, "filter.yaml")
	os.WriteFile(cfgPath, []byte("annotations:\n  - \"Internal\"\n  - \"HasAnyRole\"\n"), 0o644)

	stderr, code := runBinary(t, bin,
		"--input", testdataDir(t, "annotations"),
		"--output", outDir,
		"--config", cfgPath,
	)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", code, stderr)
	}

	// mixed_annotations.proto should exist with only TradingService.GetPositions
	mixedOut := filepath.Join(outDir, "mixed_annotations.proto")
	content, err := os.ReadFile(mixedOut)
	if err != nil {
		t.Fatalf("mixed_annotations.proto should exist: %v", err)
	}
	out := string(content)
	if strings.Contains(out, "MonitoringService") {
		t.Error("output should NOT contain MonitoringService (annotated @Internal)")
	}
	if !strings.Contains(out, "TradingService") {
		t.Error("output should contain TradingService (not annotated at service level)")
	}
	if strings.Contains(out, "ForceClose") {
		t.Error("output should NOT contain ForceClose (annotated @HasAnyRole at method level)")
	}
	if !strings.Contains(out, "GetPositions") {
		t.Error("output should contain GetPositions (not annotated)")
	}

	// internal_only.proto should NOT exist (all methods have @HasAnyRole)
	internalOut := filepath.Join(outDir, "internal_only.proto")
	if _, err := os.Stat(internalOut); err == nil {
		t.Error("internal_only.proto should NOT be in output (all methods annotated)")
	}
}

// T017: Test service annotation filtering verbose output
func TestServiceAnnotationFilteringVerbose(t *testing.T) {
	bin := buildBinary(t)
	outDir := t.TempDir()
	cfgDir := t.TempDir()

	cfgPath := filepath.Join(cfgDir, "filter.yaml")
	os.WriteFile(cfgPath, []byte("annotations:\n  - \"Internal\"\n"), 0o644)

	stderr, code := runBinary(t, bin,
		"--input", testdataDir(t, "annotations"),
		"--output", outDir,
		"--config", cfgPath,
		"--verbose",
	)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", code, stderr)
	}

	if !strings.Contains(stderr, "services by annotation") {
		t.Errorf("verbose output should contain 'services by annotation', got: %s", stderr)
	}
}

// setupCrossFileInput copies only the .proto input files (not expected/) to a temp dir
func setupCrossFileInput(t *testing.T) string {
	t.Helper()
	srcDir := testdataDir(t, "crossfile")
	inputDir := t.TempDir()
	for _, name := range []string{"common.proto", "orders.proto", "payments.proto"} {
		data, err := os.ReadFile(filepath.Join(srcDir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if err := os.WriteFile(filepath.Join(inputDir, name), data, 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return inputDir
}

// T010: Cross-file annotation filtering CLI integration test
func TestCrossFileAnnotationFilteringCLI(t *testing.T) {
	bin := buildBinary(t)
	inputDir := setupCrossFileInput(t)
	outDir := t.TempDir()
	cfgDir := t.TempDir()

	cfgPath := filepath.Join(cfgDir, "filter.yaml")
	os.WriteFile(cfgPath, []byte("annotations:\n  - \"Internal\"\n  - \"HasAnyRole\"\n"), 0o644)

	stderr, code := runBinary(t, bin,
		"--input", inputDir,
		"--output", outDir,
		"--config", cfgPath,
		"--verbose",
	)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", code, stderr)
	}

	// common.proto should exist with Pagination and Money preserved
	commonOut := filepath.Join(outDir, "common.proto")
	commonContent, err := os.ReadFile(commonOut)
	if err != nil {
		t.Fatalf("common.proto should exist in output: %v", err)
	}
	commonStr := string(commonContent)
	if !strings.Contains(commonStr, "Pagination") {
		t.Error("common.proto should contain Pagination (referenced by surviving OrderService)")
	}
	if !strings.Contains(commonStr, "Money") {
		t.Error("common.proto should contain Money (referenced by surviving OrderService)")
	}

	// orders.proto should exist with only ListOrders method
	ordersOut := filepath.Join(outDir, "orders.proto")
	ordersContent, err := os.ReadFile(ordersOut)
	if err != nil {
		t.Fatalf("orders.proto should exist: %v", err)
	}
	ordersStr := string(ordersContent)
	if !strings.Contains(ordersStr, "OrderService") {
		t.Error("orders.proto should contain OrderService")
	}
	if !strings.Contains(ordersStr, "ListOrders") {
		t.Error("orders.proto should contain ListOrders (not annotated)")
	}
	if strings.Contains(ordersStr, "GetOrderDetails") {
		t.Error("orders.proto should NOT contain GetOrderDetails (annotated @HasAnyRole)")
	}
	if strings.Contains(ordersStr, "GetOrderDetailsRequest") {
		t.Error("orders.proto should NOT contain GetOrderDetailsRequest (orphaned)")
	}
	if strings.Contains(ordersStr, "GetOrderDetailsResponse") {
		t.Error("orders.proto should NOT contain GetOrderDetailsResponse (orphaned)")
	}

	// payments.proto should NOT exist (PaymentService removed, no remaining definitions)
	paymentsOut := filepath.Join(outDir, "payments.proto")
	if _, err := os.Stat(paymentsOut); err == nil {
		t.Error("payments.proto should NOT be in output (PaymentService removed by @Internal)")
	}
}

// T011: Cross-file shared types survive partial service removal
func TestCrossFileSharedTypesSurvivePartialServiceRemoval(t *testing.T) {
	bin := buildBinary(t)
	inputDir := setupCrossFileInput(t)
	outDir := t.TempDir()
	cfgDir := t.TempDir()

	// Only filter @Internal (service-level), not @HasAnyRole (method-level)
	cfgPath := filepath.Join(cfgDir, "filter.yaml")
	os.WriteFile(cfgPath, []byte("annotations:\n  - \"Internal\"\n"), 0o644)

	stderr, code := runBinary(t, bin,
		"--input", inputDir,
		"--output", outDir,
		"--config", cfgPath,
	)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", code, stderr)
	}

	// common.proto should exist with all shared types preserved
	commonOut := filepath.Join(outDir, "common.proto")
	commonContent, err := os.ReadFile(commonOut)
	if err != nil {
		t.Fatalf("common.proto should exist: %v", err)
	}
	commonStr := string(commonContent)
	if !strings.Contains(commonStr, "Pagination") {
		t.Error("Pagination should survive (referenced by surviving OrderService)")
	}
	if !strings.Contains(commonStr, "Money") {
		t.Error("Money should survive (referenced by surviving OrderService)")
	}

	// orders.proto should be fully intact (both methods survive)
	ordersOut := filepath.Join(outDir, "orders.proto")
	ordersContent, err := os.ReadFile(ordersOut)
	if err != nil {
		t.Fatalf("orders.proto should exist: %v", err)
	}
	ordersStr := string(ordersContent)
	if !strings.Contains(ordersStr, "ListOrders") {
		t.Error("orders.proto should contain ListOrders")
	}
	if !strings.Contains(ordersStr, "GetOrderDetails") {
		t.Error("orders.proto should contain GetOrderDetails (not filtered by @Internal)")
	}

	// payments.proto should NOT exist
	paymentsOut := filepath.Join(outDir, "payments.proto")
	if _, err := os.Stat(paymentsOut); err == nil {
		t.Error("payments.proto should NOT be in output")
	}
}

// T012: Cross-file golden file comparison
func TestCrossFileGoldenFileComparison(t *testing.T) {
	bin := buildBinary(t)
	inputDir := setupCrossFileInput(t)
	outDir := t.TempDir()
	cfgDir := t.TempDir()

	cfgPath := filepath.Join(cfgDir, "filter.yaml")
	os.WriteFile(cfgPath, []byte("annotations:\n  - \"Internal\"\n  - \"HasAnyRole\"\n"), 0o644)

	stderr, code := runBinary(t, bin,
		"--input", inputDir,
		"--output", outDir,
		"--config", cfgPath,
	)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", code, stderr)
	}

	// Compare each output file against golden files
	goldenDir := filepath.Join(testdataDir(t, "crossfile"), "expected")

	for _, name := range []string{"common.proto", "orders.proto"} {
		actual, err := os.ReadFile(filepath.Join(outDir, name))
		if err != nil {
			t.Fatalf("read output %s: %v", name, err)
		}
		expected, err := os.ReadFile(filepath.Join(goldenDir, name))
		if err != nil {
			t.Fatalf("read golden %s: %v", name, err)
		}
		if string(actual) != string(expected) {
			t.Errorf("%s: output does not match golden file\n--- ACTUAL ---\n%s\n--- EXPECTED ---\n%s",
				name, string(actual), string(expected))
		}
	}

	// payments.proto should not exist in output
	if _, err := os.Stat(filepath.Join(outDir, "payments.proto")); err == nil {
		t.Error("payments.proto should NOT be in output (no golden file expected)")
	}
}

// setupSingleProtoInput copies a single proto file from testdata/annotations to a temp dir
func setupSingleProtoInput(t *testing.T, name string) string {
	t.Helper()
	// Try annotations dir first, then substitution dir
	for _, sub := range []string{"annotations", "substitution"} {
		srcPath := filepath.Join(testdataDir(t, sub), name)
		data, err := os.ReadFile(srcPath)
		if err != nil {
			continue
		}
		inputDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(inputDir, name), data, 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		return inputDir
	}
	t.Fatalf("proto file %s not found in testdata", name)
	return ""
}

// T012: CLI integration test for bracket-style annotation filtering
func TestBracketAnnotationFilteringCLI(t *testing.T) {
	bin := buildBinary(t)
	inputDir := setupSingleProtoInput(t, "bracket_service.proto")
	outDir := t.TempDir()
	cfgDir := t.TempDir()

	cfgPath := filepath.Join(cfgDir, "filter.yaml")
	os.WriteFile(cfgPath, []byte("annotations:\n  - \"HasAnyRole\"\n"), 0o644)

	stderr, code := runBinary(t, bin,
		"--input", inputDir,
		"--output", outDir,
		"--config", cfgPath,
		"--verbose",
	)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", code, stderr)
	}

	// bracket_service.proto should exist with only ListOrders
	outFile := filepath.Join(outDir, "bracket_service.proto")
	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("bracket_service.proto should exist: %v", err)
	}
	out := string(content)

	if !strings.Contains(out, "ListOrders") {
		t.Error("output should contain ListOrders (not annotated)")
	}
	if strings.Contains(out, "rpc CreateOrder") {
		t.Error("output should NOT contain rpc CreateOrder (annotated [HasAnyRole])")
	}
	if strings.Contains(out, "rpc DeleteOrder") {
		t.Error("output should NOT contain rpc DeleteOrder (annotated [HasAnyRole])")
	}
	if strings.Contains(out, "CreateOrderRequest") {
		t.Error("output should NOT contain CreateOrderRequest (orphaned)")
	}
}

// T016: CLI integration test for mixed-style annotation filtering
func TestMixedStyleAnnotationFilteringCLI(t *testing.T) {
	bin := buildBinary(t)
	inputDir := setupSingleProtoInput(t, "mixed_styles.proto")
	outDir := t.TempDir()
	cfgDir := t.TempDir()

	cfgPath := filepath.Join(cfgDir, "filter.yaml")
	os.WriteFile(cfgPath, []byte("annotations:\n  - \"HasAnyRole\"\n"), 0o644)

	stderr, code := runBinary(t, bin,
		"--input", inputDir,
		"--output", outDir,
		"--config", cfgPath,
		"--verbose",
	)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", code, stderr)
	}

	// mixed_styles.proto should exist with only ListOrders
	outFile := filepath.Join(outDir, "mixed_styles.proto")
	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("mixed_styles.proto should exist: %v", err)
	}
	out := string(content)

	if !strings.Contains(out, "ListOrders") {
		t.Error("output should contain ListOrders (not annotated)")
	}
	if strings.Contains(out, "rpc CreateOrder") {
		t.Error("output should NOT contain rpc CreateOrder (annotated @HasAnyRole)")
	}
	if strings.Contains(out, "rpc DeleteOrder") {
		t.Error("output should NOT contain rpc DeleteOrder (annotated [HasAnyRole])")
	}
	if strings.Contains(out, "CreateOrderRequest") {
		t.Error("output should NOT contain CreateOrderRequest (orphaned)")
	}
	if strings.Contains(out, "DeleteOrderRequest") {
		t.Error("output should NOT contain DeleteOrderRequest (orphaned)")
	}
}

// T017: CLI integration test for include-mode annotation filtering
func TestIncludeAnnotationFilteringCLI(t *testing.T) {
	bin := buildBinary(t)
	inputDir := setupSingleProtoInput(t, "include_service.proto")
	outDir := t.TempDir()
	cfgDir := t.TempDir()

	cfgPath := filepath.Join(cfgDir, "filter.yaml")
	os.WriteFile(cfgPath, []byte("annotations:\n  include:\n    - \"Public\"\n"), 0o644)

	stderr, code := runBinary(t, bin,
		"--input", inputDir,
		"--output", outDir,
		"--config", cfgPath,
		"--verbose",
	)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", code, stderr)
	}

	// include_service.proto should exist with only annotated methods
	outFile := filepath.Join(outDir, "include_service.proto")
	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("include_service.proto should exist: %v", err)
	}
	out := string(content)

	if !strings.Contains(out, "rpc CreateOrder") {
		t.Error("output should contain rpc CreateOrder (has @Public)")
	}
	if !strings.Contains(out, "rpc DeleteOrder") {
		t.Error("output should contain rpc DeleteOrder (has [Public])")
	}
	if strings.Contains(out, "ListOrders") {
		t.Error("output should NOT contain ListOrders (no Public annotation)")
	}
	if strings.Contains(out, "ListOrdersRequest") {
		t.Error("output should NOT contain ListOrdersRequest (orphaned)")
	}
}

// T021: CLI integration test for structured exclude annotation filtering
func TestStructuredExcludeAnnotationFilteringCLI(t *testing.T) {
	bin := buildBinary(t)
	inputDir := setupSingleProtoInput(t, "service.proto")
	outDir := t.TempDir()
	cfgDir := t.TempDir()

	cfgPath := filepath.Join(cfgDir, "filter.yaml")
	os.WriteFile(cfgPath, []byte("annotations:\n  exclude:\n    - \"HasAnyRole\"\n"), 0o644)

	stderr, code := runBinary(t, bin,
		"--input", inputDir,
		"--output", outDir,
		"--config", cfgPath,
	)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", code, stderr)
	}

	outFile := filepath.Join(outDir, "service.proto")
	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("service.proto should exist: %v", err)
	}
	out := string(content)

	if !strings.Contains(out, "ListOrders") {
		t.Error("output should contain ListOrders (not annotated)")
	}
	if strings.Contains(out, "rpc CreateOrder") {
		t.Error("output should NOT contain rpc CreateOrder (annotated @HasAnyRole)")
	}
	if strings.Contains(out, "rpc DeleteOrder") {
		t.Error("output should NOT contain rpc DeleteOrder (annotated @HasAnyRole)")
	}
}

// --- Annotation Substitution CLI Tests (Feature 008) ---

// T013 (008): CLI integration test for substitution replacement
func TestSubstitutionReplacementCLI(t *testing.T) {
	bin := buildBinary(t)
	inputDir := setupSingleProtoInput(t, "substitution_service.proto")
	outDir := t.TempDir()
	cfgDir := t.TempDir()

	cfgPath := filepath.Join(cfgDir, "filter.yaml")
	os.WriteFile(cfgPath, []byte(`substitutions:
  HasAnyRole: "Requires authentication"
  Internal: "For internal use only"
  Public: "Available to all users"
`), 0o644)

	stderr, code := runBinary(t, bin,
		"--input", inputDir,
		"--output", outDir,
		"--config", cfgPath,
		"--verbose",
	)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", code, stderr)
	}

	outFile := filepath.Join(outDir, "substitution_service.proto")
	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("output file missing: %v", err)
	}
	out := string(content)

	if !strings.Contains(out, "Requires authentication") {
		t.Error("output should contain 'Requires authentication'")
	}
	if !strings.Contains(out, "For internal use only") {
		t.Error("output should contain 'For internal use only'")
	}
	if !strings.Contains(out, "Available to all users") {
		t.Error("output should contain 'Available to all users'")
	}
	if strings.Contains(out, "@HasAnyRole") {
		t.Error("output should NOT contain @HasAnyRole")
	}
	if strings.Contains(out, "@Internal") {
		t.Error("output should NOT contain @Internal")
	}
	if strings.Contains(out, "[Public]") {
		t.Error("output should NOT contain [Public]")
	}
	if !strings.Contains(stderr, "substituted") {
		t.Errorf("verbose output should contain 'substituted', got: %s", stderr)
	}
}

// T020 (008): CLI integration test for empty substitution removal
func TestSubstitutionEmptyRemovalCLI(t *testing.T) {
	bin := buildBinary(t)
	inputDir := setupSingleProtoInput(t, "substitution_service.proto")
	outDir := t.TempDir()
	cfgDir := t.TempDir()

	cfgPath := filepath.Join(cfgDir, "filter.yaml")
	os.WriteFile(cfgPath, []byte(`substitutions:
  HasAnyRole: ""
  Internal: ""
  Public: ""
`), 0o644)

	stderr, code := runBinary(t, bin,
		"--input", inputDir,
		"--output", outDir,
		"--config", cfgPath,
	)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", code, stderr)
	}

	outFile := filepath.Join(outDir, "substitution_service.proto")
	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("output file missing: %v", err)
	}
	out := string(content)

	if strings.Contains(out, "@HasAnyRole") {
		t.Error("output should NOT contain @HasAnyRole")
	}
	if strings.Contains(out, "@Internal") {
		t.Error("output should NOT contain @Internal")
	}
	if strings.Contains(out, "[Public]") {
		t.Error("output should NOT contain [Public]")
	}
	// Descriptive text should remain
	if !strings.Contains(out, "Creates a new order") {
		t.Error("output should contain 'Creates a new order'")
	}
	if !strings.Contains(out, "Lists all orders") {
		t.Error("output should contain 'Lists all orders'")
	}
}

// T027 (008): CLI integration test for strict substitution error
func TestStrictSubstitutionErrorCLI(t *testing.T) {
	bin := buildBinary(t)
	inputDir := setupSingleProtoInput(t, "substitution_service.proto")
	outDir := t.TempDir()
	cfgDir := t.TempDir()

	cfgPath := filepath.Join(cfgDir, "filter.yaml")
	os.WriteFile(cfgPath, []byte(`substitutions:
  HasAnyRole: "Auth required"
strict_substitutions: true
`), 0o644)

	stderr, code := runBinary(t, bin,
		"--input", inputDir,
		"--output", outDir,
		"--config", cfgPath,
	)
	if code != 2 {
		t.Errorf("expected exit code 2, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stderr, "unsubstituted annotations") {
		t.Errorf("stderr should contain 'unsubstituted annotations', got: %s", stderr)
	}
	if !strings.Contains(stderr, "Internal") {
		t.Errorf("stderr should list 'Internal' as missing, got: %s", stderr)
	}
	if !strings.Contains(stderr, "Public") {
		t.Errorf("stderr should list 'Public' as missing, got: %s", stderr)
	}

	// No output files should be written
	outFile := filepath.Join(outDir, "substitution_service.proto")
	if _, err := os.Stat(outFile); err == nil {
		t.Error("no output files should be written on strict failure")
	}
}

// T028 (008): CLI integration test for strict substitution success
func TestStrictSubstitutionSuccessCLI(t *testing.T) {
	bin := buildBinary(t)
	inputDir := setupSingleProtoInput(t, "substitution_service.proto")
	outDir := t.TempDir()
	cfgDir := t.TempDir()

	cfgPath := filepath.Join(cfgDir, "filter.yaml")
	os.WriteFile(cfgPath, []byte(`substitutions:
  HasAnyRole: "Auth required"
  Internal: "Internal only"
  Public: "Public API"
strict_substitutions: true
`), 0o644)

	stderr, code := runBinary(t, bin,
		"--input", inputDir,
		"--output", outDir,
		"--config", cfgPath,
	)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr)
	}

	outFile := filepath.Join(outDir, "substitution_service.proto")
	if _, err := os.Stat(outFile); err != nil {
		t.Error("output file should be written on strict success")
	}
}

// T029 (008): CLI integration test for strict mode with no annotations
func TestStrictSubstitutionNoAnnotationsCLI(t *testing.T) {
	bin := buildBinary(t)
	// Create a proto file with no annotations
	inputDir := t.TempDir()
	os.WriteFile(filepath.Join(inputDir, "plain.proto"), []byte(`syntax = "proto3";
package plain;
message Foo { string bar = 1; }
`), 0o644)
	outDir := t.TempDir()
	cfgDir := t.TempDir()

	cfgPath := filepath.Join(cfgDir, "filter.yaml")
	os.WriteFile(cfgPath, []byte(`strict_substitutions: true
`), 0o644)

	stderr, code := runBinary(t, bin,
		"--input", inputDir,
		"--output", outDir,
		"--config", cfgPath,
	)
	if code != 0 {
		t.Errorf("expected exit code 0 (no annotations to flag), got %d; stderr: %s", code, stderr)
	}
}

// T032 (008): CLI integration test for substitution with annotation filtering
func TestSubstitutionWithAnnotationFilterCLI(t *testing.T) {
	bin := buildBinary(t)
	inputDir := setupSingleProtoInput(t, "substitution_service.proto")
	outDir := t.TempDir()
	cfgDir := t.TempDir()

	cfgPath := filepath.Join(cfgDir, "filter.yaml")
	os.WriteFile(cfgPath, []byte(`annotations:
  exclude:
    - "Internal"
substitutions:
  HasAnyRole: "Auth required"
  Public: "Public API"
`), 0o644)

	stderr, code := runBinary(t, bin,
		"--input", inputDir,
		"--output", outDir,
		"--config", cfgPath,
	)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", code, stderr)
	}

	outFile := filepath.Join(outDir, "substitution_service.proto")
	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("output file missing: %v", err)
	}
	out := string(content)

	// Internal method should be removed entirely
	if strings.Contains(out, "DeleteOrder") {
		t.Error("output should NOT contain DeleteOrder (filtered by @Internal)")
	}
	// HasAnyRole should be substituted
	if !strings.Contains(out, "Auth required") {
		t.Error("output should contain 'Auth required' (substituted)")
	}
	if strings.Contains(out, "@HasAnyRole") {
		t.Error("output should NOT contain @HasAnyRole (substituted)")
	}
	// Public should be substituted
	if !strings.Contains(out, "Public API") {
		t.Error("output should contain 'Public API' (substituted)")
	}
}

// --- Annotation Error Location CLI Tests (Feature 009) ---

// T004: CLI integration test for strict substitution error with location lines
func TestStrictSubstitutionErrorWithLocations(t *testing.T) {
	bin := buildBinary(t)
	inputDir := setupSingleProtoInput(t, "substitution_service.proto")
	outDir := t.TempDir()
	cfgDir := t.TempDir()

	cfgPath := filepath.Join(cfgDir, "filter.yaml")
	os.WriteFile(cfgPath, []byte(`substitutions:
  HasAnyRole: "Auth required"
strict_substitutions: true
`), 0o644)

	stderr, code := runBinary(t, bin,
		"--input", inputDir,
		"--output", outDir,
		"--config", cfgPath,
	)
	if code != 2 {
		t.Errorf("expected exit code 2, got %d; stderr: %s", code, stderr)
	}

	// Summary line must be present (backward compatibility)
	if !strings.Contains(stderr, "proto-filter: error: unsubstituted annotations found: Internal, Public") {
		t.Errorf("stderr should contain summary line, got: %s", stderr)
	}

	// Location lines must be present with file:line: token format
	if !strings.Contains(stderr, "  substitution_service.proto:10: @Internal") {
		t.Errorf("stderr should contain location for @Internal at line 10, got: %s", stderr)
	}
	if !strings.Contains(stderr, "  substitution_service.proto:13: [Public]") {
		t.Errorf("stderr should contain location for [Public] at line 13, got: %s", stderr)
	}

	// HasAnyRole should NOT appear in location lines (it has a mapping)
	lines := strings.Split(stderr, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "  ") && strings.Contains(line, "HasAnyRole") {
			t.Errorf("stderr should NOT have location line for HasAnyRole (has mapping), got line: %s", line)
		}
	}

	// No output files should be written
	outFile := filepath.Join(outDir, "substitution_service.proto")
	if _, err := os.Stat(outFile); err == nil {
		t.Error("no output files should be written on strict failure")
	}
}

// T010: CLI integration test for multi-file location ordering (FR-005)
func TestStrictErrorMultiFileOrdering(t *testing.T) {
	bin := buildBinary(t)
	inputDir := t.TempDir()

	// Create two proto files with annotations
	os.WriteFile(filepath.Join(inputDir, "orders.proto"), []byte(`syntax = "proto3";
package orders;
service OrderService {
  // @Deprecated
  rpc OldMethod(OldReq) returns (OldResp);
}
message OldReq {}
message OldResp {}
`), 0o644)

	os.WriteFile(filepath.Join(inputDir, "accounts.proto"), []byte(`syntax = "proto3";
package accounts;
service AccountService {
  // @Internal
  rpc AdminAction(AdminReq) returns (AdminResp);
  // @Deprecated
  rpc LegacyLogin(LoginReq) returns (LoginResp);
}
message AdminReq {}
message AdminResp {}
message LoginReq {}
message LoginResp {}
`), 0o644)

	outDir := t.TempDir()
	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "filter.yaml")
	os.WriteFile(cfgPath, []byte("strict_substitutions: true\n"), 0o644)

	stderr, code := runBinary(t, bin,
		"--input", inputDir,
		"--output", outDir,
		"--config", cfgPath,
	)
	if code != 2 {
		t.Errorf("expected exit code 2, got %d; stderr: %s", code, stderr)
	}

	// Verify summary line contains both annotation names
	if !strings.Contains(stderr, "Deprecated") || !strings.Contains(stderr, "Internal") {
		t.Errorf("summary should list Deprecated and Internal, got: %s", stderr)
	}

	// Parse location lines and verify ordering: accounts.proto before orders.proto
	lines := strings.Split(strings.TrimSpace(stderr), "\n")
	if len(lines) < 4 {
		t.Fatalf("expected at least 4 lines (1 summary + 3 locations), got %d: %v", len(lines), lines)
	}

	// Location lines (skip summary)
	locLines := lines[1:]
	// accounts.proto locations should come before orders.proto (alphabetical)
	foundAccountsFirst := false
	foundOrders := false
	for _, line := range locLines {
		if strings.Contains(line, "accounts.proto") && !foundOrders {
			foundAccountsFirst = true
		}
		if strings.Contains(line, "orders.proto") {
			foundOrders = true
		}
	}
	if !foundAccountsFirst {
		t.Errorf("accounts.proto locations should come before orders.proto; got lines: %v", locLines)
	}

	// Within accounts.proto, @Internal (line 4) should come before @Deprecated (line 6)
	var accountsLines []string
	for _, line := range locLines {
		if strings.Contains(line, "accounts.proto") {
			accountsLines = append(accountsLines, line)
		}
	}
	if len(accountsLines) == 2 {
		if strings.Contains(accountsLines[0], "Deprecated") && strings.Contains(accountsLines[1], "Internal") {
			t.Errorf("within accounts.proto, @Internal should come before @Deprecated; got: %v", accountsLines)
		}
	}
}

// T008: CLI integration test verifying summary line is preserved (backward compatibility)
func TestStrictErrorSummaryLinePreserved(t *testing.T) {
	bin := buildBinary(t)
	// Create a single proto file with exactly one annotation
	inputDir := t.TempDir()
	os.WriteFile(filepath.Join(inputDir, "single.proto"), []byte(`syntax = "proto3";
package single;
service Svc {
  // @Deprecated
  rpc Foo(FooReq) returns (FooResp);
}
message FooReq {}
message FooResp {}
`), 0o644)
	outDir := t.TempDir()
	cfgDir := t.TempDir()

	cfgPath := filepath.Join(cfgDir, "filter.yaml")
	os.WriteFile(cfgPath, []byte("strict_substitutions: true\n"), 0o644)

	stderr, code := runBinary(t, bin,
		"--input", inputDir,
		"--output", outDir,
		"--config", cfgPath,
	)
	if code != 2 {
		t.Errorf("expected exit code 2, got %d; stderr: %s", code, stderr)
	}

	lines := strings.Split(strings.TrimSpace(stderr), "\n")
	// First line must be the summary line (backward compatible)
	if len(lines) < 1 || !strings.HasPrefix(lines[0], "proto-filter: error: unsubstituted annotations found: Deprecated") {
		t.Errorf("first line should be the summary line, got: %v", lines)
	}

	// Exactly one location line should follow
	if len(lines) != 2 {
		t.Errorf("expected exactly 2 lines (summary + 1 location), got %d: %v", len(lines), lines)
	}
	if len(lines) >= 2 && !strings.Contains(lines[1], "single.proto:4: @Deprecated") {
		t.Errorf("second line should be location for @Deprecated, got: %s", lines[1])
	}
}

// T025: CLI integration test for mutual exclusivity error
func TestMutualExclusivityErrorCLI(t *testing.T) {
	bin := buildBinary(t)
	outDir := t.TempDir()
	cfgDir := t.TempDir()

	cfgPath := filepath.Join(cfgDir, "filter.yaml")
	os.WriteFile(cfgPath, []byte("annotations:\n  include:\n    - \"Public\"\n  exclude:\n    - \"Internal\"\n"), 0o644)

	stderr, code := runBinary(t, bin,
		"--input", testdataDir(t, "simple"),
		"--output", outDir,
		"--config", cfgPath,
	)
	if code != 2 {
		t.Errorf("expected exit code 2, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stderr, "mutually exclusive") {
		t.Errorf("stderr should contain 'mutually exclusive', got: %s", stderr)
	}
}
