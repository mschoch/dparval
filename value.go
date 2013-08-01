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
	"fmt"
	"strconv"

	jsonpointer "github.com/dustin/go-jsonpointer"
	json "github.com/dustin/gojson"
)

type Undefined struct {
	Path string
}

func (this *Undefined) Error() string {
	if this.Path != "" {
		return fmt.Sprintf("%s is not defined", this.Path)
	}
	return fmt.Sprint("not defined")
}

type ValueChannel chan *Value
type ValueCollection []*Value

// type Value interface {
// 	Path(path string) (Value, error)
// 	SetPath(path string, val interface{})
// 	Index(index int) (Value, error)
// 	SetIndex(index int, val interface{})
// 	Type() int
// 	Value() interface{}

// 	AddMeta(key string, val interface{})
// 	Meta() Value
// }

type Value struct {
	raw         []byte
	parsedValue interface{}
	alias       map[string]*Value
	parsedType  int
	meta        *Value
}

func NewValue(val interface{}) *Value {
	switch val := val.(type) {
	case nil:
		return newNullValue()
	case bool:
		return newBooleanValue(val)
	case float64:
		return newNumberValue(val)
	case string:
		return newStringValue(val)
	case []interface{}:
		return newArrayValue(val)
	case map[string]interface{}:
		return newObjectValue(val)
	case *Value:
		return val
	default:
		panic(fmt.Sprintf("Cannot create value for type %T", val))
	}
}

func newNullValue() *Value {
	rv := Value{
		parsedType: NULL,
	}
	return &rv
}

func newBooleanValue(val bool) *Value {
	rv := Value{
		parsedType:  BOOLEAN,
		parsedValue: val,
	}
	return &rv
}

func newNumberValue(val float64) *Value {
	rv := Value{
		parsedType:  NUMBER,
		parsedValue: val,
	}
	return &rv
}

func newStringValue(val string) *Value {
	rv := Value{
		parsedType:  STRING,
		parsedValue: val,
	}
	return &rv
}

func newArrayValue(val []interface{}) *Value {
	rv := Value{
		parsedType: ARRAY,
	}

	parsedValue := make([]*Value, len(val))
	for i, v := range val {
		switch v := v.(type) {
		case *Value:
			parsedValue[i] = v
		default:
			parsedValue[i] = NewValue(v)
		}
	}
	rv.parsedValue = parsedValue

	return &rv
}

func newObjectValue(val map[string]interface{}) *Value {
	rv := Value{
		parsedType: OBJECT,
	}

	parsedValue := make(map[string]*Value)
	for k, v := range val {
		switch v := v.(type) {
		case *Value:
			parsedValue[k] = v
		default:
			parsedValue[k] = NewValue(v)
		}
	}
	rv.parsedValue = parsedValue
	return &rv
}

func NewValueFromBytes(bytes []byte) *Value {
	rv := Value{
		raw:         bytes,
		parsedType:  -1,
		parsedValue: nil,
		alias:       nil,
	}
	err := json.Validate(bytes)
	if err != nil {
		rv.parsedType = NOT_JSON
	} else {
		rv.parsedType = identifyType(bytes)
	}
	return &rv
}

func identifyType(bytes []byte) int {
	for _, b := range bytes {
		switch b {
		case '{':
			return OBJECT
		case '[':
			return ARRAY
		case '"':
			return STRING
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return NUMBER
		case 't', 'f':
			return BOOLEAN
		case 'n':
			return NULL
		}
	}
	panic("Unable to identify type of valid JSON")
	return -1
}

func (this *Value) Path(path string) (*Value, error) {
	// aliases always have priority

	if this.alias != nil {
		result, ok := this.alias[path]
		if ok {
			return result, nil
		}
	}
	// next we already parsed, used that
	switch parsedValue := this.parsedValue.(type) {
	case map[string]*Value:
		result, ok := parsedValue[path]
		if ok {
			return result, nil
		}
	}
	// finally, consult the raw bytes
	if this.raw != nil {
		res, err := jsonpointer.Find(this.raw, "/"+path)
		if err != nil {
			return nil, err
		}
		if res != nil {
			return NewValueFromBytes(res), nil
		}
	}

	return nil, &Undefined{path}
}

func (this *Value) SetPath(path string, val interface{}) {

	if this.parsedType == OBJECT {
		switch parsedValue := this.parsedValue.(type) {
		case map[string]*Value:
			// if we've already parsed the object, store it there
			switch val := val.(type) {
			case *Value:
				parsedValue[path] = val
			default:
				parsedValue[path] = NewValue(val)
			}
		case nil:
			// if not store it in alias
			if this.alias == nil {
				this.alias = make(map[string]*Value)
			}
			switch val := val.(type) {
			case *Value:
				this.alias[path] = val
			default:
				this.alias[path] = NewValue(val)
			}

		}
	}
}

