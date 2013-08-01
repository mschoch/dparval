//  Copyright (c) 2013 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package dparval

import (
	"reflect"
	"testing"
)

func TestTypeRecognition(t *testing.T) {

	var tests = []struct {
		input        []byte
		expectedType int
	}{
		{[]byte(`asdf`), NOT_JSON},
		{[]byte(`null`), NULL},
		{[]byte(`3.65`), NUMBER},
		{[]byte(`-3.65`), NUMBER},
		{[]byte(`"hello"`), STRING},
		{[]byte(`["hello"]`), ARRAY},
		{[]byte(`{"hello":7}`), OBJECT},

		// with misc whitespace
		{[]byte(` asdf`), NOT_JSON},
		{[]byte(` null`), NULL},
		{[]byte(` 3.65`), NUMBER},
		{[]byte(` "hello"`), STRING},
		{[]byte("\t[\"hello\"]"), ARRAY},
		{[]byte("\n{\"hello\":7}"), OBJECT},
	}

	for _, test := range tests {
		val := NewValueFromBytes(test.input)
		actualType := val.Type()
		if actualType != test.expectedType {
			t.Errorf("Expected type of %s to be %d, got %d", string(test.input), test.expectedType, actualType)
		}
	}
}

func TestPathAccess(t *testing.T) {

	val := NewValueFromBytes([]byte(`{"name":"marty","address":{"street":"sutton oaks"}}`))

	var tests = []struct {
		path   string
		result *Value
		err    error
	}{
		{"name", &Value{raw: []byte(`"marty"`), parsedType: STRING}, nil},
		{"address", &Value{raw: []byte(`{"street":"sutton oaks"}`), parsedType: OBJECT}, nil},
		{"dne", nil, &Undefined{"dne"}},
	}

	for _, test := range tests {
		newval, err := val.Path(test.path)
		if !reflect.DeepEqual(err, test.err) {
			t.Errorf("Expected error %v got error %v for path %s", test.err, err, test.path)
		}
		if !reflect.DeepEqual(newval, test.result) {
			t.Errorf("Expected %v got %v for path %s", test.result, newval, test.path)
		}
	}
}

func TestIndexAccess(t *testing.T) {

	val := NewValueFromBytes([]byte(`["marty",{"type":"contact"}]`))

	var tests = []struct {
		index  int
		result *Value
		err    error
	}{
		{0, &Value{raw: []byte(`"marty"`), parsedType: STRING}, nil},
		{1, &Value{raw: []byte(`{"type":"contact"}`), parsedType: OBJECT}, nil},
		{2, nil, &Undefined{}},
	}

	for _, test := range tests {
		newval, err := val.Index(test.index)
		if !reflect.DeepEqual(err, test.err) {
			t.Errorf("Expected error %v got error %v for index %d", test.err, err, test.index)
		}
		if !reflect.DeepEqual(newval, test.result) {
			t.Errorf("Expected %v got %v for index %d", test.result, newval, test.index)
		}
	}

	val = NewValue([]interface{}{"marty", map[string]interface{}{"type": "contact"}})

	tests = []struct {
		index  int
		result *Value
		err    error
	}{
		{0, &Value{parsedValue: "marty", parsedType: STRING}, nil},
		{1, &Value{parsedValue: map[string]*Value{"type": NewValue("contact")}, parsedType: OBJECT}, nil},
		{2, nil, &Undefined{}},
	}

	for _, test := range tests {
		newval, err := val.Index(test.index)
		if !reflect.DeepEqual(err, test.err) {
			t.Errorf("Expected error %v got error %v for index %d", test.err, err, test.index)
		}
		if !reflect.DeepEqual(newval, test.result) {
			t.Errorf("Expected %v got %v for index %d", test.result, newval, test.index)
		}
	}
}

func TestAliasOverrides(t *testing.T) {
	val := NewValueFromBytes([]byte(`{"name":"marty","address":{"street":"sutton oaks"}}`))
	val.SetPath("name", "steve")
	name, err := val.Path("name")
	if err != nil {
		t.Errorf("Error getting path name")
	}
	nameVal := name.Value()
	if nameVal != "steve" {
		t.Errorf("Expected name to be steve, got %v", nameVal)
	}

	val = NewValueFromBytes([]byte(`["marty",{"type":"contact"}]`))
	val.SetIndex(0, "gerald")
	name, err = val.Index(0)
	if err != nil {
		t.Errorf("Error getting path name")
	}
	nameVal = name.Value()
	if nameVal != "gerald" {
		t.Errorf("Expected name to be gerald, got %v", nameVal)
	}
}

func TestMeta(t *testing.T) {
	val := NewValueFromBytes([]byte(`{"name":"marty","address":{"street":"sutton oaks"}}`))
	val.AddMeta("id", "doc1")

	idVal, err := val.Meta().Path("id")
	if err != nil {
		t.Errorf("Error access id path in meta")
	}
	id := idVal.Value()
	if id != "doc1" {
		t.Errorf("Expected id doc1, got %v", id)
	}

}

