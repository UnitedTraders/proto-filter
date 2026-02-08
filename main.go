package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/emicklei/proto"

	"github.com/unitedtraders/proto-filter/internal/config"
	"github.com/unitedtraders/proto-filter/internal/deps"
	"github.com/unitedtraders/proto-filter/internal/filter"
	"github.com/unitedtraders/proto-filter/internal/parser"
	"github.com/unitedtraders/proto-filter/internal/writer"
)

func main() {
	os.Exit(run())
}

func run() int {
	inputDir := flag.String("input", "", "path to directory containing source .proto files")
	outputDir := flag.String("output", "", "path to directory where filtered .proto files are written")
	configFile := flag.String("config", "", "path to YAML filter configuration file")
	verbose := flag.Bool("verbose", false, "print processing summary to stderr")

	flag.Parse()

	if *inputDir == "" || *outputDir == "" {
		fmt.Fprintln(os.Stderr, "proto-filter: error: --input and --output flags are required")
		flag.Usage()
		return 1
	}

	absInput, err := filepath.Abs(*inputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "proto-filter: error: %v\n", err)
		return 1
	}
	absOutput, err := filepath.Abs(*outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "proto-filter: error: %v\n", err)
		return 1
	}

	if absInput == absOutput {
		fmt.Fprintln(os.Stderr, "proto-filter: error: input and output directories must be different")
		return 1
	}

	info, err := os.Stat(absInput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "proto-filter: error: input directory not found: %s\n", absInput)
		return 1
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "proto-filter: error: input path is not a directory: %s\n", absInput)
		return 1
	}

	// Load filter config if provided
	var cfg *config.FilterConfig
	if *configFile != "" {
		var err error
		cfg, err = config.LoadConfig(*configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "proto-filter: error: %v\n", err)
			return 2
		}
	}

	// Discover proto files
	files, err := parser.DiscoverProtoFiles(absInput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "proto-filter: error: discovering proto files: %v\n", err)
		return 1
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "proto-filter: warning: no .proto files found in input directory")
		return 0
	}

	// Parse all files
	type parsedFile struct {
		rel string
		def *proto.Proto
		pkg string
	}
	parsed := make([]parsedFile, 0, len(files))

	for _, rel := range files {
		inPath := filepath.Join(absInput, rel)
		def, err := parser.ParseProtoFile(inPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "proto-filter: error: parsing %s: %v\n", rel, err)
			return 1
		}
		pkg := parser.ExtractPackage(def)
		parsed = append(parsed, parsedFile{rel, def, pkg})
	}

	// Determine total definitions count
	totalDefs := 0
	graph := deps.NewGraph()
	for _, pf := range parsed {
		defs := parser.ExtractDefinitions(pf.def, pf.pkg)
		totalDefs += len(defs)
		for _, d := range defs {
			graph.AddDefinition(&deps.Definition{
				FQN:        d.FQN,
				Kind:       d.Kind,
				File:       pf.rel,
				References: d.References,
			})
		}
	}

	// Apply filtering if config provided
	filesToWrite := make(map[string]bool)
	var keepFQNs map[string]bool
	includedCount := totalDefs
	excludedCount := 0

	if cfg != nil && !cfg.IsPassThrough() {
		allFQNs := make([]string, 0, len(graph.Nodes))
		for fqn := range graph.Nodes {
			allFQNs = append(allFQNs, fqn)
		}

		included, err := filter.ApplyFilter(cfg, allFQNs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "proto-filter: error: %v\n", err)
			return 2
		}

		// Resolve transitive dependencies
		includedList := make([]string, 0, len(included))
		for fqn := range included {
			includedList = append(includedList, fqn)
		}
		allNeeded := graph.TransitiveDeps(includedList)

		keepFQNs = make(map[string]bool)
		for _, fqn := range allNeeded {
			keepFQNs[fqn] = true
		}

		includedCount = len(keepFQNs)
		excludedCount = totalDefs - includedCount

		// Determine required files
		requiredFiles := graph.RequiredFiles(allNeeded)
		for _, f := range requiredFiles {
			filesToWrite[f] = true
		}
	} else {
		// No filtering: write all files
		for _, pf := range parsed {
			filesToWrite[pf.rel] = true
		}
	}

	// Prune and write
	writtenCount := 0
	for _, pf := range parsed {
		if !filesToWrite[pf.rel] {
			continue
		}
		if keepFQNs != nil {
			filter.PruneAST(pf.def, pf.pkg, keepFQNs)
		}
		outPath := filepath.Join(absOutput, pf.rel)
		if err := writer.WriteProtoFile(pf.def, outPath); err != nil {
			fmt.Fprintf(os.Stderr, "proto-filter: error: writing %s: %v\n", pf.rel, err)
			return 1
		}
		writtenCount++
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "proto-filter: processed %d files, %d definitions\n", len(files), totalDefs)
		fmt.Fprintf(os.Stderr, "proto-filter: included %d definitions, excluded %d\n", includedCount, excludedCount)
		fmt.Fprintf(os.Stderr, "proto-filter: wrote %d files to %s\n", writtenCount, *outputDir)
	}

	return 0
}