func (this *Value) Index(index int) (*Value, error) {
	// aliases always have priority
	if this.alias != nil {
		result, ok := this.alias[strconv.Itoa(index)]
		if ok {
			return result, nil
		}
	}
	// next we already parsed, used that
	switch parsedValue := this.parsedValue.(type) {
	case []*Value:
		if index >= 0 && index < len(parsedValue) {
			result := parsedValue[index]
			return result, nil
		} else {
			// this way it behaves consistent with jsonpointer below
			return nil, &Undefined{}
		}
	}
	// finally, consult the raw bytes
	if this.raw != nil {
		res, err := jsonpointer.Find(this.raw, "/"+strconv.Itoa(index))
		if err != nil {
			return nil, err
		}
		if res != nil {
			return NewValueFromBytes(res), nil
		}
	}
	return nil, &Undefined{}
}

func (this *Value) SetIndex(index int, val interface{}) {
	if this.parsedType == ARRAY && index >= 0 {
		switch parsedValue := this.parsedValue.(type) {
		case []*Value:
			if index < len(parsedValue) {
				// if we've already parsed the object, store it there
				switch val := val.(type) {
				case *Value:
					parsedValue[index] = val
				default:
					parsedValue[index] = NewValue(val)
				}
			}
		case nil:
			// if not store it in alias
			if this.alias == nil {
				this.alias = make(map[string]*Value)
			}
			switch val := val.(type) {
			case *Value:
				this.alias[strconv.Itoa(index)] = val
			default:
				this.alias[strconv.Itoa(index)] = NewValue(val)
			}

		}
	}
}

func (this *Value) Type() int {
	return this.parsedType
}

func (this *Value) AddMeta(key string, val interface{}) {
	if this.meta == nil {
		this.meta = NewValue(map[string]interface{}{})
	}
	this.meta.SetPath(key, val)
}

func (this *Value) Meta() *Value {
	return this.meta
}

// the Value function exits this type system
// it can be the most expensive call, as it may force parsing
// large amounts of previously unparsed json
func (this *Value) Value() interface{} {
	if this.parsedValue != nil || this.parsedType == NULL {
		rv := devalue(this.parsedValue)
		if this.alias != nil {
			overlayAlias(rv, this.alias)
		}
		return rv
	} else if this.parsedType != NOT_JSON {
		err := json.Unmarshal(this.raw, &this.parsedValue)
		if err != nil {
			panic("unexpected parse error on valid JSON")
		}
		// if there are any aliases, we must make a safe copy
		// and then overlay them
		if this.alias != nil {
			// we cannot damange the original parsed value
			rv := safeCopy(this.parsedValue)
			overlayAlias(rv, this.alias)
			return rv
		} else {
			// otherwise its safe to return directly
			return this.parsedValue
		}
	} else {
		return nil
	}
}

func devalue(base interface{}) interface{} {
	switch base := base.(type) {
	case map[string]*Value:
		rv := make(map[string]interface{}, len(base))
		for k, v := range base {
			if v.Type() != NOT_JSON {
				rv[k] = v.Value()
			}
		}
		return rv
	case []*Value:
		rv := make([]interface{}, len(base))
		for i, v := range base {
			rv[i] = v.Value()
		}
		return rv
	default:
		return base
	}
}

func safeCopy(base interface{}) interface{} {
	switch base := base.(type) {
	case map[string]interface{}:
		rv := make(map[string]interface{}, len(base))
		for k, v := range base {
			rv[k] = v
		}
		return rv
	case []interface{}:
		rv := make([]interface{}, len(base))
		for i, v := range base {
			rv[i] = v
		}
		return rv
	default:
		return base
	}
}

func overlayAlias(base interface{}, alias map[string]*Value) {
	switch base := base.(type) {
	case map[string]interface{}:
		for k, v := range alias {
			if v.Type() != NOT_JSON {
				base[k] = v.Value()
			}
		}
	case []interface{}:
		for k, v := range alias {
			bigi, err := strconv.ParseInt(k, 10, 32)
			if err != nil {
				panic("alias index for array could not be converted to int")
			}
			i := int(bigi)
			if i >= 0 && i < len(base) {
				if v.Type() != NOT_JSON {
					base[i] = v.Value()
				}
			}
		}
	}
}

const (
	NOT_JSON = iota
	NULL
	BOOLEAN
	NUMBER
	STRING
	ARRAY
	OBJECT
)
