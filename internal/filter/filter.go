package filter

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/emicklei/proto"

	"github.com/unitedtraders/proto-filter/internal/config"
)

var annotationRegex = regexp.MustCompile(`@(\w[\w.]*)|\[(\w[\w.]*)(?:\([^)]*\))?\]`)
var substitutionRegex = regexp.MustCompile(`@(\w[\w.]*)(?:\(([^)]*)\))?|\[(\w[\w.]*)(?:\(([^)]*)\))?\]`)

// AnnotationLocation represents a single annotation occurrence found in a
// proto source file, with its file path and line number.
type AnnotationLocation struct {
	File  string // relative file path
	Line  int    // 1-based line number in source
	Name  string // annotation name (for substitution map lookup)
	Token string // full annotation token as it appears in source
}

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
// Annotations follow the pattern @Name, @Name(...), [Name], or [Name(...)].
// Returns nil if comment is nil or contains no annotations.
func ExtractAnnotations(comment *proto.Comment) []string {
	if comment == nil {
		return nil
	}
	var annotations []string
	for _, line := range comment.Lines {
		matches := annotationRegex.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			if m[1] != "" {
				annotations = append(annotations, m[1])
			} else if m[2] != "" {
				annotations = append(annotations, m[2])
			}
		}
	}
	return annotations
}

// FilterServicesByAnnotation removes entire services from the proto AST
// whose comments contain any of the specified annotations. Returns the
// number of services removed.
func FilterServicesByAnnotation(def *proto.Proto, annotations []string) int {
	if len(annotations) == 0 {
		return 0
	}
	annotSet := make(map[string]bool, len(annotations))
	for _, a := range annotations {
		annotSet[a] = true
	}

	filtered := make([]proto.Visitee, 0, len(def.Elements))
	removed := 0
	for _, elem := range def.Elements {
		svc, ok := elem.(*proto.Service)
		if !ok {
			filtered = append(filtered, elem)
			continue
		}
		annots := ExtractAnnotations(svc.Comment)
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
			filtered = append(filtered, elem)
		}
	}
	def.Elements = filtered
	return removed
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

// IncludeServicesByAnnotation removes services from the proto AST
// whose comments contain annotations but NONE of them match the
// specified include list. Services without any annotations are kept
// (their methods will be filtered individually by IncludeMethodsByAnnotation).
// Returns the number of services removed.
func IncludeServicesByAnnotation(def *proto.Proto, annotations []string) int {
	if len(annotations) == 0 {
		return 0
	}
	annotSet := make(map[string]bool, len(annotations))
	for _, a := range annotations {
		annotSet[a] = true
	}

	filtered := make([]proto.Visitee, 0, len(def.Elements))
	removed := 0
	for _, elem := range def.Elements {
		svc, ok := elem.(*proto.Service)
		if !ok {
			filtered = append(filtered, elem)
			continue
		}
		annots := ExtractAnnotations(svc.Comment)
		hasMatch := false
		for _, a := range annots {
			if annotSet[a] {
				hasMatch = true
				break
			}
		}
		if hasMatch {
			filtered = append(filtered, elem)
		} else {
			removed++
		}
	}
	def.Elements = filtered
	return removed
}

// IncludeMethodsByAnnotation removes RPC methods from services in the
// given proto AST whose comments do NOT contain any of the specified
// annotations. This is the inverse of FilterMethodsByAnnotation.
// Returns the number of methods removed.
func IncludeMethodsByAnnotation(def *proto.Proto, annotations []string) int {
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
			hasMatch := false
			for _, a := range annots {
				if annotSet[a] {
					hasMatch = true
					break
				}
			}
			if hasMatch {
				filtered = append(filtered, svcElem)
			} else {
				removed++
			}
		}
		svc.Elements = filtered
	}
	return removed
}

// FilterFieldsByAnnotation removes individual message fields from the proto
// AST whose comments contain any of the specified annotations. Handles
// NormalField, MapField, and OneOfField (within Oneof containers). Also
// recurses into nested messages. Returns the total count of fields removed.
func FilterFieldsByAnnotation(def *proto.Proto, annotations []string) int {
	if len(annotations) == 0 {
		return 0
	}
	annotSet := make(map[string]bool, len(annotations))
	for _, a := range annotations {
		annotSet[a] = true
	}

	removed := 0
	for _, elem := range def.Elements {
		msg, ok := elem.(*proto.Message)
		if !ok {
			continue
		}
		removed += filterFieldsInMessage(msg, annotSet)
	}
	return removed
}

