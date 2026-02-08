package writer

import (
	"os"
	"path/filepath"

	"github.com/emicklei/proto"
	"github.com/emicklei/proto-contrib/pkg/protofmt"
)

// WriteProtoFile formats the AST and writes it to the given path,
// creating parent directories as needed.
func WriteProtoFile(definition *proto.Proto, outputPath string) error {
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	formatter := protofmt.NewFormatter(f, "  ")
	formatter.Format(definition)
	return nil
}
