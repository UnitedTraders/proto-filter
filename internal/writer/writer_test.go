package writer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/emicklei/proto"

	"github.com/unitedtraders/proto-filter/internal/parser"
)

func testdataDir(t *testing.T, sub string) string {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("..", "..", "testdata", sub))
	if err != nil {
		t.Fatalf("testdata path: %v", err)
	}
	return dir
}

// T012: Test proto file writing round-trip
func TestWriteProtoFile(t *testing.T) {
	inputPath := filepath.Join(testdataDir(t, "simple"), "service.proto")
	def, err := parser.ParseProtoFile(inputPath)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, "service.proto")

	if err := WriteProtoFile(def, outputPath); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("output file missing: %v", err)
	}

	// Verify output is parseable
	reparsed, err := parser.ParseProtoFile(outputPath)
	if err != nil {
		t.Fatalf("reparse: %v", err)
	}

	// Verify key elements present in reparsed output
	var hasService, hasMsg, hasEnum bool
	var pkgName string
	proto.Walk(reparsed,
		proto.WithPackage(func(p *proto.Package) {
			pkgName = p.Name
		}),
		proto.WithService(func(s *proto.Service) {
			if s.Name == "OrderService" {
				hasService = true
			}
		}),
		proto.WithMessage(func(m *proto.Message) {
			if m.Name == "CreateOrderRequest" {
				hasMsg = true
			}
		}),
		proto.WithEnum(func(e *proto.Enum) {
			if e.Name == "Status" {
				hasEnum = true
			}
		}),
	)

	if pkgName != "simple" {
		t.Errorf("package: expected simple, got %s", pkgName)
	}
	if !hasService {
		t.Error("missing OrderService in output")
	}
	if !hasMsg {
		t.Error("missing CreateOrderRequest in output")
	}
	if !hasEnum {
		t.Error("missing Status enum in output")
	}
}

func TestWriteProtoFileCreatesDirectories(t *testing.T) {
	inputPath := filepath.Join(testdataDir(t, "simple"), "service.proto")
	def, err := parser.ParseProtoFile(inputPath)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	outputDir := t.TempDir()
	deepPath := filepath.Join(outputDir, "a", "b", "c", "service.proto")

	if err := WriteProtoFile(def, deepPath); err != nil {
		t.Fatalf("write to deep path: %v", err)
	}

	if _, err := os.Stat(deepPath); err != nil {
		t.Fatalf("file not created at deep path: %v", err)
	}
}

// T013: Test comment preservation through write
func TestWritePreservesComments(t *testing.T) {
	inputPath := filepath.Join(testdataDir(t, "comments"), "commented.proto")
	def, err := parser.ParseProtoFile(inputPath)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, "commented.proto")

	if err := WriteProtoFile(def, outputPath); err != nil {
		t.Fatalf("write: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	out := string(content)

	// Verify leading comments preserved
	comments := []string{
		"OrderService handles order operations",
		"CreateOrder creates a new order",
		"OrderRequest is used to create a new order",
		"OrderResponse contains the created order details",
		"Status represents the state of an entity",
		"The name of the order",
		"The price in USD",
	}
	for _, c := range comments {
		if !strings.Contains(out, c) {
			t.Errorf("comment not preserved: %q", c)
		}
	}

	// Verify inline comments preserved
	inlineComments := []string{
		"Must be positive",
		"Unique identifier",
		"Currently active",
		"Main creation endpoint",
	}
	for _, c := range inlineComments {
		if !strings.Contains(out, c) {
			t.Errorf("inline comment not preserved: %q", c)
		}
	}
}

// Test multiline block comments (/* ... */) survive the write round-trip
func TestWritePreservesMultilineBlockComments(t *testing.T) {
	inputPath := filepath.Join(testdataDir(t, "comments"), "multiline.proto")
	def, err := parser.ParseProtoFile(inputPath)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, "multiline.proto")

	if err := WriteProtoFile(def, outputPath); err != nil {
		t.Fatalf("write: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	out := string(content)

	// Verify multiline block comments on top-level definitions
	blockComments := []string{
		"PaymentStatus tracks the lifecycle",
		"settlement or failure",
		"PaymentRequest contains all information",
		"amount: the payment amount",
		"currency: ISO 4217 currency code",
		"recipient: the payee identifier",
		"PaymentService provides operations",
		"bearer token",
		"MakePayment initiates a new payment",
		"pending status",
	}
	for _, c := range blockComments {
		if !strings.Contains(out, c) {
			t.Errorf("multiline block comment not preserved: %q", c)
		}
	}

	// Verify block comments on enum values
	enumValueComments := []string{
		"created but not yet processed",
		"successfully processed",
		"funds have been transferred",
		"insufficient funds",
	}
	for _, c := range enumValueComments {
		if !strings.Contains(out, c) {
			t.Errorf("enum value block comment not preserved: %q", c)
		}
	}

	// Verify block comments on message fields
	fieldComments := []string{
		"Amount in minor currency units",
		"ISO 4217 currency code",
		"Identifier of the payment recipient",
		"account ID or email address",
	}
	for _, c := range fieldComments {
		if !strings.Contains(out, c) {
			t.Errorf("field block comment not preserved: %q", c)
		}
	}
}

// Test consecutive single-line // comments survive the write round-trip
func TestWritePreservesConsecutiveSingleLineComments(t *testing.T) {
	inputPath := filepath.Join(testdataDir(t, "comments"), "multiline.proto")
	def, err := parser.ParseProtoFile(inputPath)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, "multiline.proto")

	if err := WriteProtoFile(def, outputPath); err != nil {
		t.Fatalf("write: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	out := string(content)

	// PaymentResponse has 3 consecutive // lines
	consecutiveComments := []string{
		"PaymentResponse is returned after initiating",
		"transaction ID and current status",
		"query for updates",
	}
	for _, c := range consecutiveComments {
		if !strings.Contains(out, c) {
			t.Errorf("consecutive single-line comment not preserved: %q", c)
		}
	}

	// PaymentResponse field comments (multi-line //)
	fieldComments := []string{
		"Unique transaction identifier",
		"Format: UUID v4",
		"Current status of the payment",
	}
	for _, c := range fieldComments {
		if !strings.Contains(out, c) {
			t.Errorf("field consecutive comment not preserved: %q", c)
		}
	}

	// Verify output is still parseable after round-trip
	_, err = parser.ParseProtoFile(outputPath)
	if err != nil {
		t.Fatalf("output not parseable after round-trip: %v", err)
	}
}
