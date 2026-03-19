package main

import (
	"net/mail"

	"github.com/binadel/esdigo/json"
	"github.com/binadel/esdigo/json/types"
	"github.com/binadel/esdigo/validation"
)

type Product struct {
	Title       types.String
	Description types.String
	Price       types.Number
	IsPublished types.Boolean
	Email       types.String
}

func (p *Product) CreateValue() *Product {
	return &Product{}
}

func (p *Product) MarshalJSON() ([]byte, error) {
	w := json.NewWriter(128)
	p.WriteJSON(w)
	return w.Build()
}

func (p *Product) UnmarshalJSON(data []byte) error {
	r := json.NewReader(data)
	p.ReadJSON(r)
	return r.Error()
}

func (p *Product) WriteJSON(w *json.Writer) bool {
	if p == nil {
		w.WriteNull()
		return true
	}
	needsComma := false
	w.BeginObject()
	if p.Title.IsPresent() {
		w.WriteRawString(`"title":`)
		if !p.Title.WriteJSON(w) {
			return false
		}
		needsComma = true
	}
	if p.Price.IsPresent() {
		if needsComma {
			w.ValueSeparator()
		}
		w.WriteRawString(`"price":`)
		if !p.Price.WriteJSON(w) {
			return false
		}
		needsComma = true
	}
	if p.IsPublished.IsPresent() {
		if needsComma {
			w.ValueSeparator()
		}
		w.WriteRawString(`"isPublished":`)
		if !p.IsPublished.WriteJSON(w) {
			return false
		}
		needsComma = true
	}
	if p.Email.IsPresent() {
		if needsComma {
			w.ValueSeparator()
		}
		w.WriteRawString(`"email":`)
		if !p.Email.WriteJSON(w) {
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
					case "email":
						ok = p.Email.ReadJSON(r)
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

type ValidatedProduct struct {
	Title       validation.Result[string]
	IsPublished validation.Result[bool]
	Email       validation.Result[*mail.Address]
}

type ProductValidator struct {
	Title       *validation.String
	IsPublished *validation.Boolean
	Email       *validation.Email
}

func NewProductValidator() *ProductValidator {
	return &ProductValidator{
		Title:       validation.NewString("title").Required().NotNull().MinLength(2).MaxLength(256),
		IsPublished: validation.NewBoolean("isPublished").Required().NotNull(),
		Email:       validation.NewString("email").Required().Pattern(validation.PatternEmail).Email(),
	}
}

func (v *ProductValidator) Validate(p *Product) *ValidatedProduct {
	return &ValidatedProduct{
		Title:       v.Title.Validate(p.Title),
		IsPublished: v.IsPublished.Validate(p.IsPublished),
		Email:       v.Email.Validate(p.Email),
	}
}
