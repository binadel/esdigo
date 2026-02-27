package main

import (
	"github.com/bindadel/esdigo/json"
	"github.com/bindadel/esdigo/json/types"
)

type ProductResponse struct {
	Product types.Object[*Product] `json:"product"`
}

func (p *ProductResponse) MarshalJSON() ([]byte, error) {
	w := json.NewWriter()
	p.WriteJSON(w)
	return w.Build()
}

func (p *ProductResponse) UnmarshalJSON(data []byte) error {
	r := json.NewReader(data)
	p.ReadJSON(r)
	return r.Error()
}

func (p *ProductResponse) WriteJSON(w *json.Writer) bool {
	w.BeginObject()
	if p.Product.ShouldWrite() {
		w.WriteRawString(`"product":`)
		if !p.Product.WriteJSON(w) {
			return false
		}
	}
	w.EndObject()
	return true
}

func (p *ProductResponse) ReadJSON(r *json.Reader) bool {
	r.SkipWhitespace()
	if r.BeginObject() {
		r.SkipWhitespace()

		if r.EndObject() {
			return true
		}

		for {
			if name, ok := r.ReadString(); ok {
				r.SkipWhitespace()

				if r.NameSeparator() {
					ok := false
					switch name {
					case "product":
						ok = p.Product.ReadJSON(r)
					default:
						ok = r.SkipValue()
					}
					if !ok {
						return false
					}
				} else {
					r.SetSyntaxError("expected a name-separator ':' after name")
					return false
				}

				r.SkipWhitespace()

				if r.EndObject() {
					return true
				}

				if !r.ValueSeparator() {
					r.SetSyntaxError("expected either end-object '}' or value-separator ','")
					return false
				}

				r.SkipWhitespace()
			} else {
				r.SetSyntaxError("expected a name after begin-object '{' or value-separator ','")
				return false
			}
		}
	}
	return false
}
