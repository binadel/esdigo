package example

import (
	"strings"
	"testing"
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
	// billing was absent
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

// TestOrderValidation checks the object-level validators plus manual recursion
// into a child validator.
func TestOrderValidation(t *testing.T) {
	var o Order
	in := `{"id":"123e4567-e89b-12d3-a456-426614174000",` +
		`"customer":{"name":"Ada","email":"ada@example.com"},` +
		`"shippingAddress":{"city":"Paris"}}`
	if err := o.UnmarshalJSON([]byte(in)); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	r := NewOrderValidator().Validate(&o)
	if !r.Id.IsValid() || !r.Customer.IsValid() || !r.ShippingAddress.IsValid() {
		t.Errorf("well-formed order should validate: id=%v customer=%v ship=%v",
			r.Id.Errors, r.Customer.Errors, r.ShippingAddress.Errors)
	}
	// recurse manually into the child validator
	if r.Customer.Value == nil {
		t.Fatalf("customer result should carry the value")
	}
	cv := NewOrderCustomerValidator().Validate(r.Customer.Value)
	if !cv.Name.IsValid() || cv.Name.Value != "Ada" {
		t.Errorf("nested customer.name should validate: %+v", cv.Name.Errors)
	}
}

// TestOrderMissingRequired confirms the object-level required checks fire.
func TestOrderMissingRequired(t *testing.T) {
	var o Order
	if err := o.UnmarshalJSON([]byte(`{}`)); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	r := NewOrderValidator().Validate(&o)
	if r.Id.IsValid() {
		t.Errorf("missing required id should be invalid")
	}
	if r.Customer.IsValid() {
		t.Errorf("missing required customer should be invalid")
	}
}
