package main

import (
	"fmt"

	"github.com/bindadel/esdigo/json"
	"github.com/bindadel/esdigo/json/types"
)

type Product struct {
	Title       types.String
	Description types.String
	Price       types.Number
	IsPublished types.Boolean
}

func (p *Product) MarshalJSON() ([]byte, error) {
	w := json.NewWriter()
	p.WriteJSON(w)
	return w.Build()
}

func (p *Product) UnmarshalJSON(data []byte) error {
	r := json.NewReader(data)
	p.ReadJSON(r)
	return r.Error()
}

func (p *Product) WriteJSON(w *json.Writer) bool {
	needsComma := false
	w.BeginObject()
	if p.Title.ShouldWrite() {
		w.WriteRawString(`"title":`)
		if !p.Title.WriteJSON(w) {
			return false
		}
		needsComma = true
	}
	if p.Description.ShouldWrite() {
		if needsComma {
			w.ValueSeparator()
		}
		w.WriteRawString(`"description":`)
		if !p.Description.WriteJSON(w) {
			return false
		}
		needsComma = true
	}
	if p.Price.ShouldWrite() {
		if needsComma {
			w.ValueSeparator()
		}
		w.WriteRawString(`"price":`)
		if !p.Price.WriteJSON(w) {
			return false
		}
		needsComma = true
	}
	if p.IsPublished.ShouldWrite() {
		if needsComma {
			w.ValueSeparator()
		}
		w.WriteRawString(`"isPublished":`)
		if !p.IsPublished.WriteJSON(w) {
			return false
		}
		needsComma = true
	}
	w.EndObject()
	return true
}

func (p *Product) ReadJSON(r *json.Reader) bool {
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
					case "title":
						ok = p.Title.ReadJSON(r)
					case "description":
						ok = p.Description.ReadJSON(r)
					case "price":
						ok = p.Price.ReadJSON(r)
					case "isPublished":
						ok = p.IsPublished.ReadJSON(r)
					}
					if !ok {
						return false
					}
				} else {
					r.SetSyntaxError("expected a name-separator ':' after name")
					return false
				}

				if r.EndObject() {
					return true
				}

				if !r.ValueSeparator() {
					r.SetSyntaxError("expected either end-object '}' or value-separator ','")
					return false
				}
			} else {
				r.SetSyntaxError("expected a name after begin-object '{' or value-separator ','")
				return false
			}
		}
	}
	return false
}

func main() {
	p := &Product{}
	p.Title.Set("MacBook Air M5")
	p.Price.SetNull()
	p.IsPublished.Set(true)

	data, err := p.MarshalJSON()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(data))
	}

	p1 := &Product{}
	err = p1.UnmarshalJSON(data)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(p1)
}
