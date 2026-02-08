package filter

import (
	"fmt"
	"path"
	"strings"

	"github.com/emicklei/proto"

	"github.com/unitedtraders/proto-filter/internal/config"
)

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

func qualifiedName(pkg, name string) string {
	if pkg == "" {
		return name
	}
	return pkg + "." + name
}