func filterFieldsInMessage(msg *proto.Message, annotSet map[string]bool) int {
	removed := 0
	filtered := make([]proto.Visitee, 0, len(msg.Elements))
	for _, elem := range msg.Elements {
		switch f := elem.(type) {
		case *proto.NormalField:
			if fieldHasAnnotation(f.Comment, f.InlineComment, annotSet) {
				removed++
				continue
			}
		case *proto.MapField:
			if fieldHasAnnotation(f.Comment, f.InlineComment, annotSet) {
				removed++
				continue
			}
		case *proto.Oneof:
			removed += filterFieldsInOneof(f, annotSet)
		case *proto.Message:
			removed += filterFieldsInMessage(f, annotSet)
		}
		filtered = append(filtered, elem)
	}
	msg.Elements = filtered
	return removed
}

func filterFieldsInOneof(oneof *proto.Oneof, annotSet map[string]bool) int {
	removed := 0
	filtered := make([]proto.Visitee, 0, len(oneof.Elements))
	for _, elem := range oneof.Elements {
		if f, ok := elem.(*proto.OneOfField); ok {
			if fieldHasAnnotation(f.Comment, f.InlineComment, annotSet) {
				removed++
				continue
			}
		}
		filtered = append(filtered, elem)
	}
	oneof.Elements = filtered
	return removed
}

func fieldHasAnnotation(comment, inlineComment *proto.Comment, annotSet map[string]bool) bool {
	for _, a := range ExtractAnnotations(comment) {
		if annotSet[a] {
			return true
		}
	}
	for _, a := range ExtractAnnotations(inlineComment) {
		if annotSet[a] {
			return true
		}
	}
	return false
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

// ConvertBlockComments walks the proto AST and converts all C-style
// block comments (/* ... */) to single-line // comments. Leading
// asterisk prefixes are stripped from each line.
func ConvertBlockComments(def *proto.Proto) {
	for _, elem := range def.Elements {
		switch v := elem.(type) {
		case *proto.Service:
			convertComment(v.Comment)
			for _, svcElem := range v.Elements {
				if rpc, ok := svcElem.(*proto.RPC); ok {
					convertComment(rpc.Comment)
					convertComment(rpc.InlineComment)
				}
			}
		case *proto.Message:
			convertComment(v.Comment)
			for _, mElem := range v.Elements {
				switch f := mElem.(type) {
				case *proto.NormalField:
					convertComment(f.Comment)
					convertComment(f.InlineComment)
				case *proto.MapField:
					convertComment(f.Comment)
					convertComment(f.InlineComment)
				case *proto.OneOfField:
					convertComment(f.Comment)
					convertComment(f.InlineComment)
				}
			}
		case *proto.Enum:
			convertComment(v.Comment)
			for _, eElem := range v.Elements {
				if ef, ok := eElem.(*proto.EnumField); ok {
					convertComment(ef.Comment)
					convertComment(ef.InlineComment)
				}
			}
		}
	}
}

// convertComment converts a single block comment to single-line style.
// It sets Cstyle to false and strips leading asterisk prefixes from lines.
// Nil or non-Cstyle comments are left unchanged.
func convertComment(c *proto.Comment) {
	if c == nil || !c.Cstyle {
		return
	}
	c.Cstyle = false
	cleaned := make([]string, 0, len(c.Lines))
	for _, line := range c.Lines {
		cleaned = append(cleaned, " "+cleanBlockCommentLine(line))
	}
	// Trim leading/trailing empty lines from block comment framing
	for len(cleaned) > 0 && strings.TrimSpace(cleaned[0]) == "" {
		cleaned = cleaned[1:]
	}
	for len(cleaned) > 0 && strings.TrimSpace(cleaned[len(cleaned)-1]) == "" {
		cleaned = cleaned[:len(cleaned)-1]
	}
	c.Lines = cleaned
}

// cleanBlockCommentLine strips block comment formatting from a single line.
func cleanBlockCommentLine(line string) string {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "* ") {
		return trimmed[2:]
	}
	if trimmed == "*" {
		return ""
	}
	if strings.HasPrefix(trimmed, "*") {
		return trimmed[1:]
	}
	return trimmed
}

