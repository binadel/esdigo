package validation

import (
	"testing"

	"github.com/binadel/esdigo/json"
	"github.com/binadel/esdigo/json/types"
)

// objElem reads only a JSON object; anything else fails its ReadJSON, so the
// enclosing Object becomes Defined-but-invalid (wrong type).
type objElem struct{}

func (o *objElem) ReadJSON(r *json.Reader) bool {
	if t, _ := r.PeekType(); t != json.ValueTypeObject {
		return false
	}
	return r.SkipValue()
}
func (o *objElem) WriteJSON(w *json.Writer) bool { w.BeginObject(); w.EndObject(); return true }

type objField = types.Object[objElem, *objElem]

func readObjField(s string) objField { return readInto[objField](s) }

func newObj() *Object[objElem, *objElem] { return NewObject[objElem, *objElem]("o") }

func TestObjectPresenceAndNull(t *testing.T) {
	var absent objField
	if r := newObj().Required().Validate(absent); r.IsValid() {
		t.Errorf("required+absent should fail")
	} else {
		mustContain(t, "required", resultJSON(r), "REQUIRED")
	}

	null := readObjField("null")
	if r := newObj().Validate(null); !r.IsValid() || r.Defined {
		t.Errorf("nullable null: valid=%v defined=%v", r.IsValid(), r.Defined)
	}
	if r := newObj().NotNull().Validate(null); r.IsValid() {
		t.Errorf("notNull null should fail")
	} else {
		mustContain(t, "notNull", resultJSON(r), "NOT_NULL")
	}

	// wrong type -> OBJECT (not STRING)
	wrong := readObjField("42")
	if r := newObj().Validate(wrong); r.IsValid() {
		t.Errorf("number into object should be invalid")
	} else {
		mustContain(t, "wrong-type", resultJSON(r), "OBJECT")
	}
}

func TestObjectValue(t *testing.T) {
	v := readObjField("{}")
	if r := newObj().Validate(v); !r.IsValid() || r.Value == nil {
		t.Errorf("valid object: valid=%v value=%v", r.IsValid(), r.Value)
	}
}
