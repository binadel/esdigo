package errors

import (
	"testing"

	"github.com/binadel/esdigo/json"
	"github.com/binadel/esdigo/validation/rules"
)

// writeJSON renders an error's WriteJSON output as a string.
func writeJSON(t *testing.T, e interface{ WriteJSON(*json.Writer) bool }) string {
	t.Helper()
	w := json.NewWriter(64)
	if !e.WriteJSON(w) {
		t.Fatalf("WriteJSON returned false")
	}
	b, err := w.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return string(b)
}

// TestBasicError: Error() returns the message, and WriteJSON emits code + message.
func TestBasicError(t *testing.T) {
	if Required.Error() != "field is required" {
		t.Errorf("Error() = %q", Required.Error())
	}
	if Required.Code != rules.Required {
		t.Errorf("Required.Code = %q, want %q", Required.Code, rules.Required)
	}
	want := `{"code":"REQUIRED","message":"field is required"}`
	if got := writeJSON(t, Required); got != want {
		t.Errorf("BasicError JSON = %s, want %s", got, want)
	}
}

// TestIntParamError: an integer parameter is serialized as a JSON number under its
// param key (note the key strings carry a leading comma).
func TestIntParamError(t *testing.T) {
	e := &IntParamError{BasicError: MinLength, ParamKey: ParamKeyMinLength, ParamValue: 3}
	want := `{"code":"MIN_LENGTH","message":"value must be at least the minimum length","minLength":3}`
	if got := writeJSON(t, e); got != want {
		t.Errorf("IntParamError JSON = %s, want %s", got, want)
	}
}

// TestNumberParamError: a numeric bound is written verbatim from pre-formatted JSON
// number bytes, so any numeric type reports without knowing its Go type here.
func TestNumberParamError(t *testing.T) {
	e := &NumberParamError{BasicError: Minimum, ParamKey: ParamKeyMinimum, ParamValue: []byte("5")}
	want := `{"code":"MINIMUM","message":"value must be at least the minimum","minimum":5}`
	if got := writeJSON(t, e); got != want {
		t.Errorf("NumberParamError JSON = %s, want %s", got, want)
	}
}

// TestRawParamError: a pre-serialized parameter (an enum array) is echoed verbatim.
func TestRawParamError(t *testing.T) {
	e := &RawParamError{BasicError: Enum, ParamKey: ParamKeyEnum, ParamValue: []byte(`["a","b"]`)}
	want := `{"code":"ENUM","message":"value must be one of the allowed values","enum":["a","b"]}`
	if got := writeJSON(t, e); got != want {
		t.Errorf("RawParamError JSON = %s, want %s", got, want)
	}
}