// SubstituteAnnotations replaces annotation tokens in comments across the
// proto AST using the provided substitutions map. For each annotation found,
// if its name exists in the map, the full annotation token is replaced with
// the mapped description text. Empty description values cause the annotation
// token to be removed; if all content is removed from a comment line, the
// line is dropped; if all lines are dropped, the comment is set to nil on the
// element. Returns the total count of substitutions made.
// StripAnnotations removes annotation markers from comments by substituting
// each annotation name with an empty string. It reuses SubstituteAnnotations
// internally, which handles removing empty comment lines and nil-ing comments.
func StripAnnotations(def *proto.Proto, annotations []string) int {
	stripMap := make(map[string]string, len(annotations))
	for _, name := range annotations {
		stripMap[name] = ""
	}
	return SubstituteAnnotations(def, stripMap)
}

func SubstituteAnnotations(def *proto.Proto, substitutions map[string]string) int {
	if len(substitutions) == 0 {
		return 0
	}
	count := 0
	for _, elem := range def.Elements {
		switch v := elem.(type) {
		case *proto.Service:
			count += substituteInComment(&v.Comment, substitutions)
			for _, svcElem := range v.Elements {
				if rpc, ok := svcElem.(*proto.RPC); ok {
					count += substituteInComment(&rpc.Comment, substitutions)
					count += substituteInComment(&rpc.InlineComment, substitutions)
				}
			}
		case *proto.Message:
			count += substituteInComment(&v.Comment, substitutions)
			for _, mElem := range v.Elements {
				switch f := mElem.(type) {
				case *proto.NormalField:
					count += substituteInComment(&f.Comment, substitutions)
					count += substituteInComment(&f.InlineComment, substitutions)
				case *proto.MapField:
					count += substituteInComment(&f.Comment, substitutions)
					count += substituteInComment(&f.InlineComment, substitutions)
				case *proto.OneOfField:
					count += substituteInComment(&f.Comment, substitutions)
					count += substituteInComment(&f.InlineComment, substitutions)
				}
			}
		case *proto.Enum:
			count += substituteInComment(&v.Comment, substitutions)
			for _, eElem := range v.Elements {
				if ef, ok := eElem.(*proto.EnumField); ok {
					count += substituteInComment(&ef.Comment, substitutions)
					count += substituteInComment(&ef.InlineComment, substitutions)
				}
			}
		}
	}
	return count
}

// substituteInComment performs annotation substitution on a single comment.
// Accepts a pointer-to-pointer so the comment can be set to nil if all lines
// are removed.
func substituteInComment(cp **proto.Comment, substitutions map[string]string) int {
	if cp == nil || *cp == nil {
		return 0
	}
	c := *cp
	count := 0
	var cleaned []string
	for _, line := range c.Lines {
		newLine := substitutionRegex.ReplaceAllStringFunc(line, func(match string) string {
			// Parse the match to extract annotation name and argument
			submatch := substitutionRegex.FindStringSubmatch(match)
			name := submatch[1]
			args := submatch[2]
			if name == "" {
				name = submatch[3]
				args = submatch[4]
			}
			if replacement, ok := substitutions[name]; ok {
				count++
				if strings.Contains(replacement, "%s") {
					return strings.Replace(replacement, "%s", args, 1)
				}
				return replacement
			}
			return match
		})
		// Trim the line
		trimmed := strings.TrimSpace(newLine)
		if trimmed == "" {
			// Line became empty after substitution — drop it
			continue
		}
		cleaned = append(cleaned, " "+trimmed)
	}
	if len(cleaned) == 0 {
		*cp = nil
	} else {
		c.Lines = cleaned
	}
	return count
}

