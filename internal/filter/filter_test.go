package filter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/emicklei/proto"

	"github.com/unitedtraders/proto-filter/internal/config"
	"github.com/unitedtraders/proto-filter/internal/deps"
	"github.com/unitedtraders/proto-filter/internal/parser"
	"github.com/unitedtraders/proto-filter/internal/writer"
)

// T021: Test glob pattern matching
func TestMatchesAny(t *testing.T) {
	tests := []struct {
		fqn      string
		patterns []string
		want     bool
	}{
		{"my.package.OrderService", []string{"my.package.OrderService"}, true},
		{"my.package.OrderService", []string{"my.package.*"}, true},
		{"my.package.sub.Other", []string{"my.package.*"}, false},
		{"my.package.OrderService", []string{"*.OrderService"}, true},
		{"my.package.UserService", []string{"my.package.OrderService"}, false},
		{"filter.OrderService", []string{"filter.*"}, true},
		{"filter.Money", []string{"filter.Money", "filter.Status"}, true},
	}

	for _, tc := range tests {
		t.Run(tc.fqn, func(t *testing.T) {
			got, err := MatchesAny(tc.fqn, tc.patterns)
			if err != nil {
				t.Fatalf("MatchesAny: %v", err)
			}
			if got != tc.want {
				t.Errorf("MatchesAny(%q, %v) = %v, want %v", tc.fqn, tc.patterns, got, tc.want)
			}
		})
	}
}

