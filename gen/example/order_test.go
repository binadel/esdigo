package example

import (
	"strings"
	"testing"

	"github.com/binadel/esdigo/json"
	"github.com/binadel/esdigo/validation"
)

// TestOrderRoundTrip exercises nested object read/write across an inline object
// (customer) and a shared $ref (shippingAddress -> Address).
func TestOrderRoundTrip(t *testing.T) {
	in := `{"id":"123e4567-e89b-12d3-a456-426614174000",` +
		`"customer":{"name":"Ada","email":"ada@example.com"},` +
		`"shippingAddress":{"city":"Paris","street":"1 rue"}}`

	var o Order
	if err := o.UnmarshalJSON([]byte(in)); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if o.Customer.Value == nil || string(o.Customer.Value.Name.Value) != "Ada" {
		t.Errorf("nested customer.name not decoded: %+v", o.Customer.Value)
	}
	if o.ShippingAddress.Value == nil || string(o.ShippingAddress.Value.City.Value) != "Paris" {
		t.Errorf("ref shippingAddress.city not decoded: %+v", o.ShippingAddress.Value)
	}
	if o.BillingAddress.IsPresent() {
		t.Errorf("billingAddress should be absent")
	}

	out, err := o.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, want := range []string{`"customer":{`, `"name":"Ada"`, `"shippingAddress":{`, `"city":"Paris"`} {
		if !strings.Contains(string(out), want) {
			t.Errorf("output missing %s: %s", want, out)
		}
	}
}

// TestOrderValidation validates a well-formed order and reaches typed values
// through the recursive result, with no manual descent.
func TestOrderValidation(t *testing.T) {
	var o Order
	in := `{"id":"123e4567-e89b-12d3-a456-426614174000",` +
		`"customer":{"name":"Ada","email":"ada@example.com"},` +
		`"shippingAddress":{"city":"Paris"}}`
	if err := o.UnmarshalJSON([]byte(in)); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	r := NewOrderValidator().Validate(&o)
	if !r.IsValid() {
		t.Errorf("well-formed order should be valid; failures=%s", failuresJSON(r.Failures()))
	}
	// nested typed values, reached directly
	if r.Customer.Name.Value != "Ada" {
		t.Errorf("customer.name value: %q", r.Customer.Name.Value)
	}
	if r.Customer.Email.Value == nil || r.Customer.Email.Value.Address != "ada@example.com" {
		t.Errorf("customer.email should parse to *mail.Address")
	}
	if r.ShippingAddress.City.Value != "Paris" {
		t.Errorf("shippingAddress.city value: %q", r.ShippingAddress.City.Value)
	}
}

// TestOrderMissingRequired confirms the object-level required checks fire and the
// whole tree reports invalid.
func TestOrderMissingRequired(t *testing.T) {
	var o Order
	if err := o.UnmarshalJSON([]byte(`{}`)); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	r := NewOrderValidator().Validate(&o)
	if r.IsValid() {
		t.Errorf("empty order should be invalid")
	}
	if r.Id.IsValid() {
		t.Errorf("missing required id should be invalid")
	}
	if r.Customer.Object.IsValid() {
		t.Errorf("missing required customer should fail object-level check")
	}
}

// TestOrderNestedErrorPath is the key check for recursion: a broken nested field
// surfaces in the flat report with its FULL path, e.g. ["customer","name"].
func TestOrderNestedErrorPath(t *testing.T) {
	var o Order
	in := `{"id":"123e4567-e89b-12d3-a456-426614174000",` +
		`"customer":{"name":""},` + // name below minLength 1
		`"shippingAddress":{"city":"Paris"}}`
	if err := o.UnmarshalJSON([]byte(in)); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	r := NewOrderValidator().Validate(&o)
	if r.IsValid() {
		t.Errorf("order with empty customer.name should be invalid")
	}
	if r.Customer.Name.IsValid() {
		t.Errorf("customer.name should be invalid")
	}
	report := failuresJSON(r.Failures())
	if !strings.Contains(report, `["customer","name"]`) {
		t.Errorf("flat report should carry the nested path [customer,name]: %s", report)
	}
	if !strings.Contains(report, "MIN_LENGTH") {
		t.Errorf("flat report should carry the MIN_LENGTH code: %s", report)
	}
}

// failuresJSON serializes a flat failure list as a JSON array for assertions.
func failuresJSON(failures []validation.FieldResult) string {
	w := json.NewWriter(128)
	w.BeginArray()
	for i, f := range failures {
		if i > 0 {
			w.ValueSeparator()
		}
		f.WriteJSON(w)
	}
	w.EndArray()
	return string(w.Bytes())
}