// CollectAnnotationLocations walks all elements in the proto AST and collects
// the location of each annotation occurrence. Returns a slice of AnnotationLocation
// with the file path, line number, annotation name, and full token.
func CollectAnnotationLocations(def *proto.Proto, relPath string) []AnnotationLocation {
	var locations []AnnotationLocation
	for _, elem := range def.Elements {
		switch v := elem.(type) {
		case *proto.Service:
			locations = collectLocationsFromComment(v.Comment, relPath, locations)
			for _, svcElem := range v.Elements {
				if rpc, ok := svcElem.(*proto.RPC); ok {
					locations = collectLocationsFromComment(rpc.Comment, relPath, locations)
					locations = collectLocationsFromComment(rpc.InlineComment, relPath, locations)
				}
			}
		case *proto.Message:
			locations = collectLocationsFromComment(v.Comment, relPath, locations)
			for _, mElem := range v.Elements {
				switch f := mElem.(type) {
				case *proto.NormalField:
					locations = collectLocationsFromComment(f.Comment, relPath, locations)
					locations = collectLocationsFromComment(f.InlineComment, relPath, locations)
				case *proto.MapField:
					locations = collectLocationsFromComment(f.Comment, relPath, locations)
					locations = collectLocationsFromComment(f.InlineComment, relPath, locations)
				case *proto.OneOfField:
					locations = collectLocationsFromComment(f.Comment, relPath, locations)
					locations = collectLocationsFromComment(f.InlineComment, relPath, locations)
				}
			}
		case *proto.Enum:
			locations = collectLocationsFromComment(v.Comment, relPath, locations)
			for _, eElem := range v.Elements {
				if ef, ok := eElem.(*proto.EnumField); ok {
					locations = collectLocationsFromComment(ef.Comment, relPath, locations)
					locations = collectLocationsFromComment(ef.InlineComment, relPath, locations)
				}
			}
		}
	}
	return locations
}

func collectLocationsFromComment(c *proto.Comment, relPath string, locations []AnnotationLocation) []AnnotationLocation {
	if c == nil {
		return locations
	}
	for i, line := range c.Lines {
		matches := substitutionRegex.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			name := m[1]
			if name == "" {
				name = m[3]
			}
			locations = append(locations, AnnotationLocation{
				File:  relPath,
				Line:  c.Position.Line + i,
				Name:  name,
				Token: m[0],
			})
		}
	}
	return locations
}

// CollectAllAnnotations walks all elements in the proto AST and collects all
// unique annotation names from comments. Returns a map of annotation names.
func CollectAllAnnotations(def *proto.Proto) map[string]bool {
	result := make(map[string]bool)
	for _, elem := range def.Elements {
		switch v := elem.(type) {
		case *proto.Service:
			collectAnnotationsFromComment(v.Comment, result)
			for _, svcElem := range v.Elements {
				if rpc, ok := svcElem.(*proto.RPC); ok {
					collectAnnotationsFromComment(rpc.Comment, result)
					collectAnnotationsFromComment(rpc.InlineComment, result)
				}
			}
		case *proto.Message:
			collectAnnotationsFromComment(v.Comment, result)
			for _, mElem := range v.Elements {
				switch f := mElem.(type) {
				case *proto.NormalField:
					collectAnnotationsFromComment(f.Comment, result)
					collectAnnotationsFromComment(f.InlineComment, result)
				case *proto.MapField:
					collectAnnotationsFromComment(f.Comment, result)
					collectAnnotationsFromComment(f.InlineComment, result)
				case *proto.OneOfField:
					collectAnnotationsFromComment(f.Comment, result)
					collectAnnotationsFromComment(f.InlineComment, result)
				}
			}
		case *proto.Enum:
			collectAnnotationsFromComment(v.Comment, result)
			for _, eElem := range v.Elements {
				if ef, ok := eElem.(*proto.EnumField); ok {
					collectAnnotationsFromComment(ef.Comment, result)
					collectAnnotationsFromComment(ef.InlineComment, result)
				}
			}
		}
	}
	return result
}

func collectAnnotationsFromComment(c *proto.Comment, result map[string]bool) {
	if c == nil {
		return
	}
	for _, name := range ExtractAnnotations(c) {
		result[name] = true
	}
}

func qualifiedName(pkg, name string) string {
	if pkg == "" {
		return name
	}
	return pkg + "." + name
}
