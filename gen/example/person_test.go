package example

import (
	"strings"
	"testing"

	"github.com/binadel/esdigo/json"
)

// TestPersonRoundTrip exercises the generated ReadJSON/WriteJSON: a document
// reads into the struct and writes back to an equivalent document.
func TestPersonRoundTrip(t *testing.T) {
	in := `{"firstName":"Ada","age":36,"score":9.5,"role":"admin","active":true}`

	var p Person
	if err := p.UnmarshalJSON([]byte(in)); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if string(p.FirstName.Value) != "Ada" || p.Age.Value != 36 || string(p.Role.Value) != "admin" || p.Active.Value != true {
		t.Errorf("decoded fields wrong: %+v", p)
	}
	// lastName absent -> not present
	if p.LastName.IsPresent() {
		t.Errorf("lastName should be absent")
	}

	out, err := p.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	// only present fields are written; order follows the struct
	for _, want := range []string{`"firstName":"Ada"`, `"age":36`, `"role":"admin"`, `"active":true`} {
		if !strings.Contains(string(out), want) {
			t.Errorf("output missing %s: %s", want, out)
		}
	}
	if strings.Contains(string(out), "lastName") {
		t.Errorf("absent lastName should not be written: %s", out)
	}
}

// TestPersonValidatorOK validates a well-formed document, including the format
// fields whose validators return typed values.
func TestPersonValidatorOK(t *testing.T) {
	var p Person
	in := `{"firstName":"Ada","age":36,"score":9.5,"role":"user","active":false,` +
		`"id":"123e4567-e89b-12d3-a456-426614174000","email":"ada@example.com","homepage":"https://example.com"}`
	if err := p.UnmarshalJSON([]byte(in)); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	r := NewPersonValidator().Validate(&p)
	if !r.FirstName.IsValid() || r.FirstName.Value != "Ada" {
		t.Errorf("firstName: %+v", r.FirstName.Errors)
	}
	if !r.Age.IsValid() || r.Age.Value != 36 {
		t.Errorf("age: %+v", r.Age.Errors)
	}
	if !r.Role.IsValid() || !r.Score.IsValid() || !r.Active.IsValid() {
		t.Errorf("role/score/active should be valid")
	}
	// format validators map to typed values
	if !r.Email.IsValid() || r.Email.Value == nil || r.Email.Value.Address != "ada@example.com" {
		t.Errorf("email should parse to *mail.Address: %+v", r.Email.Errors)
	}
	if !r.Id.IsValid() || r.Id.Value.String() != "123e4567-e89b-12d3-a456-426614174000" {
		t.Errorf("id should parse to uuid.UUID: %+v", r.Id.Errors)
	}
	if !r.Homepage.IsValid() || r.Homepage.Value == nil || r.Homepage.Value.Host != "example.com" {
		t.Errorf("homepage should parse to *url.URL: %+v", r.Homepage.Errors)
	}
}

// TestPersonFormatErrors checks that a malformed format field is reported.
func TestPersonFormatErrors(t *testing.T) {
	var p Person
	if err := p.UnmarshalJSON([]byte(`{"firstName":"Ada","age":36,"email":"not-an-email","id":"nope"}`)); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	r := NewPersonValidator().Validate(&p)
	if r.Email.IsValid() {
		t.Errorf("bad email should be invalid")
	}
	if r.Id.IsValid() {
		t.Errorf("bad uuid should be invalid")
	}
}

// TestPersonValidatorErrors validates a document that breaks several constraints.
func TestPersonValidatorErrors(t *testing.T) {
	// firstName missing (required), age above max, score not a multiple of 0.5,
	// role not in enum.
	var p Person
	if err := p.UnmarshalJSON([]byte(`{"age":200,"score":9.3,"role":"root","active":true}`)); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	r := NewPersonValidator().Validate(&p)

	if r.FirstName.IsValid() {
		t.Errorf("missing required firstName should be invalid")
	}
	if r.Age.IsValid() {
		t.Errorf("age 200 > max 150 should be invalid")
	}
	if r.Score.IsValid() {
		t.Errorf("score 9.3 not a multiple of 0.5 should be invalid")
	}
	if r.Role.IsValid() {
		t.Errorf("role 'root' not in enum should be invalid")
	}
	// structured error output carries the codes
	w := json.NewWriter(64)
	r.Role.WriteJSON(w)
	if !strings.Contains(string(w.Bytes()), "ENUM") {
		t.Errorf("role error should carry ENUM code: %s", w.Bytes())
	}
}
