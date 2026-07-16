package json

import (
	"strings"
	"testing"
)

func nestedArray(n int) string {
	return strings.Repeat("[", n) + strings.Repeat("]", n)
}

func nestedObject(n int) string {
	return strings.Repeat(`{"a":`, n-1) + "{}" + strings.Repeat("}", n-1)
}

func TestMaxDepth_Boundary(t *testing.T) {
	// Exactly at the default limit parses; one level past it is rejected.
	if _, err := NewReader([]byte(nestedArray(defaultMaxDepth))).ReadJSON(); err != nil {
		t.Errorf("%d-deep array rejected: %v", defaultMaxDepth, err)
	}
	if _, err := NewReader([]byte(nestedArray(defaultMaxDepth + 1))).ReadJSON(); err == nil {
		t.Errorf("%d-deep array accepted, want depth error", defaultMaxDepth+1)
	}
	if _, err := NewReader([]byte(nestedObject(defaultMaxDepth))).ReadJSON(); err != nil {
		t.Errorf("%d-deep object rejected: %v", defaultMaxDepth, err)
	}
	if _, err := NewReader([]byte(nestedObject(defaultMaxDepth + 1))).ReadJSON(); err == nil {
		t.Errorf("%d-deep object accepted, want depth error", defaultMaxDepth+1)
	}
}

// TestMaxDepth_NoStackOverflow is the whole point of the guard: a pathologically
// deep payload must be REJECTED, never crash the process with a stack overflow
// (a fatal error recover() cannot catch). Both the DOM and skip paths are checked.
func TestMaxDepth_NoStackOverflow(t *testing.T) {
	huge := []byte(strings.Repeat("[", 1_000_000))

	if _, err := NewReader(huge).ReadJSON(); err == nil {
		t.Error("million-deep input accepted by ReadJSON")
	}
	r := NewReader(huge)
	r.SkipWhitespace()
	if r.SkipValue() {
		t.Error("million-deep input accepted by SkipValue")
	}
}

func TestMaxDepth_Configurable(t *testing.T) {
	r := NewReader([]byte(nestedArray(5)))
	r.SetMaxDepth(4)
	if _, err := r.ReadJSON(); err == nil {
		t.Error("5-deep with maxDepth=4 accepted")
	}

	r = NewReader([]byte(nestedArray(4)))
	r.SetMaxDepth(4)
	if _, err := r.ReadJSON(); err != nil {
		t.Errorf("4-deep with maxDepth=4 rejected: %v", err)
	}

	// negative disables the limit (safe to actually recurse a few hundred frames)
	r = NewReader([]byte(nestedArray(500)))
	r.SetMaxDepth(-1)
	if _, err := r.ReadJSON(); err != nil {
		t.Errorf("500-deep with maxDepth=-1 (unlimited) rejected: %v", err)
	}
}

// Depth must be released on EndArray/EndObject so that many sibling containers
// (wide but shallow) do not accumulate depth.
func TestMaxDepth_BalancesSiblings(t *testing.T) {
	var sb strings.Builder
	sb.WriteByte('[')
	for i := 0; i < 20000; i++ {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(`{"k":[1,2]}`)
	}
	sb.WriteByte(']')
	if _, err := NewReader([]byte(sb.String())).ReadJSON(); err != nil {
		t.Errorf("wide shallow document rejected: %v", err)
	}
}
