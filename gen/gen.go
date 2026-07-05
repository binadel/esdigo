// Package gen generates esdigo model + validator Go source from a JSON Schema.
//
// The pipeline is schema.Parse (input) → ir.Build (resolve to Go types and
// validator chains) → emit.File (render gofmt'd source). OpenAPI 3.1 uses the
// same schema dialect, so its component schemas can feed the same pipeline.
package gen

import (
	"fmt"

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

// GenerateOpenAPI turns an OpenAPI 3.1 document into a Go source file, generating
// a type for every schema under components.schemas (their 2020-12 dialect and
// #/components/schemas/ refs flow through the same pipeline).
func GenerateOpenAPI(data []byte, pkg string) ([]byte, error) {
	doc, err := schema.ParseOpenAPI(data)
	if err != nil {
		return nil, err
	}
	if len(doc.Components.Schemas) == 0 {
		return nil, fmt.Errorf("no components.schemas in the OpenAPI document")
	}
	file, err := ir.BuildAll(pkg, doc.Components.Schemas)
	if err != nil {
		return nil, err
	}
	return emit.File(file)
}

// GenerateAuto detects the input: an OpenAPI document generates all its component
// schemas (name is ignored); a bare JSON Schema generates the single named root.
func GenerateAuto(data []byte, pkg, name string) ([]byte, error) {
	if schema.IsOpenAPI(data) {
		return GenerateOpenAPI(data, pkg)
	}
	return Generate(data, pkg, name)
}
