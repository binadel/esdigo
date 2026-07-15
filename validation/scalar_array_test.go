package validation

import (
	"testing"

	"github.com/binadel/esdigo/json"
	"github.com/binadel/esdigo/json/types"
)

func readStrArr(s string) *types.StringArray {
	var v types.StringArray
	v.ReadJSON(json.NewReader([]byte(s)))
	return &v
}

func scalarArrayJSON(r Result[[]string]) string {
	w := json.NewWriter(64)
	r.WriteJSON(w)
	return string(w.Bytes())
}

func TestScalarArray(t *testing.T) {
	if r := NewScalarArray[string]("t").MinItems(1).MaxItems(3).Validate(readStrArr(`["a","b"]`)); !r.IsValid() || len(r.Value) != 2 {
		t.Errorf("[a,b] in [1,3]: %s", scalarArrayJSON(r))
	}
	if r := NewScalarArray[string]("t").MinItems(2).Validate(readStrArr(`["a"]`)); r.IsValid() {
		t.Errorf("minItems 2 on 1 should fail")
	} else {
		mustContain(t, "min", scalarArrayJSON(r), "MIN_ITEMS")
	}
	if r := NewScalarArray[string]("t").MaxItems(1).Validate(readStrArr(`["a","b"]`)); r.IsValid() {
		t.Errorf("maxItems 1 on 2 should fail")
	} else {
		mustContain(t, "max", scalarArrayJSON(r), "MAX_ITEMS")
	}
	if r := NewScalarArray[string]("t").UniqueItems().Validate(readStrArr(`["a","a"]`)); r.IsValid() {
		t.Errorf("duplicate should fail uniqueItems")
	} else {
		mustContain(t, "uniq", scalarArrayJSON(r), "UNIQUE_ITEMS")
	}

	// presence / null
	if r := NewScalarArray[string]("t").Validate(readStrArr("null")); !r.IsValid() || r.Defined {
		t.Errorf("nullable null should pass")
	}
	if r := NewScalarArray[string]("t").NotNull().Validate(readStrArr("null")); r.IsValid() {
		t.Errorf("notNull null should fail")
	}
	var absent types.StringArray
	if r := NewScalarArray[string]("t").Required().Validate(&absent); r.IsValid() {
		t.Errorf("required absent should fail")
	}
	if r := NewScalarArray[string]("t").Validate(readStrArr("42")); r.IsValid() {
		t.Errorf("number into array should be invalid")
	} else {
		mustContain(t, "wrong-type", scalarArrayJSON(r), "ARRAY")
	}
}

// TestScalarArrayTypes covers the numeric and boolean instantiations.
func TestScalarArrayTypes(t *testing.T) {
	var ints types.Int64Array
	ints.ReadJSON(json.NewReader([]byte(`[1,2,2]`)))
	if r := NewScalarArray[int64]("n").UniqueItems().Validate(&ints); r.IsValid() {
		t.Errorf("duplicate ints should fail uniqueItems")
	}

	var bools types.BooleanArray
	bools.ReadJSON(json.NewReader([]byte(`[true,false]`)))
	if r := NewScalarArray[bool]("b").Validate(&bools); !r.IsValid() || len(r.Value) != 2 {
		t.Errorf("bools should pass")
	}
}
