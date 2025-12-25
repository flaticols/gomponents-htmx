// cem is a Custom Elements Manifest generator that uses the TypeScript-Go compiler
// to parse TypeScript source files and extract web component metadata.
//
// Usage:
//
//	cem [flags] <files...>
//
// Flags:
//
//	-o, --output    Output file path (default: custom-elements.json)
//	-d, --dir       Base directory for relative paths (default: current directory)
//	-p, --pretty    Pretty-print the JSON output
//	-h, --help      Show help message
//
// Example:
//
//	cem -o manifest.json src/*.ts
//	cem --pretty ./components/**/*.ts
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	var (
		outputFile string
		baseDir    string
		prettyJSON bool
		showHelp   bool
	)

	flag.StringVar(&outputFile, "o", "custom-elements.json", "Output file path")
	flag.StringVar(&outputFile, "output", "custom-elements.json", "Output file path")
	flag.StringVar(&baseDir, "d", ".", "Base directory for relative paths")
	flag.StringVar(&baseDir, "dir", ".", "Base directory for relative paths")
	flag.BoolVar(&prettyJSON, "p", false, "Pretty-print JSON output")
	flag.BoolVar(&prettyJSON, "pretty", false, "Pretty-print JSON output")
	flag.BoolVar(&showHelp, "h", false, "Show help message")
	flag.BoolVar(&showHelp, "help", false, "Show help message")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "cem - Custom Elements Manifest Generator\n\n")
		fmt.Fprintf(os.Stderr, "Uses the TypeScript-Go compiler to parse TypeScript files and\n")
		fmt.Fprintf(os.Stderr, "generate a custom-elements.json manifest with full inheritance support.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  cem [flags] <files...>\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  cem src/*.ts\n")
		fmt.Fprintf(os.Stderr, "  cem -o manifest.json -p src/components/*.ts\n")
		fmt.Fprintf(os.Stderr, "  cem --dir ./packages/ui --output dist/custom-elements.json src/**/*.ts\n")
	}

	flag.Parse()

	if showHelp || flag.NArg() == 0 {
		flag.Usage()
		if showHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}

	// Resolve base directory
	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving base directory: %v\n", err)
		os.Exit(1)
	}

	analyzer := NewAnalyzer(absBaseDir)

	// Collect all files from arguments (supports glob patterns)
	files := make([]string, 0)
	for _, pattern := range flag.Args() {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error expanding pattern %q: %v\n", pattern, err)
			continue
		}
		if len(matches) == 0 {
			// If no matches, treat as literal file path
			files = append(files, pattern)
		} else {
			files = append(files, matches...)
		}
	}

	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "No files to analyze\n")
		os.Exit(1)
	}

	// Analyze each file
	for _, file := range files {
		absPath, err := filepath.Abs(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving path %q: %v\n", file, err)
			continue
		}

		// Check file extension
		ext := filepath.Ext(absPath)
		if ext != ".ts" && ext != ".tsx" && ext != ".js" && ext != ".jsx" {
			fmt.Fprintf(os.Stderr, "Skipping non-TypeScript file: %s\n", file)
			continue
		}

		fmt.Fprintf(os.Stderr, "Analyzing: %s\n", file)
		if err := analyzer.AnalyzeFile(absPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error analyzing %q: %v\n", file, err)
			continue
		}
	}

	// Resolve inheritance across all classes
	fmt.Fprintf(os.Stderr, "Resolving class inheritance...\n")
	analyzer.ResolveInheritance()

	// Generate manifest
	manifest := analyzer.GetManifest()

	// Marshal to JSON
	var jsonData []byte
	if prettyJSON {
		jsonData, err = json.MarshalIndent(manifest, "", "  ")
	} else {
		jsonData, err = json.Marshal(manifest)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating JSON: %v\n", err)
		os.Exit(1)
	}

	// Write output
	if outputFile == "-" {
		fmt.Println(string(jsonData))
	} else {
		if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Manifest written to: %s\n", outputFile)
	}
}
