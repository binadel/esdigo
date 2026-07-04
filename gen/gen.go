// Package gen generates esdigo model + validator Go source from a JSON Schema.
//
// The pipeline is schema.Parse (input) → ir.Build (resolve to Go types and
// validator chains) → emit.File (render gofmt'd source). OpenAPI 3.1 uses the
// same schema dialect, so its component schemas can feed the same pipeline.
package gen

import (
	"github.com/binadel/esdigo/gen/emit"
	"github.com/binadel/esdigo/gen/ir"
	"github.com/binadel/esdigo/gen/schema"
)

// Generate turns a JSON Schema document into a Go source file. pkg is the output
// package name and name is the Go type name for the root object.
func Generate(data []byte, pkg, name string) ([]byte, error) {
	root, err := schema.Parse(data)
	if err != nil {
		return nil, err
	}
	file, err := ir.Build(pkg, name, root)
	if err != nil {
		return nil, err
	}
	return emit.File(file)
}
