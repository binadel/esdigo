// Command esdigo-gen reads a JSON Schema or OpenAPI document (JSON or YAML) and
// writes the generated esdigo model + validator Go source.
//
// Usage:
//
//	esdigo-gen [flags] <schema.json|.yaml>   # one schema -> -o / -outdir / stdout
//	esdigo-gen [flags] < schema.yaml         # read from stdin
//	esdigo-gen [flags] <schema-dir>          # every *.json/*.yaml -> combined .go
//
// Flags:
//
//	-pkg     output package name (default "models")
//	-name    root Go type name (single schema only; default: derived from the filename)
//	-o       output file for the combined single file (default: stdout)
//	-outdir  output directory; writes <pkg>.go there (created if missing)
//	-split   write one file per generated type into -outdir (e.g. asset_response.go)
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/binadel/esdigo/gen"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

// run is the testable entry point: it returns the process exit code.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("esdigo-gen", flag.ContinueOnError)
	fs.SetOutput(stderr)
	pkg := fs.String("pkg", "models", "output package name")
	name := fs.String("name", "", "root Go type name (default: derived from the input filename)")
	out := fs.String("o", "", "output file for the combined single file (default: stdout)")
	outdir := fs.String("outdir", "", "output directory; writes <pkg>.go there (created if missing)")
	split := fs.Bool("split", false, "write one file per generated type into -outdir")
	fs.Usage = func() {
		_, _ = fmt.Fprintln(stderr, "usage: esdigo-gen [flags] <schema.json|.yaml|schema-dir>")
		_, _ = fmt.Fprintln(stderr, "reads a JSON/YAML schema or OpenAPI doc (or a directory of them) and writes generated Go; omit the file to read stdin.")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() > 1 {
		_, _ = fmt.Fprintln(stderr, "error: expected at most one schema file or directory")
		return 2
	}
	if *out != "" && *split {
		_, _ = fmt.Fprintln(stderr, "error: -split writes multiple files; use -outdir, not -o")
		return 2
	}
	if *out != "" && *outdir != "" {
		_, _ = fmt.Fprintln(stderr, "error: use either -o or -outdir, not both")
		return 2
	}

	// A directory input merges every schema into one namespace.
	if path := fs.Arg(0); path != "" {
		info, err := os.Stat(path)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		if info.IsDir() {
			return runDir(path, *pkg, *outdir, *split, stderr)
		}
	}

	data, typeName, err := readSchema(fs.Arg(0), *name, stdin)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	// -split writes one file per type into the output directory.
	if *split {
		files, err := gen.GenerateAutoFiles(data, *pkg, typeName)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		return writeFiles(files, resolveOutdir(*outdir, fs.Arg(0)), stderr)
	}

	src, err := gen.GenerateAuto(data, *pkg, typeName)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	// -outdir writes the combined file as <pkg>.go into the directory.
	if *outdir != "" {
		return writeFiles(map[string][]byte{*pkg + ".go": src}, *outdir, stderr)
	}
	if *out == "" {
		if _, err := stdout.Write(src); err != nil {
			_, _ = fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		return 0
	}
	if err := os.WriteFile(*out, src, 0o644); err != nil {
		_, _ = fmt.Fprintf(stderr, "error: writing %s: %v\n", *out, err)
		return 1
	}
	return 0
}

// runDir generates from every *.json/*.yaml schema in dir, into outdir (defaulting
// to dir): one combined <pkg>.go, or one file per type with split. The schemas
// share one namespace: types are deduplicated by name and $ref resolves across files.
func runDir(dir, pkg, outdir string, split bool, stderr io.Writer) int {
	if outdir == "" {
		outdir = dir
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	files := map[string][]byte{}
	for _, e := range entries {
		if e.IsDir() || !isSchemaFile(e.Name()) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "error: %s: %v\n", e.Name(), err)
			return 1
		}
		files[e.Name()] = data
	}
	if len(files) == 0 {
		_, _ = fmt.Fprintf(stderr, "error: no .json/.yaml schema files in %s\n", dir)
		return 1
	}

	if split {
		out, err := gen.GenerateDirFiles(files, pkg)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		return writeFiles(out, outdir, stderr)
	}

	src, err := gen.GenerateDir(files, pkg)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	return writeFiles(map[string][]byte{pkg + ".go": src}, outdir, stderr)
}

// writeFiles writes each name->source into dir, creating dir if missing.
func writeFiles(files map[string][]byte, dir string, stderr io.Writer) int {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		_, _ = fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	for name, src := range files {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, src, 0o644); err != nil {
			_, _ = fmt.Fprintf(stderr, "error: writing %s: %v\n", p, err)
			return 1
		}
	}
	return 0
}

// resolveOutdir is the output directory for a single-schema input: the -outdir
// flag when set, else the input file's directory, else the current directory
// (stdin has no path).
func resolveOutdir(outdir, path string) string {
	switch {
	case outdir != "":
		return outdir
	case path != "":
		return filepath.Dir(path)
	default:
		return "."
	}
}

// readSchema loads the schema from path (or stdin when path is empty) and resolves
// the root type name — the -name flag if set, else derived from the filename, else
// "Root" for stdin.
func readSchema(path, name string, stdin io.Reader) (data []byte, typeName string, err error) {
	if path == "" {
		data, err = io.ReadAll(stdin)
		typeName = name
		if typeName == "" {
			typeName = "Root"
		}
		return data, typeName, err
	}

	data, err = os.ReadFile(path)
	typeName = name
	if typeName == "" {
		typeName = typeNameFromFile(path)
	}
	return data, typeName, err
}

// typeNameFromFile derives a type name from a schema filename by stripping the
// .schema/.json/.yaml/.yml extensions; gen normalizes it to an exported Go
// identifier (e.g. "user-profile.schema.json" -> "user-profile" -> "UserProfile").
func typeNameFromFile(path string) string {
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, ".json")
	base = strings.TrimSuffix(base, ".yaml")
	base = strings.TrimSuffix(base, ".yml")
	base = strings.TrimSuffix(base, ".schema")
	if base == "" {
		return "Root"
	}
	return base
}

// isSchemaFile reports whether name has a schema-file extension the generator
// reads: JSON or YAML.
func isSchemaFile(name string) bool {
	return strings.HasSuffix(name, ".json") ||
		strings.HasSuffix(name, ".yaml") ||
		strings.HasSuffix(name, ".yml")
}
