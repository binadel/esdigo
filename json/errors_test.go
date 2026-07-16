package json

import (
	"errors"
	"strings"
	"testing"
)

func TestErrorKinds(t *testing.T) {
	cases := []struct {
		name  string
		input string
		kind  ErrorKind
	}{
		{"unexpected char", `{"a":x}`, KindSyntax},
		{"leading zero", `01`, KindSyntax},
		{"truncated object", `{"a":`, KindTruncated},
		{"truncated string", `"abc`, KindTruncated},
		{"empty input", ``, KindTruncated},
		{"depth exceeded", strings.Repeat("[", 200), KindDepthLimit},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := NewReader([]byte(c.input)).ReadJSON()
			if err == nil {
				t.Fatalf("expected an error for %q", c.input)
			}
			var se *SyntaxError
			if !errors.As(err, &se) {
				t.Fatalf("error is not *SyntaxError: %T", err)
			}
			if se.Kind != c.kind {
				t.Errorf("Kind = %q, want %q", se.Kind, c.kind)
			}
			if se.Offset < 0 {
				t.Errorf("Offset = %d, want >= 0", se.Offset)
			}
		})
	}
}
