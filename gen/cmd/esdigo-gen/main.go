// Command esdigo-gen reads a JSON Schema and writes the generated esdigo model +
// validator Go source.
//
// Usage:
//
//	esdigo-gen [flags] <schema.json>
//	esdigo-gen [flags] < schema.json    # read from stdin
//
// Flags:
//
//	-pkg   output package name (default "models")
//	-name  root Go type name (default: derived from the input filename)
//	-o     output file (default: stdout)
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
	out := fs.String("o", "", "output file (default: stdout)")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "usage: esdigo-gen [flags] <schema.json>")
		fmt.Fprintln(stderr, "reads a JSON Schema and writes generated Go; omit the file to read stdin.")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() > 1 {
		fmt.Fprintln(stderr, "error: expected at most one schema file")
		return 2
	}

	data, typeName, err := readSchema(fs.Arg(0), *name, stdin)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	src, err := gen.Generate(data, *pkg, typeName)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	if *out == "" {
		if _, err := stdout.Write(src); err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		return 0
	}
	if err := os.WriteFile(*out, src, 0o644); err != nil {
		fmt.Fprintf(stderr, "error: writing %s: %v\n", *out, err)
		return 1
	}
	return 0
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
// .schema/.json extensions; gen normalizes it to an exported Go identifier
// (e.g. "user-profile.schema.json" -> "user-profile" -> "UserProfile").
func typeNameFromFile(path string) string {
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, ".json")
	base = strings.TrimSuffix(base, ".schema")
	if base == "" {
		return "Root"
	}
	return base
}
