package parser

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/emicklei/proto"
	"github.com/emicklei/proto-contrib/pkg/protofmt"

	"github.com/unitedtraders/proto-filter/internal/writer"
)

func testdataDir(t *testing.T, sub string) string {
	t.Helper()
	// internal/parser -> repo root -> testdata/sub
	dir, err := filepath.Abs(filepath.Join("..", "..", "testdata", sub))
	if err != nil {
		t.Fatalf("testdata path: %v", err)
	}
	return dir
}

// T010: Test proto file discovery
func TestDiscoverProtoFiles(t *testing.T) {
	dir := testdataDir(t, "simple")
	files, err := DiscoverProtoFiles(dir)
	if err != nil {
		t.Fatalf("DiscoverProtoFiles: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d: %v", len(files), files)
	}
	if files[0] != "service.proto" {
		t.Errorf("expected service.proto, got %s", files[0])
	}
}

func TestDiscoverProtoFilesNested(t *testing.T) {
	dir := testdataDir(t, "nested")
	files, err := DiscoverProtoFiles(dir)
	if err != nil {
		t.Fatalf("DiscoverProtoFiles: %v", err)
	}
	sort.Strings(files)
	expected := []string{
		filepath.Join("a", "orders.proto"),
		filepath.Join("b", "users.proto"),
	}
	if len(files) != len(expected) {
		t.Fatalf("expected %d files, got %d: %v", len(expected), len(files), files)
	}
	for i, f := range files {
		if f != expected[i] {
			t.Errorf("file[%d]: expected %s, got %s", i, expected[i], f)
		}
	}
}

func TestDiscoverProtoFilesIgnoresNonProto(t *testing.T) {
	// Create a temp dir with proto and non-proto files
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "a.proto"), []byte("syntax = \"proto3\";"), 0o644)
	os.WriteFile(filepath.Join(tmp, "b.txt"), []byte("not proto"), 0o644)
	os.WriteFile(filepath.Join(tmp, "c.go"), []byte("package main"), 0o644)

	files, err := DiscoverProtoFiles(tmp)
	if err != nil {
		t.Fatalf("DiscoverProtoFiles: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 proto file, got %d: %v", len(files), files)
	}
	if files[0] != "a.proto" {
		t.Errorf("expected a.proto, got %s", files[0])
	}
}

// T011: Test single proto file parsing
func TestParseProtoFile(t *testing.T) {
	path := filepath.Join(testdataDir(t, "simple"), "service.proto")
	def, err := ParseProtoFile(path)
	if err != nil {
		t.Fatalf("ParseProtoFile: %v", err)
	}

	var (
		hasService bool
		hasMsgReq  bool
		hasMsgResp bool
		hasEnum    bool
		pkgName    string
	)
	proto.Walk(def,
		proto.WithPackage(func(p *proto.Package) {
			pkgName = p.Name
		}),
		proto.WithService(func(s *proto.Service) {
			if s.Name == "OrderService" {
				hasService = true
			}
		}),
		proto.WithMessage(func(m *proto.Message) {
			switch m.Name {
			case "CreateOrderRequest":
				hasMsgReq = true
			case "CreateOrderResponse":
				hasMsgResp = true
			}
		}),
		proto.WithEnum(func(e *proto.Enum) {
			if e.Name == "Status" {
				hasEnum = true
			}
		}),
	)

	if pkgName != "simple" {
		t.Errorf("expected package simple, got %s", pkgName)
	}
	if !hasService {
		t.Error("missing OrderService")
	}
	if !hasMsgReq {
		t.Error("missing CreateOrderRequest")
	}
	if !hasMsgResp {
		t.Error("missing CreateOrderResponse")
	}
	if !hasEnum {
		t.Error("missing Status enum")
	}
}

// T014: Integration test for pass-through pipeline
func TestIntegrationPassThrough(t *testing.T) {
	cases := []struct {
		name      string
		fixture   string
		wantFiles []string
	}{
		{
			name:      "simple",
			fixture:   "simple",
			wantFiles: []string{"service.proto"},
		},
		{
			name:    "nested",
			fixture: "nested",
			wantFiles: []string{
				filepath.Join("a", "orders.proto"),
				filepath.Join("b", "users.proto"),
			},
		},
		{
			name:    "imports",
			fixture: "imports",
			wantFiles: []string{
				"common.proto",
				"service.proto",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inputDir := testdataDir(t, tc.fixture)
			outputDir := t.TempDir()

			// Discover
			files, err := DiscoverProtoFiles(inputDir)
			if err != nil {
				t.Fatalf("discover: %v", err)
			}

			sort.Strings(files)
			sort.Strings(tc.wantFiles)

			if len(files) != len(tc.wantFiles) {
				t.Fatalf("expected %d files, got %d", len(tc.wantFiles), len(files))
			}

			// Parse and write each
			for _, rel := range files {
				def, err := ParseProtoFile(filepath.Join(inputDir, rel))
				if err != nil {
					t.Fatalf("parse %s: %v", rel, err)
				}
				outPath := filepath.Join(outputDir, rel)
				if err := writer.WriteProtoFile(def, outPath); err != nil {
					t.Fatalf("write %s: %v", rel, err)
				}
			}

			// Verify output files exist and can be re-parsed
			for _, rel := range tc.wantFiles {
				outPath := filepath.Join(outputDir, rel)
				if _, err := os.Stat(outPath); err != nil {
					t.Errorf("output file missing: %s", rel)
					continue
				}
				// Verify re-parseable
				_, err := ParseProtoFile(outPath)
				if err != nil {
					t.Errorf("output not parseable: %s: %v", rel, err)
				}
			}

			// For imports fixture, verify import statement preserved
			if tc.fixture == "imports" {
				outPath := filepath.Join(outputDir, "service.proto")
				content, _ := os.ReadFile(outPath)
				if !strings.Contains(string(content), `"common.proto"`) {
					t.Error("import statement not preserved in service.proto output")
				}
			}
		})
	}
}

