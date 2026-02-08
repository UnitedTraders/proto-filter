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

func testdataDir(t *testing.T, sub string) string {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("..", "..", "testdata", sub))
	if err != nil {
		t.Fatalf("testdata path: %v", err)
	}
	return dir
}
