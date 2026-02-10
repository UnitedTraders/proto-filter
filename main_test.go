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
	srcDir := testdataDir(t, "annotations")
	inputDir := t.TempDir()
	data, err := os.ReadFile(filepath.Join(srcDir, name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	if err := os.WriteFile(filepath.Join(inputDir, name), data, 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return inputDir
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