// T011 supplement: verify comments are on parsed AST
func TestParsePreservesComments(t *testing.T) {
	path := filepath.Join(testdataDir(t, "comments"), "commented.proto")
	def, err := ParseProtoFile(path)
	if err != nil {
		t.Fatalf("ParseProtoFile: %v", err)
	}

	var serviceComment string
	var messageComment string
	proto.Walk(def,
		proto.WithService(func(s *proto.Service) {
			if s.Name == "OrderService" && s.Comment != nil {
				serviceComment = strings.Join(s.Comment.Lines, "\n")
			}
		}),
		proto.WithMessage(func(m *proto.Message) {
			if m.Name == "OrderRequest" && m.Comment != nil {
				messageComment = strings.Join(m.Comment.Lines, "\n")
			}
		}),
	)

	if !strings.Contains(serviceComment, "OrderService handles order operations") {
		t.Errorf("service comment not preserved, got: %q", serviceComment)
	}
	if !strings.Contains(messageComment, "OrderRequest is used to create a new order") {
		t.Errorf("message comment not preserved, got: %q", messageComment)
	}
}

// Test multiline block comments (/* ... */) are preserved on parsed AST
func TestParsePreservesMultilineComments(t *testing.T) {
	path := filepath.Join(testdataDir(t, "comments"), "multiline.proto")
	def, err := ParseProtoFile(path)
	if err != nil {
		t.Fatalf("ParseProtoFile: %v", err)
	}

	var serviceComment string
	var messageComment string
	var enumComment string
	proto.Walk(def,
		proto.WithService(func(s *proto.Service) {
			if s.Name == "PaymentService" && s.Comment != nil {
				serviceComment = strings.Join(s.Comment.Lines, "\n")
			}
		}),
		proto.WithMessage(func(m *proto.Message) {
			if m.Name == "PaymentRequest" && m.Comment != nil {
				messageComment = strings.Join(m.Comment.Lines, "\n")
			}
		}),
		proto.WithEnum(func(e *proto.Enum) {
			if e.Name == "PaymentStatus" && e.Comment != nil {
				enumComment = strings.Join(e.Comment.Lines, "\n")
			}
		}),
	)

	// Verify multiline block comment on service
	if !strings.Contains(serviceComment, "PaymentService provides operations") {
		t.Errorf("service block comment not preserved, got: %q", serviceComment)
	}
	if !strings.Contains(serviceComment, "bearer token") {
		t.Errorf("service block comment missing second line, got: %q", serviceComment)
	}

	// Verify multiline block comment on message
	if !strings.Contains(messageComment, "PaymentRequest contains all information") {
		t.Errorf("message block comment not preserved, got: %q", messageComment)
	}
	if !strings.Contains(messageComment, "currency") {
		t.Errorf("message block comment missing field docs, got: %q", messageComment)
	}

	// Verify multiline block comment on enum
	if !strings.Contains(enumComment, "PaymentStatus tracks the lifecycle") {
		t.Errorf("enum block comment not preserved, got: %q", enumComment)
	}
	if !strings.Contains(enumComment, "settlement or failure") {
		t.Errorf("enum block comment missing second line, got: %q", enumComment)
	}
}

// Test multi-line // comments (consecutive single-line) are preserved
func TestParsePreservesConsecutiveSingleLineComments(t *testing.T) {
	path := filepath.Join(testdataDir(t, "comments"), "multiline.proto")
	def, err := ParseProtoFile(path)
	if err != nil {
		t.Fatalf("ParseProtoFile: %v", err)
	}

	var responseComment string
	proto.Walk(def,
		proto.WithMessage(func(m *proto.Message) {
			if m.Name == "PaymentResponse" && m.Comment != nil {
				responseComment = strings.Join(m.Comment.Lines, "\n")
			}
		}),
	)

	// PaymentResponse uses consecutive // comments across 3 lines
	if !strings.Contains(responseComment, "PaymentResponse is returned") {
		t.Errorf("consecutive single-line comment line 1 not preserved, got: %q", responseComment)
	}
	if !strings.Contains(responseComment, "transaction ID and current status") {
		t.Errorf("consecutive single-line comment line 2 not preserved, got: %q", responseComment)
	}
	if !strings.Contains(responseComment, "query for updates") {
		t.Errorf("consecutive single-line comment line 3 not preserved, got: %q", responseComment)
	}
}

// Helper to format AST to string for comparison
func formatToString(def *proto.Proto) string {
	var b strings.Builder
	formatter := protofmt.NewFormatter(&b, "  ")
	formatter.Format(def)
	return b.String()
}
