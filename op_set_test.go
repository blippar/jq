// Copyright (c) 2016 Matt Ho <matt.ho@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package jq_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/blippar/jq"
)

var noCompilOptiSet reflect.Value

func BenchmarkSet(t *testing.B) {
	op := jq.Set(json.RawMessage(`{"Hello":"hello"}`))
	data := reflect.ValueOf(&struct{ Hello string }{Hello: "world"}) // {"hello":"world"}

	for i := 0; i < t.N; i++ {
		rv, err := op.Apply(data)
		noCompilOptiSet = rv
		if err != nil {
			t.Errorf("%v", err)
			t.FailNow()
			return
		}
	}
}

func TestSet(t *testing.T) {
	t.Parallel()
	testCases := map[string]struct {
		In         interface{}
		Op         jq.Op
		Expected   interface{}
		InExpected interface{}
		HasError   bool
	}{
		"struct": {
			In:         &struct{ Hello string }{Hello: "world"}, //`{"hello":"world"}`
			Op:         jq.Set(struct{ Hello string }{Hello: "you"}),
			Expected:   struct{ Hello string }{Hello: "you"},
			InExpected: struct{ Hello string }{Hello: "you"},
		},
		"nested": {
			In:         &BenchStruct{A: ABenchStruct{B: "world"}}, // `{"a":{"b":"world"}}`,
			Op:         jq.Chain(jq.Dot("A"), jq.Set(ABenchStruct{B: "you"})),
			Expected:   ABenchStruct{B: "you"},                 //`"world"`,
			InExpected: BenchStruct{A: ABenchStruct{B: "you"}}, // `{"a":{"b":"world"}}`,
		},
		"not a pointer": {
			In:         struct{ Hello string }{Hello: "world"}, //`{"hello":"world"}`
			Op:         jq.Set(struct{ Hello string }{Hello: "world"}),
			Expected:   nil,
			InExpected: nil,
			HasError:   true,
		},
		"notsamestruct": {
			In:         &struct{}{},
			Op:         jq.Chain(jq.Set(&struct{ A int }{A: 1})),
			Expected:   struct{}{},
			InExpected: struct{}{},
			HasError:   true,
		},
		"mapreplace": {
			In:         &map[int]int{1: 1},
			Op:         jq.Set(map[int]int{2: 2}),
			Expected:   map[int]int{2: 2},
			InExpected: map[int]int{2: 2},
		},
		"set a field not present": {
			In:         &struct{ A string }{A: ""},
			Op:         jq.Set(json.RawMessage(`{"B":"hello"}`)),
			Expected:   struct{}{},
			InExpected: struct{}{},
			HasError:   true,
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			data, err := tc.Op.Apply(reflect.ValueOf(tc.In))
			if tc.HasError {
				if err == nil {
					t.Errorf("Expected an error (%v) , got %v ", tc.HasError, err)
					t.FailNow()
				}
			} else {
				if err != nil {
					t.Errorf("Expected an error (%v) , got %v ", tc.HasError, err)
					t.FailNow()
				}
				if !reflect.DeepEqual(data.Interface(), tc.Expected) {
					t.Errorf("Expected %v (%T), got %v (%T)", tc.Expected, tc.Expected, data.Interface(), data.Interface())
				}

				if !reflect.DeepEqual(reflect.ValueOf(tc.In).Elem().Interface(), tc.InExpected) {
					t.Errorf("Expected %v (%T), got %v (%T)", tc.InExpected, tc.InExpected, reflect.ValueOf(tc.In).Elem().Interface(), reflect.ValueOf(tc.In).Elem().Interface())
				}
				if err != nil {
					t.Errorf("Expected %v (%T), got %v (%T)", tc.Expected, tc.Expected, data.Interface(), data.Interface())
				}
			}
		})
	}
}
