package filter

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/emicklei/proto"

	"github.com/unitedtraders/proto-filter/internal/config"
)

var annotationRegex = regexp.MustCompile(`@(\w[\w.]*)`)

// MatchesAny returns true if fqn matches any of the glob patterns.
// Dots in FQNs are treated as path separators so that `*` matches
// a single package segment (e.g., `my.package.*` matches
// `my.package.Foo` but not `my.package.sub.Bar`).
// A leading or trailing `*` segment also attempts suffix/prefix
// matching (e.g., `*.OrderService` matches `my.package.OrderService`).
func MatchesAny(fqn string, patterns []string) (bool, error) {
	for _, p := range patterns {
		matched, err := matchGlob(fqn, p)
		if err != nil {
			return false, fmt.Errorf("invalid glob pattern %q: %w", p, err)
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

func matchGlob(fqn, pattern string) (bool, error) {
	fqnPath := strings.ReplaceAll(fqn, ".", "/")
	pPath := strings.ReplaceAll(pattern, ".", "/")

	// Try exact match first
	matched, err := path.Match(pPath, fqnPath)
	if err != nil {
		return false, err
	}
	if matched {
		return true, nil
	}

	// If pattern starts with "*/" (e.g., *.OrderService → */OrderService),
	// try suffix matching: check if the FQN ends with the non-star part
	if strings.HasPrefix(pPath, "*/") {
		suffix := pPath[2:] // e.g., "OrderService"
		if strings.HasSuffix(fqnPath, "/"+suffix) || fqnPath == suffix {
			return true, nil
		}
	}

	return false, nil
}

// ApplyFilter takes a FilterConfig and a list of all FQNs, applies
// include/exclude rules, and returns the set of FQNs to keep.
// Returns an error if conflicting rules are detected.
func ApplyFilter(cfg *config.FilterConfig, allFQNs []string) (map[string]bool, error) {
	result := make(map[string]bool)

	if cfg.IsPassThrough() {
		for _, fqn := range allFQNs {
			result[fqn] = true
		}
		return result, nil
	}

	// Apply include rules: if include is non-empty, only keep matching
	if len(cfg.Include) > 0 {
		for _, fqn := range allFQNs {
			matched, err := MatchesAny(fqn, cfg.Include)
			if err != nil {
				return nil, err
			}
			if matched {
				result[fqn] = true
			}
		}
	} else {
		// No include rules: start with all
		for _, fqn := range allFQNs {
			result[fqn] = true
		}
	}

	// Apply exclude rules: remove matching from result
	if len(cfg.Exclude) > 0 {
		for fqn := range result {
			matched, err := MatchesAny(fqn, cfg.Exclude)
			if err != nil {
				return nil, err
			}
			if matched {
				// Check for conflict: included explicitly AND excluded
				if len(cfg.Include) > 0 {
					includedExplicitly, _ := MatchesAny(fqn, cfg.Include)
					if includedExplicitly {
						return nil, fmt.Errorf("conflicting rules: %q matches both include and exclude patterns", fqn)
					}
				}
				delete(result, fqn)
			}
		}
	}

	return result, nil
}

// PruneAST removes top-level elements from a parsed proto AST that
// are not in the keepFQNs set. Preserves syntax, package, options,
// and import statements.
func PruneAST(def *proto.Proto, pkg string, keepFQNs map[string]bool) {
	filtered := make([]proto.Visitee, 0, len(def.Elements))
	for _, elem := range def.Elements {
		switch v := elem.(type) {
		case *proto.Service:
			fqn := qualifiedName(pkg, v.Name)
			if keepFQNs[fqn] {
				filtered = append(filtered, elem)
			}
		case *proto.Message:
			fqn := qualifiedName(pkg, v.Name)
			if keepFQNs[fqn] {
				filtered = append(filtered, elem)
			}
		case *proto.Enum:
			fqn := qualifiedName(pkg, v.Name)
			if keepFQNs[fqn] {
				filtered = append(filtered, elem)
			}
		default:
			// Keep syntax, package, imports, options, comments
			filtered = append(filtered, elem)
		}
	}
	def.Elements = filtered
}

// ExtractAnnotations returns annotation names found in a proto comment.
// Annotations follow the pattern @Name or @Name(...).
// Returns nil if comment is nil or contains no annotations.
func ExtractAnnotations(comment *proto.Comment) []string {
	if comment == nil {
		return nil
	}
	var annotations []string
	for _, line := range comment.Lines {
		matches := annotationRegex.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			annotations = append(annotations, m[1])
		}
	}
	return annotations
}

// FilterMethodsByAnnotation removes RPC methods from services in the
// given proto AST whose comments contain any of the specified annotations.
// Returns the number of methods removed.
func FilterMethodsByAnnotation(def *proto.Proto, annotations []string) int {
	if len(annotations) == 0 {
		return 0
	}
	annotSet := make(map[string]bool, len(annotations))
	for _, a := range annotations {
		annotSet[a] = true
	}

	removed := 0
	for _, elem := range def.Elements {
		svc, ok := elem.(*proto.Service)
		if !ok {
			continue
		}
		filtered := make([]proto.Visitee, 0, len(svc.Elements))
		for _, svcElem := range svc.Elements {
			rpc, ok := svcElem.(*proto.RPC)
			if !ok {
				filtered = append(filtered, svcElem)
				continue
			}
			annots := ExtractAnnotations(rpc.Comment)
			shouldRemove := false
			for _, a := range annots {
				if annotSet[a] {
					shouldRemove = true
					break
				}
			}
			if shouldRemove {
				removed++
			} else {
				filtered = append(filtered, svcElem)
			}
		}
		svc.Elements = filtered
	}
	return removed
}

// RemoveEmptyServices removes service definitions that have zero RPC
// method children. Returns the count of removed services.
func RemoveEmptyServices(def *proto.Proto) int {
	filtered := make([]proto.Visitee, 0, len(def.Elements))
	removed := 0
	for _, elem := range def.Elements {
		if svc, ok := elem.(*proto.Service); ok {
			hasRPC := false
			for _, svcElem := range svc.Elements {
				if _, ok := svcElem.(*proto.RPC); ok {
					hasRPC = true
					break
				}
			}
			if !hasRPC {
				removed++
				continue
			}
		}
		filtered = append(filtered, elem)
	}
	def.Elements = filtered
	return removed
}

// HasRemainingDefinitions returns true if the proto AST contains at
// least one service, message, or enum definition.
func HasRemainingDefinitions(def *proto.Proto) bool {
	for _, elem := range def.Elements {
		switch elem.(type) {
		case *proto.Service, *proto.Message, *proto.Enum:
			return true
		}
	}
	return false
}

// CollectReferencedTypes walks the AST and collects all FQNs referenced
// by remaining RPC methods (request/response types) and message fields.
func CollectReferencedTypes(def *proto.Proto, pkg string) map[string]bool {
	refs := make(map[string]bool)

	for _, elem := range def.Elements {
		switch v := elem.(type) {
		case *proto.Service:
			for _, svcElem := range v.Elements {
				if rpc, ok := svcElem.(*proto.RPC); ok {
					addRef(refs, pkg, rpc.RequestType)
					addRef(refs, pkg, rpc.ReturnsType)
				}
			}
		case *proto.Message:
			collectMessageRefs(refs, pkg, v)
		}
	}
	return refs
}

func collectMessageRefs(refs map[string]bool, pkg string, m *proto.Message) {
	for _, elem := range m.Elements {
		switch f := elem.(type) {
		case *proto.NormalField:
			if isUserType(f.Type) {
				addRef(refs, pkg, f.Type)
			}
		case *proto.MapField:
			if isUserType(f.Type) {
				addRef(refs, pkg, f.Type)
			}
		case *proto.OneOfField:
			if isUserType(f.Type) {
				addRef(refs, pkg, f.Type)
			}
		}
	}
}

func addRef(refs map[string]bool, pkg, typeName string) {
	if typeName == "" {
		return
	}
	if strings.Contains(typeName, ".") {
		refs[typeName] = true
	} else {
		refs[pkg+"."+typeName] = true
	}
}

func isUserType(typeName string) bool {
	switch typeName {
	case "double", "float", "int32", "int64", "uint32", "uint64",
		"sint32", "sint64", "fixed32", "fixed64", "sfixed32", "sfixed64",
		"bool", "string", "bytes":
		return false
	}
	return true
}

// RemoveOrphanedDefinitions iteratively removes messages and enums
// that are no longer referenced by any remaining RPC method or message.
// Returns the total count of removed definitions.
func RemoveOrphanedDefinitions(def *proto.Proto, pkg string) int {
	totalRemoved := 0
	for {
		refs := CollectReferencedTypes(def, pkg)

		filtered := make([]proto.Visitee, 0, len(def.Elements))
		removed := 0
		for _, elem := range def.Elements {
			switch v := elem.(type) {
			case *proto.Message:
				fqn := qualifiedName(pkg, v.Name)
				if refs[fqn] {
					filtered = append(filtered, elem)
				} else {
					removed++
				}
			case *proto.Enum:
				fqn := qualifiedName(pkg, v.Name)
				if refs[fqn] {
					filtered = append(filtered, elem)
				} else {
					removed++
				}
			default:
				filtered = append(filtered, elem)
			}
		}
		def.Elements = filtered
		totalRemoved += removed
		if removed == 0 {
			break
		}
	}
	return totalRemoved
}

func qualifiedName(pkg, name string) string {
	if pkg == "" {
		return name
	}
	return pkg + "." + name
}