// T024: Test AST pruning
func TestPruneAST(t *testing.T) {
	dir := testdataDir(t, "simple")
	def, err := parser.ParseProtoFile(filepath.Join(dir, "service.proto"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// Keep only OrderService and CreateOrderRequest
	keepFQNs := map[string]bool{
		"simple.OrderService":        true,
		"simple.CreateOrderRequest":  true,
		"simple.CreateOrderResponse": true,
	}

	PruneAST(def, "simple", keepFQNs)

	// Count remaining definitions
	var services, messages, enums int
	proto.Walk(def,
		proto.WithService(func(s *proto.Service) { services++ }),
		proto.WithMessage(func(m *proto.Message) { messages++ }),
		proto.WithEnum(func(e *proto.Enum) { enums++ }),
	)

	if services != 1 {
		t.Errorf("expected 1 service, got %d", services)
	}
	if messages != 2 {
		t.Errorf("expected 2 messages, got %d", messages)
	}
	if enums != 0 {
		t.Errorf("expected 0 enums (Status was excluded), got %d", enums)
	}
}

// T025: Test conflicting rules detection
func TestApplyFilterConflict(t *testing.T) {
	cfg := &config.FilterConfig{
		Include: []string{"filter.OrderService"},
		Exclude: []string{"filter.OrderService"},
	}
	allFQNs := []string{"filter.OrderService", "filter.UserService"}

	_, err := ApplyFilter(cfg, allFQNs)
	if err == nil {
		t.Fatal("expected error for conflicting rules")
	}
	if !strings.Contains(err.Error(), "conflicting") {
		t.Errorf("error should mention conflicting, got: %v", err)
	}
}

func TestApplyFilterIncludeOnly(t *testing.T) {
	cfg := &config.FilterConfig{
		Include: []string{"filter.OrderService", "filter.Money"},
	}
	allFQNs := []string{
		"filter.OrderService",
		"filter.UserService",
		"filter.Money",
		"filter.Status",
	}

	result, err := ApplyFilter(cfg, allFQNs)
	if err != nil {
		t.Fatalf("ApplyFilter: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 included, got %d", len(result))
	}
	if !result["filter.OrderService"] {
		t.Error("OrderService should be included")
	}
	if !result["filter.Money"] {
		t.Error("Money should be included")
	}
}

func TestApplyFilterExcludeOnly(t *testing.T) {
	cfg := &config.FilterConfig{
		Exclude: []string{"filter.UserService"},
	}
	allFQNs := []string{
		"filter.OrderService",
		"filter.UserService",
		"filter.Money",
	}

	result, err := ApplyFilter(cfg, allFQNs)
	if err != nil {
		t.Fatalf("ApplyFilter: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2, got %d", len(result))
	}
	if result["filter.UserService"] {
		t.Error("UserService should be excluded")
	}
}

func TestApplyFilterPassThrough(t *testing.T) {
	cfg := &config.FilterConfig{}
	allFQNs := []string{"a.B", "a.C"}

	result, err := ApplyFilter(cfg, allFQNs)
	if err != nil {
		t.Fatalf("ApplyFilter: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("pass-through should keep all, got %d", len(result))
	}
}

// T026: Integration test for filtered pipeline
func TestIntegrationFilter(t *testing.T) {
	inputDir := testdataDir(t, "filter")
	outputDir := t.TempDir()

	// Discover and parse all files
	files, err := parser.DiscoverProtoFiles(inputDir)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}

	type parsedFile struct {
		rel string
		def *proto.Proto
		pkg string
	}
	var parsed []parsedFile
	graph := deps.NewGraph()

	for _, rel := range files {
		def, err := parser.ParseProtoFile(filepath.Join(inputDir, rel))
		if err != nil {
			t.Fatalf("parse %s: %v", rel, err)
		}
		pkg := parser.ExtractPackage(def)
		parsed = append(parsed, parsedFile{rel, def, pkg})

		defs := parser.ExtractDefinitions(def, pkg)
		for _, d := range defs {
			graph.AddDefinition(&deps.Definition{
				FQN:        d.FQN,
				Kind:       d.Kind,
				File:       rel,
				References: d.References,
			})
		}
	}

	// Apply filter: include only OrderService
	cfg := &config.FilterConfig{
		Include: []string{"filter.OrderService"},
	}

	allFQNs := make([]string, 0)
	for fqn := range graph.Nodes {
		allFQNs = append(allFQNs, fqn)
	}

	included, err := ApplyFilter(cfg, allFQNs)
	if err != nil {
		t.Fatalf("apply filter: %v", err)
	}

	// Resolve transitive dependencies
	includedList := make([]string, 0, len(included))
	for fqn := range included {
		includedList = append(includedList, fqn)
	}
	allNeeded := graph.TransitiveDeps(includedList)

	keepSet := make(map[string]bool)
	for _, fqn := range allNeeded {
		keepSet[fqn] = true
	}

	// Determine which files are needed
	requiredFiles := graph.RequiredFiles(allNeeded)
	requiredFileSet := make(map[string]bool)
	for _, f := range requiredFiles {
		requiredFileSet[f] = true
	}

	// Prune and write
	for _, pf := range parsed {
		if !requiredFileSet[pf.rel] {
			continue
		}
		PruneAST(pf.def, pf.pkg, keepSet)
		outPath := filepath.Join(outputDir, pf.rel)
		if err := writer.WriteProtoFile(pf.def, outPath); err != nil {
			t.Fatalf("write %s: %v", pf.rel, err)
		}
	}

	// Verify: orders.proto should exist with OrderService
	ordersOut := filepath.Join(outputDir, "orders.proto")
	ordersContent, err := os.ReadFile(ordersOut)
	if err != nil {
		t.Fatalf("read orders output: %v", err)
	}
	ordersStr := string(ordersContent)

	if !strings.Contains(ordersStr, "OrderService") {
		t.Error("output should contain OrderService")
	}
	if strings.Contains(ordersStr, "UserService") {
		t.Error("output should NOT contain UserService")
	}

	// Verify: common.proto should exist (Money and Status are deps)
	commonOut := filepath.Join(outputDir, "common.proto")
	if _, err := os.Stat(commonOut); err != nil {
		t.Error("common.proto should be in output (transitive dep)")
	}

	// Verify: users.proto should NOT exist
	usersOut := filepath.Join(outputDir, "users.proto")
	if _, err := os.Stat(usersOut); err == nil {
		t.Error("users.proto should NOT be in output")
	}

	// Verify OrderService's request/response types are present
	if !strings.Contains(ordersStr, "CreateOrderRequest") {
		t.Error("output should contain CreateOrderRequest (dep of OrderService)")
	}
	if !strings.Contains(ordersStr, "CreateOrderResponse") {
		t.Error("output should contain CreateOrderResponse (dep of OrderService)")
	}

	// Verify excluded messages are not in orders output
	if strings.Contains(ordersStr, "ListUsersRequest") {
		t.Error("output should NOT contain ListUsersRequest")
	}
}

// T006: Test annotation extraction from comments
func TestExtractAnnotations(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  []string
	}{
		{"simple annotation", []string{" @HasAnyRole"}, []string{"HasAnyRole"}},
		{"annotation with args", []string{" @HasAnyRole({\"ADMIN\"})"}, []string{"HasAnyRole"}},
		{"dotted annotation", []string{" @com.example.Secure"}, []string{"com.example.Secure"}},
		{"no annotations", []string{" Creates a new order."}, nil},
		{"mixed lines", []string{" @HasAnyRole({\"ADMIN\"})", " Creates a new order."}, []string{"HasAnyRole"}},
		{"multiple annotations", []string{" @HasAnyRole", " @Deprecated"}, []string{"HasAnyRole", "Deprecated"}},
		{"annotation in block comment", []string{" * @HasAnyRole({\"ADMIN\"})", " * Some description"}, []string{"HasAnyRole"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			comment := &proto.Comment{Lines: tc.lines}
			got := ExtractAnnotations(comment)
			if len(got) != len(tc.want) {
				t.Fatalf("ExtractAnnotations: got %v, want %v", got, tc.want)
			}
			for i, g := range got {
				if g != tc.want[i] {
					t.Errorf("annotation[%d]: got %q, want %q", i, g, tc.want[i])
				}
			}
		})
	}
}

func TestExtractAnnotationsNilComment(t *testing.T) {
	got := ExtractAnnotations(nil)
	if len(got) != 0 {
		t.Errorf("nil comment should return empty, got %v", got)
	}
}

// T007: Test method filtering by annotation
func TestFilterMethodsByAnnotation(t *testing.T) {
	inputPath := filepath.Join(testdataDir(t, "annotations"), "service.proto")
	def, err := parser.ParseProtoFile(inputPath)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	removed := FilterMethodsByAnnotation(def, []string{"HasAnyRole"})
	if removed != 2 {
		t.Errorf("expected 2 methods removed, got %d", removed)
	}

	// Verify remaining methods
	var methodNames []string
	proto.Walk(def, proto.WithRPC(func(r *proto.RPC) {
		methodNames = append(methodNames, r.Name)
	}))

	if len(methodNames) != 1 {
		t.Fatalf("expected 1 remaining method, got %d: %v", len(methodNames), methodNames)
	}
	if methodNames[0] != "ListOrders" {
		t.Errorf("expected ListOrders, got %s", methodNames[0])
	}
}

func TestFilterMethodsByAnnotationNoMatch(t *testing.T) {
	inputPath := filepath.Join(testdataDir(t, "annotations"), "service.proto")
	def, err := parser.ParseProtoFile(inputPath)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	removed := FilterMethodsByAnnotation(def, []string{"NonExistent"})
	if removed != 0 {
		t.Errorf("expected 0 methods removed, got %d", removed)
	}

	var methodCount int
	proto.Walk(def, proto.WithRPC(func(r *proto.RPC) {
		methodCount++
	}))
	if methodCount != 3 {
		t.Errorf("expected 3 methods unchanged, got %d", methodCount)
	}
}

// Test that annotated methods are kept when their annotation is not in the filter list
func TestFilterMethodsByAnnotationKeepsNonFilteredAnnotation(t *testing.T) {
	inputPath := filepath.Join(testdataDir(t, "annotations"), "service.proto")
	def, err := parser.ParseProtoFile(inputPath)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// Filter for "Internal" — none of the methods have this annotation.
	// Methods with @HasAnyRole should NOT be removed.
	removed := FilterMethodsByAnnotation(def, []string{"Internal"})
	if removed != 0 {
		t.Errorf("expected 0 methods removed, got %d", removed)
	}

	var methodNames []string
	proto.Walk(def, proto.WithRPC(func(r *proto.RPC) {
		methodNames = append(methodNames, r.Name)
	}))

	if len(methodNames) != 3 {
		t.Fatalf("expected all 3 methods kept, got %d: %v", len(methodNames), methodNames)
	}

	want := map[string]bool{"CreateOrder": true, "DeleteOrder": true, "ListOrders": true}
	for _, name := range methodNames {
		if !want[name] {
			t.Errorf("unexpected method %q in output", name)
		}
	}
}

// T008: Integration test for annotation method filtering
func TestIntegrationAnnotationFilter(t *testing.T) {
	inputDir := testdataDir(t, "annotations")
	inputPath := filepath.Join(inputDir, "service.proto")
	def, err := parser.ParseProtoFile(inputPath)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// Filter methods by annotation
	FilterMethodsByAnnotation(def, []string{"HasAnyRole"})

	// Write output and verify
	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, "service.proto")
	if err := writer.WriteProtoFile(def, outputPath); err != nil {
		t.Fatalf("write: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	out := string(content)

	// ListOrders should remain
	if !strings.Contains(out, "ListOrders") {
		t.Error("output should contain ListOrders")
	}
	// CreateOrder and DeleteOrder RPC methods should be removed
	if strings.Contains(out, "rpc CreateOrder") {
		t.Error("output should NOT contain rpc CreateOrder (annotated)")
	}
	if strings.Contains(out, "rpc DeleteOrder") {
		t.Error("output should NOT contain rpc DeleteOrder (annotated)")
	}
	// Service should still exist
	if !strings.Contains(out, "OrderService") {
		t.Error("output should contain OrderService (has remaining methods)")
	}

	// Verify output is parseable
	_, err = parser.ParseProtoFile(outputPath)
	if err != nil {
		t.Fatalf("output not parseable: %v", err)
	}
}

// T018: Test removing empty services
func TestRemoveEmptyServices(t *testing.T) {
	inputPath := filepath.Join(testdataDir(t, "annotations"), "internal_only.proto")
	def, err := parser.ParseProtoFile(inputPath)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// Remove all methods (all annotated)
	FilterMethodsByAnnotation(def, []string{"HasAnyRole"})

	removed := RemoveEmptyServices(def)
	if removed != 1 {
		t.Errorf("expected 1 empty service removed, got %d", removed)
	}

	var serviceCount int
	proto.Walk(def, proto.WithService(func(s *proto.Service) { serviceCount++ }))
	if serviceCount != 0 {
		t.Errorf("expected 0 services, got %d", serviceCount)
	}
}

func TestRemoveEmptyServicesKeepsNonEmpty(t *testing.T) {
	inputPath := filepath.Join(testdataDir(t, "annotations"), "service.proto")
	def, err := parser.ParseProtoFile(inputPath)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// Remove annotated methods (2 of 3)
	FilterMethodsByAnnotation(def, []string{"HasAnyRole"})

	removed := RemoveEmptyServices(def)
	if removed != 0 {
		t.Errorf("expected 0 services removed (service still has methods), got %d", removed)
	}

	var serviceCount int
	proto.Walk(def, proto.WithService(func(s *proto.Service) { serviceCount++ }))
	if serviceCount != 1 {
		t.Errorf("expected 1 service, got %d", serviceCount)
	}
}

// T019: Test HasRemainingDefinitions
func TestHasRemainingDefinitions(t *testing.T) {
	// File with definitions
	inputPath := filepath.Join(testdataDir(t, "annotations"), "service.proto")
	def, err := parser.ParseProtoFile(inputPath)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !HasRemainingDefinitions(def) {
		t.Error("file with definitions should return true")
	}

	// Manually remove all definitions
	filtered := make([]proto.Visitee, 0)
	for _, elem := range def.Elements {
		switch elem.(type) {
		case *proto.Service, *proto.Message, *proto.Enum:
			continue
		default:
			filtered = append(filtered, elem)
		}
	}
	def.Elements = filtered
	if HasRemainingDefinitions(def) {
		t.Error("file with no definitions should return false")
	}
}

// T020: Integration test for empty service removal
func TestIntegrationEmptyServiceRemoval(t *testing.T) {
	inputPath := filepath.Join(testdataDir(t, "annotations"), "internal_only.proto")
	def, err := parser.ParseProtoFile(inputPath)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	FilterMethodsByAnnotation(def, []string{"HasAnyRole"})
	RemoveEmptyServices(def)
	RemoveOrphanedDefinitions(def, "annotations")

	if HasRemainingDefinitions(def) {
		t.Error("all definitions should be removed (service was empty, messages orphaned)")
	}
}

// T012: Test collecting referenced types from AST
func TestCollectReferencedTypes(t *testing.T) {
	inputPath := filepath.Join(testdataDir(t, "annotations"), "shared.proto")
	def, err := parser.ParseProtoFile(inputPath)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// Filter out Refund method first
	FilterMethodsByAnnotation(def, []string{"HasAnyRole"})

	refs := CollectReferencedTypes(def, "annotations")

	// GetPaymentStatus's request/response types should be referenced
	if !refs["annotations.PaymentStatusRequest"] {
		t.Error("PaymentStatusRequest should be referenced")
	}
	if !refs["annotations.PaymentStatusResponse"] {
		t.Error("PaymentStatusResponse should be referenced")
	}
	// OrderStatus is referenced by PaymentStatusResponse
	if !refs["annotations.OrderStatus"] {
		t.Error("OrderStatus should be referenced (via PaymentStatusResponse)")
	}
	// Refund types should NOT be referenced (method was removed)
	if refs["annotations.RefundRequest"] {
		t.Error("RefundRequest should NOT be referenced (method removed)")
	}
	if refs["annotations.RefundResponse"] {
		t.Error("RefundResponse should NOT be referenced (method removed)")
	}
}

// T013: Test removing orphaned definitions
func TestRemoveOrphanedDefinitions(t *testing.T) {
	inputPath := filepath.Join(testdataDir(t, "annotations"), "shared.proto")
	def, err := parser.ParseProtoFile(inputPath)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	FilterMethodsByAnnotation(def, []string{"HasAnyRole"})
	removed := RemoveOrphanedDefinitions(def, "annotations")

	if removed != 2 {
		t.Errorf("expected 2 orphaned definitions removed, got %d", removed)
	}

	// Check remaining definitions
	var msgNames []string
	var enumNames []string
	proto.Walk(def,
		proto.WithMessage(func(m *proto.Message) { msgNames = append(msgNames, m.Name) }),
		proto.WithEnum(func(e *proto.Enum) { enumNames = append(enumNames, e.Name) }),
	)

	// PaymentStatusRequest, PaymentStatusResponse should remain
	found := make(map[string]bool)
	for _, n := range msgNames {
		found[n] = true
	}
	if !found["PaymentStatusRequest"] {
		t.Error("PaymentStatusRequest should remain (referenced by kept method)")
	}
	if !found["PaymentStatusResponse"] {
		t.Error("PaymentStatusResponse should remain (referenced by kept method)")
	}
	if found["RefundRequest"] {
		t.Error("RefundRequest should be removed (orphaned)")
	}
	if found["RefundResponse"] {
		t.Error("RefundResponse should be removed (orphaned)")
	}

	// OrderStatus should remain (referenced by PaymentStatusResponse)
	if len(enumNames) != 1 || enumNames[0] != "OrderStatus" {
		t.Errorf("OrderStatus should remain, got enums: %v", enumNames)
	}
}

// T014: Integration test for orphan removal pipeline
func TestIntegrationOrphanRemoval(t *testing.T) {
	inputPath := filepath.Join(testdataDir(t, "annotations"), "shared.proto")
	def, err := parser.ParseProtoFile(inputPath)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	FilterMethodsByAnnotation(def, []string{"HasAnyRole"})
	RemoveOrphanedDefinitions(def, "annotations")

	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, "shared.proto")
	if err := writer.WriteProtoFile(def, outputPath); err != nil {
		t.Fatalf("write: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	out := string(content)

	if !strings.Contains(out, "PaymentStatusRequest") {
		t.Error("output should contain PaymentStatusRequest")
	}
	if !strings.Contains(out, "OrderStatus") {
		t.Error("output should contain OrderStatus (referenced by kept message)")
	}
	if strings.Contains(out, "RefundRequest") {
		t.Error("output should NOT contain RefundRequest (orphaned)")
	}
	if strings.Contains(out, "RefundResponse") {
		t.Error("output should NOT contain RefundResponse (orphaned)")
	}

	// Verify parseable
	_, err = parser.ParseProtoFile(outputPath)
	if err != nil {
		t.Fatalf("output not parseable: %v", err)
	}
}

// T009: Golden file comparison helper
func testGoldenFile(t *testing.T, inputName string) {
	t.Helper()
	inputPath := filepath.Join(testdataDir(t, "comments"), inputName+".proto")
	goldenPath := filepath.Join(testdataDir(t, "comments"), "expected", inputName+".proto")

	def, err := parser.ParseProtoFile(inputPath)
	if err != nil {
		t.Fatalf("parse %s: %v", inputName, err)
	}

	ConvertBlockComments(def)

	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, inputName+".proto")
	if err := writer.WriteProtoFile(def, outputPath); err != nil {
		t.Fatalf("write: %v", err)
	}

	actual, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}

	if string(actual) != string(expected) {
		t.Errorf("output does not match golden file %s\n--- ACTUAL ---\n%s\n--- EXPECTED ---\n%s", goldenPath, string(actual), string(expected))
	}
}

// T010: Golden file test for multiline.proto
func TestGoldenFileMultiline(t *testing.T) {
	testGoldenFile(t, "multiline")
}

// T011: Golden file test for commented.proto (unchanged)
func TestGoldenFileCommented(t *testing.T) {
	testGoldenFile(t, "commented")
}

// T012: Golden file test for block_comments.proto
func TestGoldenFileBlockComments(t *testing.T) {
	testGoldenFile(t, "block_comments")
}

// T013: Test annotation preservation during conversion
func TestConvertBlockCommentsPreservesAnnotations(t *testing.T) {
	inputPath := filepath.Join(testdataDir(t, "comments"), "block_comments.proto")
	def, err := parser.ParseProtoFile(inputPath)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	ConvertBlockComments(def)

	// Find the GetPriceUpdates RPC and verify annotations are in Lines
	var found bool
	proto.Walk(def, proto.WithRPC(func(r *proto.RPC) {
		if r.Name == "GetPriceUpdates" && r.Comment != nil {
			found = true
			lines := strings.Join(r.Comment.Lines, "\n")
			if !strings.Contains(lines, "@StartsWithSnapshot") {
				t.Error("@StartsWithSnapshot annotation should be preserved")
			}
			if !strings.Contains(lines, "@SupportWindow") {
				t.Error("@SupportWindow annotation should be preserved")
			}
		}
	}))
	if !found {
		t.Error("GetPriceUpdates RPC not found")
	}
}

// T014: Test empty block comment conversion
func TestConvertBlockCommentsEmptyComment(t *testing.T) {
	comment := &proto.Comment{
		Cstyle: true,
		Lines:  []string{" "},
	}
	convertComment(comment)

	if comment.Cstyle {
		t.Error("comment should no longer be Cstyle")
	}
}

// T005: Test block comment conversion on AST level
func TestConvertBlockComments(t *testing.T) {
	inputPath := filepath.Join(testdataDir(t, "comments"), "multiline.proto")
	def, err := parser.ParseProtoFile(inputPath)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// Verify there are Cstyle comments before conversion
	hasCstyle := false
	proto.Walk(def,
		proto.WithService(func(s *proto.Service) {
			if s.Comment != nil && s.Comment.Cstyle {
				hasCstyle = true
			}
		}),
		proto.WithMessage(func(m *proto.Message) {
			if m.Comment != nil && m.Comment.Cstyle {
				hasCstyle = true
			}
		}),
		proto.WithEnum(func(e *proto.Enum) {
			if e.Comment != nil && e.Comment.Cstyle {
				hasCstyle = true
			}
		}),
	)
	if !hasCstyle {
		t.Fatal("multiline.proto should have Cstyle comments before conversion")
	}

	ConvertBlockComments(def)

	// Verify no Cstyle comments remain
	proto.Walk(def,
		proto.WithService(func(s *proto.Service) {
			if s.Comment != nil && s.Comment.Cstyle {
				t.Errorf("service %s still has Cstyle comment", s.Name)
			}
		}),
		proto.WithMessage(func(m *proto.Message) {
			if m.Comment != nil && m.Comment.Cstyle {
				t.Errorf("message %s still has Cstyle comment", m.Name)
			}
		}),
		proto.WithEnum(func(e *proto.Enum) {
			if e.Comment != nil && e.Comment.Cstyle {
				t.Errorf("enum %s still has Cstyle comment", e.Name)
			}
		}),
		proto.WithRPC(func(r *proto.RPC) {
			if r.Comment != nil && r.Comment.Cstyle {
				t.Errorf("rpc %s still has Cstyle comment", r.Name)
			}
		}),
	)
}

// T006: Test that single-line comments are preserved unchanged
func TestConvertBlockCommentsPreservesExisting(t *testing.T) {
	inputPath := filepath.Join(testdataDir(t, "comments"), "commented.proto")
	def, err := parser.ParseProtoFile(inputPath)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// Count comments before
	commentCountBefore := 0
	proto.Walk(def,
		proto.WithService(func(s *proto.Service) {
			if s.Comment != nil {
				commentCountBefore++
			}
		}),
		proto.WithMessage(func(m *proto.Message) {
			if m.Comment != nil {
				commentCountBefore++
			}
		}),
		proto.WithEnum(func(e *proto.Enum) {
			if e.Comment != nil {
				commentCountBefore++
			}
		}),
		proto.WithRPC(func(r *proto.RPC) {
			if r.Comment != nil {
				commentCountBefore++
			}
		}),
	)

	ConvertBlockComments(def)

	// Count comments after
	commentCountAfter := 0
	proto.Walk(def,
		proto.WithService(func(s *proto.Service) {
			if s.Comment != nil {
				commentCountAfter++
			}
		}),
		proto.WithMessage(func(m *proto.Message) {
			if m.Comment != nil {
				commentCountAfter++
			}
		}),
		proto.WithEnum(func(e *proto.Enum) {
			if e.Comment != nil {
				commentCountAfter++
			}
		}),
		proto.WithRPC(func(r *proto.RPC) {
			if r.Comment != nil {
				commentCountAfter++
			}
		}),
	)

	if commentCountBefore != commentCountAfter {
		t.Errorf("comment count changed: before=%d, after=%d", commentCountBefore, commentCountAfter)
	}
	if commentCountBefore == 0 {
		t.Error("commented.proto should have comments")
	}
}

func testdataDir(t *testing.T, sub string) string {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("..", "..", "testdata", sub))
	if err != nil {
		t.Fatalf("testdata path: %v", err)
	}
	return dir
}
