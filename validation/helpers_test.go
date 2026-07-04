package validation

import (
	"strings"
	"testing"

	"github.com/binadel/esdigo/json"
)

// readInto reads s into a fresh wrapper T through its *T ReadJSON, returning the
// populated value. A zero-value T (never read) models an absent field.
func readInto[T any, PT interface {
	*T
	json.ValueReader
}](s string) T {
	var v T
	PT(&v).ReadJSON(json.NewReader([]byte(s)))
	return v
}

// resultJSON serializes a Result to its JSON string for asserting on codes/params.
func resultJSON[T any](r Result[T]) string {
	w := json.NewWriter(64)
	r.WriteJSON(w)
	return string(w.Bytes())
}

// isValid reports r.IsValid on a by-value Result (which a chained Validate returns
// as a non-addressable temporary, so its pointer method can't be called inline).
func isValid[T any](r Result[T]) bool {
	return r.IsValid()
}

func mustContain(t *testing.T, name, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("%s: expected %q in output, got %s", name, want, got)
	}
}
