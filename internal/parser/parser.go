package parser

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/emicklei/proto"
)

// DiscoverProtoFiles recursively walks inputDir and returns relative
// paths of all files matching *.proto.
func DiscoverProtoFiles(inputDir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(inputDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".proto") {
			rel, err := filepath.Rel(inputDir, path)
			if err != nil {
				return err
			}
			files = append(files, rel)
		}
		return nil
	})
	return files, err
}

// ParseProtoFile parses a single .proto file and returns its AST.
func ParseProtoFile(path string) (*proto.Proto, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	parser := proto.NewParser(f)
	definition, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	return definition, nil
}

// ExtractPackage returns the package name from a parsed proto AST.
func ExtractPackage(def *proto.Proto) string {
	var pkg string
	proto.Walk(def, proto.WithPackage(func(p *proto.Package) {
		pkg = p.Name
	}))
	return pkg
}

// DefinitionInfo holds extracted info about a top-level proto definition.
type DefinitionInfo struct {
	FQN        string
	Kind       string   // "service", "message", "enum"
	Name       string
	References []string // FQNs of referenced types
}

// ExtractDefinitions walks a parsed proto AST and returns info about
// all top-level definitions with their type references.
func ExtractDefinitions(def *proto.Proto, pkg string) []DefinitionInfo {
	var defs []DefinitionInfo

	proto.Walk(def,
		proto.WithService(func(s *proto.Service) {
			fqn := qualifiedName(pkg, s.Name)
			var refs []string
			for _, elem := range s.Elements {
				if rpc, ok := elem.(*proto.RPC); ok {
					refs = appendRef(refs, pkg, rpc.RequestType)
					refs = appendRef(refs, pkg, rpc.ReturnsType)
				}
			}
			defs = append(defs, DefinitionInfo{
				FQN:        fqn,
				Kind:       "service",
				Name:       s.Name,
				References: refs,
			})
		}),
		proto.WithMessage(func(m *proto.Message) {
			if m.IsExtend {
				return
			}
			fqn := qualifiedName(pkg, m.Name)
			var refs []string
			for _, elem := range m.Elements {
				switch f := elem.(type) {
				case *proto.NormalField:
					if isUserType(f.Type) {
						refs = appendRef(refs, pkg, f.Type)
					}
				case *proto.MapField:
					if isUserType(f.Type) {
						refs = appendRef(refs, pkg, f.Type)
					}
				case *proto.OneOfField:
					if isUserType(f.Type) {
						refs = appendRef(refs, pkg, f.Type)
					}
				}
			}
			defs = append(defs, DefinitionInfo{
				FQN:        fqn,
				Kind:       "message",
				Name:       m.Name,
				References: refs,
			})
		}),
		proto.WithEnum(func(e *proto.Enum) {
			fqn := qualifiedName(pkg, e.Name)
			defs = append(defs, DefinitionInfo{
				FQN:        fqn,
				Kind:       "enum",
				Name:       e.Name,
				References: nil,
			})
		}),
	)

	return defs
}

func qualifiedName(pkg, name string) string {
	if pkg == "" {
		return name
	}
	return pkg + "." + name
}

// appendRef adds a type reference, qualifying it with the package if
// it's not already fully qualified.
func appendRef(refs []string, pkg, typeName string) []string {
	if typeName == "" {
		return refs
	}
	// If already qualified (contains a dot), use as-is
	if strings.Contains(typeName, ".") {
		return append(refs, typeName)
	}
	return append(refs, qualifiedName(pkg, typeName))
}

// isUserType returns true if the type name is not a built-in scalar.
func isUserType(typeName string) bool {
	switch typeName {
	case "double", "float", "int32", "int64", "uint32", "uint64",
		"sint32", "sint64", "fixed32", "fixed64", "sfixed32", "sfixed64",
		"bool", "string", "bytes":
		return false
	}
	return true
}
