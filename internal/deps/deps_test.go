package deps

import (
	"sort"
	"testing"
)

// T022: Test dependency graph construction
func TestGraphConstruction(t *testing.T) {
	g := NewGraph()

	g.AddDefinition(&Definition{
		FQN:        "pkg.OrderService",
		Kind:       "service",
		File:       "orders.proto",
		References: []string{"pkg.CreateOrderRequest", "pkg.CreateOrderResponse"},
	})
	g.AddDefinition(&Definition{
		FQN:        "pkg.CreateOrderRequest",
		Kind:       "message",
		File:       "orders.proto",
		References: []string{"pkg.Money"},
	})
	g.AddDefinition(&Definition{
		FQN:        "pkg.CreateOrderResponse",
		Kind:       "message",
		File:       "orders.proto",
		References: []string{"pkg.Order"},
	})
	g.AddDefinition(&Definition{
		FQN:        "pkg.Order",
		Kind:       "message",
		File:       "orders.proto",
		References: []string{"pkg.Money", "pkg.Status"},
	})
	g.AddDefinition(&Definition{
		FQN:  "pkg.Money",
		Kind: "message",
		File: "common.proto",
	})
	g.AddDefinition(&Definition{
		FQN:  "pkg.Status",
		Kind: "enum",
		File: "common.proto",
	})

	if len(g.Nodes) != 6 {
		t.Errorf("expected 6 nodes, got %d", len(g.Nodes))
	}

	edges := g.Edges["pkg.OrderService"]
	if len(edges) != 2 {
		t.Errorf("OrderService should have 2 edges, got %d", len(edges))
	}
}

// T023: Test transitive dependency resolution
func TestTransitiveDeps(t *testing.T) {
	g := NewGraph()

	g.AddDefinition(&Definition{
		FQN:        "pkg.OrderService",
		Kind:       "service",
		File:       "orders.proto",
		References: []string{"pkg.CreateOrderRequest", "pkg.CreateOrderResponse"},
	})
	g.AddDefinition(&Definition{
		FQN:        "pkg.CreateOrderRequest",
		Kind:       "message",
		File:       "orders.proto",
		References: []string{"pkg.Money"},
	})
	g.AddDefinition(&Definition{
		FQN:        "pkg.CreateOrderResponse",
		Kind:       "message",
		File:       "orders.proto",
		References: []string{"pkg.Order"},
	})
	g.AddDefinition(&Definition{
		FQN:        "pkg.Order",
		Kind:       "message",
		File:       "orders.proto",
		References: []string{"pkg.Money", "pkg.Status"},
	})
	g.AddDefinition(&Definition{
		FQN:  "pkg.Money",
		Kind: "message",
		File: "common.proto",
	})
	g.AddDefinition(&Definition{
		FQN:  "pkg.Status",
		Kind: "enum",
		File: "common.proto",
	})

	result := g.TransitiveDeps([]string{"pkg.OrderService"})
	sort.Strings(result)

	expected := []string{
		"pkg.CreateOrderRequest",
		"pkg.CreateOrderResponse",
		"pkg.Money",
		"pkg.Order",
		"pkg.OrderService",
		"pkg.Status",
	}

	if len(result) != len(expected) {
		t.Fatalf("expected %d transitive deps, got %d: %v", len(expected), len(result), result)
	}
	for i, fqn := range result {
		if fqn != expected[i] {
			t.Errorf("dep[%d]: expected %s, got %s", i, expected[i], fqn)
		}
	}
}

func TestTransitiveDepsLeafNode(t *testing.T) {
	g := NewGraph()
	g.AddDefinition(&Definition{
		FQN:  "pkg.Status",
		Kind: "enum",
		File: "common.proto",
	})

	result := g.TransitiveDeps([]string{"pkg.Status"})
	if len(result) != 1 || result[0] != "pkg.Status" {
		t.Errorf("leaf node transitive deps should be just itself, got: %v", result)
	}
}

func TestRequiredFiles(t *testing.T) {
	g := NewGraph()
	g.AddDefinition(&Definition{FQN: "pkg.A", File: "a.proto"})
	g.AddDefinition(&Definition{FQN: "pkg.B", File: "b.proto"})
	g.AddDefinition(&Definition{FQN: "pkg.C", File: "a.proto"})

	files := g.RequiredFiles([]string{"pkg.A", "pkg.B", "pkg.C"})
	sort.Strings(files)

	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(files), files)
	}
	if files[0] != "a.proto" || files[1] != "b.proto" {
		t.Errorf("expected [a.proto, b.proto], got %v", files)
	}
}
