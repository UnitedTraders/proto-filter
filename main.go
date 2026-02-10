package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

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
		if err := cfg.Validate(); err != nil {
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

	// Pass 1: Prune, filter, convert block comments, collect annotations
	type processedFile struct {
		pf  parsedFile
		skip bool // true if file has no remaining definitions after filtering
	}
	processed := make([]processedFile, 0, len(parsed))
	servicesRemoved := 0
	methodsRemoved := 0
	orphansRemoved := 0
	allAnnotations := make(map[string]bool)

	for _, pf := range parsed {
		if !filesToWrite[pf.rel] {
			continue
		}
		if keepFQNs != nil {
			filter.PruneAST(pf.def, pf.pkg, keepFQNs)
		}

		skip := false
		// Annotation-based filtering
		if cfg != nil && cfg.HasAnnotations() {
			var sr, mr int
			if cfg.HasAnnotationExclude() {
				sr = filter.FilterServicesByAnnotation(pf.def, cfg.Annotations.Exclude)
				mr = filter.FilterMethodsByAnnotation(pf.def, cfg.Annotations.Exclude)
			} else if cfg.HasAnnotationInclude() {
				mr = filter.IncludeMethodsByAnnotation(pf.def, cfg.Annotations.Include)
				sr = filter.IncludeServicesByAnnotation(pf.def, cfg.Annotations.Include)
			}
			servicesRemoved += sr
			methodsRemoved += mr
			filter.RemoveEmptyServices(pf.def)
			if sr > 0 || mr > 0 {
				orphansRemoved += filter.RemoveOrphanedDefinitions(pf.def, pf.pkg)
			}

			if !filter.HasRemainingDefinitions(pf.def) {
				skip = true
			}
		}

		// Convert block comments to single-line style
		filter.ConvertBlockComments(pf.def)

		// Collect annotations for strict mode check
		if cfg != nil && cfg.StrictSubstitutions {
			for name := range filter.CollectAllAnnotations(pf.def) {
				allAnnotations[name] = true
			}
		}

		processed = append(processed, processedFile{pf: pf, skip: skip})
	}

	// Strict substitution check: fail if any annotations lack a mapping
	if cfg != nil && cfg.StrictSubstitutions && len(allAnnotations) > 0 {
		var missing []string
		for name := range allAnnotations {
			if _, ok := cfg.Substitutions[name]; !ok {
				missing = append(missing, name)
			}
		}
		if len(missing) > 0 {
			sort.Strings(missing)
			fmt.Fprintf(os.Stderr, "proto-filter: error: unsubstituted annotations found: %s\n", joinNames(missing))
			return 2
		}
	}

	// Pass 2: Substitute annotations and write output
	writtenCount := 0
	substitutionCount := 0
	for _, pf := range processed {
		if pf.skip {
			continue
		}

		if cfg != nil && cfg.HasSubstitutions() {
			substitutionCount += filter.SubstituteAnnotations(pf.pf.def, cfg.Substitutions)
		}

		outPath := filepath.Join(absOutput, pf.pf.rel)
		if err := writer.WriteProtoFile(pf.pf.def, outPath); err != nil {
			fmt.Fprintf(os.Stderr, "proto-filter: error: writing %s: %v\n", pf.pf.rel, err)
			return 1
		}
		writtenCount++
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "proto-filter: processed %d files, %d definitions\n", len(files), totalDefs)
		fmt.Fprintf(os.Stderr, "proto-filter: included %d definitions, excluded %d\n", includedCount, excludedCount)
		if cfg != nil && cfg.HasAnnotations() {
			fmt.Fprintf(os.Stderr, "proto-filter: removed %d services by annotation, %d methods by annotation, %d orphaned definitions\n", servicesRemoved, methodsRemoved, orphansRemoved)
		}
		if cfg != nil && cfg.HasSubstitutions() {
			fmt.Fprintf(os.Stderr, "proto-filter: substituted %d annotations\n", substitutionCount)
		}
		fmt.Fprintf(os.Stderr, "proto-filter: wrote %d files to %s\n", writtenCount, *outputDir)
	}

	return 0
}

func joinNames(names []string) string {
	result := ""
	for i, name := range names {
		if i > 0 {
			result += ", "
		}
		result += name
	}
	return result
}