func TestRealWorkflow(t *testing.T) {
	// get a doc from some source
	doc := NewValueFromBytes([]byte(`{"name":"marty","address":{"street":"sutton oaks"}}`))
	doc.AddMeta("id", "doc1")

	// mutate the document somehow
	active := NewValue(true)
	doc.SetPath("active", active)

	testActiveVal, err := doc.Path("active")
	if err != nil {
		t.Errorf("Error accessing active in doc")
	}

	testActive := testActiveVal.Value()
	if testActive != true {
		t.Errorf("Expected active true, got %v", testActive)
	}

	// create an alias of this documents
	top := NewValue(map[string]interface{}{"bucket": doc, "another": "rad"})

	testDoc, err := top.Path("bucket")
	if err != nil {
		t.Errorf("Error accessing bucket in top")
	}
	if !reflect.DeepEqual(testDoc, doc) {
		t.Errorf("Expected doc %v to match testDoc %v", doc, testDoc)
	}

	testRad, err := top.Path("another")
	if err != nil {
		t.Errorf("Error accessing another in top")
	}
	expectedRad := NewValue("rad")
	if !reflect.DeepEqual(testRad, expectedRad) {
		t.Errorf("Expected %v, got %v for rad", expectedRad, testRad)
	}

	// now project some value from the doc to a top-level alias
	addressVal, err := doc.Path("address")
	if err != nil {
		t.Errorf("Error access path address in doc")
	}

	top.SetPath("a", addressVal)

	// now access "a.street"
	aVal, err := top.Path("a")
	if err != nil {
		t.Errorf("Error access a in top")
	}
	streetVal, err := aVal.Path("street")
	if err != nil {
		t.Errorf("Error accessing street in a")
	}
	street := streetVal.Value()
	if street != "sutton oaks" {
		t.Errorf("Expected sutton oaks, got %v", street)
	}
}

func TestUndefined(t *testing.T) {

	x := Undefined{"property"}
	err := x.Error()

	if err != "property is not defined" {
		t.Errorf("Expected property is not defined, got %v", err)
	}

	y := Undefined{}
	err = y.Error()

	if err != "not defined" {
		t.Errorf("Expected not defined, got %v", err)
	}

}

func TestValue(t *testing.T) {
	var tests = []struct {
		input         *Value
		expectedValue interface{}
	}{
		{NewValue(nil), nil},
		{NewValue(true), true},
		{NewValue(false), false},
		{NewValue(1.0), 1.0},
		{NewValue(3.14), 3.14},
		{NewValue(-7.0), -7.0},
		{NewValue(""), ""},
		{NewValue("marty"), "marty"},
		{NewValue([]interface{}{"marty"}), []interface{}{"marty"}},
		{NewValue([]interface{}{NewValue("marty2")}), []interface{}{"marty2"}},
		{NewValue(map[string]interface{}{"marty": "cool"}), map[string]interface{}{"marty": "cool"}},
		{NewValue(map[string]interface{}{"marty3": NewValue("cool")}), map[string]interface{}{"marty3": "cool"}},
		{NewValueFromBytes([]byte("null")), nil},
		{NewValueFromBytes([]byte("true")), true},
		{NewValueFromBytes([]byte("false")), false},
		{NewValueFromBytes([]byte("1")), 1.0},
		{NewValueFromBytes([]byte("3.14")), 3.14},
		{NewValueFromBytes([]byte("-7")), -7.0},
		{NewValueFromBytes([]byte("\"\"")), ""},
		{NewValueFromBytes([]byte("\"marty\"")), "marty"},
		{NewValueFromBytes([]byte("[\"marty\"]")), []interface{}{"marty"}},
		{NewValueFromBytes([]byte("{\"marty\": \"cool\"}")), map[string]interface{}{"marty": "cool"}},
		{NewValueFromBytes([]byte("abc")), nil},
		// new value from existing value
		{NewValue(NewValue(true)), true},
	}

	for _, test := range tests {
		val := test.input.Value()
		if !reflect.DeepEqual(val, test.expectedValue) {
			t.Errorf("Expected %#v, got %#v for %#v", test.expectedValue, val, test.input)
		}
	}
}

func TestValueOverlay(t *testing.T) {
	val := NewValueFromBytes([]byte("{\"marty\": \"cool\"}"))
	val.SetPath("marty", "ok")
	expectedVal := map[string]interface{}{"marty": "ok"}
	actualVal := val.Value()
	if !reflect.DeepEqual(expectedVal, actualVal) {
		t.Errorf("Expected %v, got %v, for value of %v", expectedVal, actualVal, val)
	}

	val = NewValue(map[string]interface{}{"marty": "cool"})
	val.SetPath("marty", "ok")
	actualVal = val.Value()
	if !reflect.DeepEqual(expectedVal, actualVal) {
		t.Errorf("Expected %v, got %v, for value of %v", expectedVal, actualVal, val)
	}

	val = NewValueFromBytes([]byte("[\"marty\"]"))
	val.SetIndex(0, "gerald")
	expectedVal2 := []interface{}{"gerald"}
	actualVal = val.Value()
	if !reflect.DeepEqual(expectedVal2, actualVal) {
		t.Errorf("Expected %v, got %v, for value of %v", expectedVal2, actualVal, val)
	}

	val = NewValue([]interface{}{"marty"})
	val.SetIndex(0, "gerald")
	expectedVal2 = []interface{}{"gerald"}
	actualVal = val.Value()
	if !reflect.DeepEqual(expectedVal2, actualVal) {
		t.Errorf("Expected %v, got %v, for value of %v", expectedVal2, actualVal, val)
	}
}

func TestComplexOverlay(t *testing.T) {
	// in this case we start with JSON bytes
	// then add an alias
	// then call value which causes it to be parsed
	// then call value again, which goes through a different path with the already parsed data
	val := NewValueFromBytes([]byte("{\"marty\": \"cool\"}"))
	val.SetPath("marty", "ok")
	expectedVal := map[string]interface{}{"marty": "ok"}
	actualVal := val.Value()
	if !reflect.DeepEqual(expectedVal, actualVal) {
		t.Errorf("Expected %v, got %v, for value of %v", expectedVal, actualVal, val)
	}
	// now repeat the call to value
	actualVal = val.Value()
	if !reflect.DeepEqual(expectedVal, actualVal) {
		t.Errorf("Expected %v, got %v, for value of %v", expectedVal, actualVal, val)
	}
}
