package api

import (
	"strings"
	"testing"

	"github.com/binadel/esdigo/json"
	"github.com/binadel/esdigo/validation"
)

// TestUserRoundTrip exercises a type generated from an OpenAPI component,
// including a cross-component $ref (User.address -> Address).
func TestUserRoundTrip(t *testing.T) {
	in := `{"id":7,"email":"ada@example.com","address":{"city":"Paris","zip":"75001"},"roles":["admin","user"]}`

	var u User
	if err := u.UnmarshalJSON([]byte(in)); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if u.Id.Value != 7 || u.Address.Value == nil || string(u.Address.Value.City.Value) != "Paris" {
		t.Errorf("decoded fields wrong: %+v", u)
	}
	if len(u.Roles.Value) != 2 || string(u.Roles.Value[0].Value) != "admin" {
		t.Errorf("roles not decoded: %+v", u.Roles.Value)
	}

	out, err := u.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(out), `"address":{"city":"Paris"`) {
		t.Errorf("nested address not written: %s", out)
	}
}

// TestUserValidation validates a good user and reports nested/cross-component
// failures with full paths.
func TestUserValidation(t *testing.T) {
	var u User
	if err := u.UnmarshalJSON([]byte(`{"id":7,"email":"ada@example.com","address":{"city":"Paris","zip":"75001"}}`)); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if r := NewUserValidator().Validate(&u); !r.IsValid() {
		t.Errorf("valid user should pass; failures=%s", failuresJSON(r.Failures()))
	}

	// invalid: malformed email, and a bad zip pattern in the referenced Address
	var bad User
	if err := bad.UnmarshalJSON([]byte(`{"id":7,"email":"notanemail","address":{"city":"Paris","zip":"abc"}}`)); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	r := NewUserValidator().Validate(&bad)
	if r.IsValid() {
		t.Errorf("bad user should be invalid")
	}
	report := failuresJSON(r.Failures())
	if !strings.Contains(report, "EMAIL") {
		t.Errorf("missing email should fail with EMAIL: %s", report)
	}
	if !strings.Contains(report, `["address","zip"]`) || !strings.Contains(report, "PATTERN") {
		t.Errorf("bad zip should fail at address.zip with PATTERN: %s", report)
	}
}

// TestUserRoleElementValidation checks per-element validation of a scalar array:
// each role must meet the item's minLength.
func TestUserRoleElementValidation(t *testing.T) {
	var u User
	if err := u.UnmarshalJSON([]byte(`{"id":7,"email":"a@b.com","address":{"city":"Paris"},"roles":["ok","x"]}`)); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	r := NewUserValidator().Validate(&u)
	if r.IsValid() {
		t.Errorf("a too-short role should make the user invalid")
	}
	if len(r.RolesItems) != 2 || !r.RolesItems[0].IsValid() || r.RolesItems[1].IsValid() {
		t.Errorf("role element validity wrong: %+v", r.RolesItems)
	}
	report := failuresJSON(r.Failures())
	if !strings.Contains(report, "MIN_LENGTH") {
		t.Errorf("short role should fail MIN_LENGTH: %s", report)
	}
	// the second role (index 1) is the offender; the path carries the index
	if !strings.Contains(report, `["roles","1"]`) {
		t.Errorf("failure should carry the indexed element path: %s", report)
	}
}

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
