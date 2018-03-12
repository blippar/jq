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

	"github.com/savaki/jq"
)

func BenchmarkAddition(t *testing.B) {
	op := jq.Addition(json.RawMessage(`["world"]`))
	data := reflect.ValueOf(&[]string{"hello"}) // {"hello":"world"}

	for i := 0; i < t.N; i++ {
		_, err := op.Apply(data)
		if err != nil {
			t.Errorf("%v", err)
			t.FailNow()
			return
		}
	}
}
func TestAddition(t *testing.T) {
	testCases := map[string]struct {
		In         interface{}
		Op         jq.Op
		Expected   interface{}
		InExpected interface{}
		HasError   bool
	}{
		"struct": {
			In:         &struct{ Hello string }{Hello: "world"}, //`{"hello":"world"}`
			Op:         jq.Addition(map[string]string{"Hello": "you"}),
			Expected:   struct{ Hello string }{Hello: "you"},
			InExpected: struct{ Hello string }{Hello: "you"},
		},
		"nested": {
			In:         &BenchStruct{A: ABenchStruct{B: "world"}}, // `{"a":{"b":"world"}}`,
			Op:         jq.Chain(jq.Dot("A"), jq.Addition(map[string]string{"b": "you"})),
			Expected:   ABenchStruct{B: "you"},                 //`"world"`,
			InExpected: BenchStruct{A: ABenchStruct{B: "you"}}, // `{"a":{"b":"world"}}`,
		},
		"not a pointer": {
			In:         struct{ Hello string }{Hello: "world"}, //`{"hello":"world"}`
			Op:         jq.Addition(map[string]string{"Hello": "you"}),
			Expected:   nil,
			InExpected: nil,
			HasError:   true,
		},
		"mapofpointer": {
			In:         &struct{ Hello []string }{Hello: []string{"world"}},
			Op:         jq.Chain(jq.Dot("Hello"), jq.Addition([]string{"hello"})),
			Expected:   []string{"world", "hello"},
			InExpected: struct{ Hello []string }{Hello: []string{"world", "hello"}},
		},
		"notafield": {
			In:         &struct{}{},
			Op:         jq.Chain(jq.Addition(map[string]string{"b": "hello"})),
			Expected:   struct{}{},
			InExpected: struct{}{},
			HasError:   true,
		},
		"othertoslice": {
			In:         &[]string{"a"},
			Op:         jq.Addition("b"),
			Expected:   []string{"a"},
			InExpected: []string{"a"},
			HasError:   true,
		},
		"filedothertype": {
			In:         &struct{ Hello string }{Hello: "world"},
			Op:         jq.Addition(map[string]int{"hello": 1}),
			Expected:   struct{ Hello string }{Hello: "world"},
			InExpected: struct{ Hello string }{Hello: "world"},
			HasError:   true,
		},
		"map": {
			In:         &map[string]int{},
			Op:         jq.Addition(map[string]int{"hello": 1}),
			Expected:   map[string]int{"hello": 1},
			InExpected: map[string]int{"hello": 1},
		},
		"mapint": {
			In:         &map[int]int{},
			Op:         jq.Addition(map[int]int{13: 1}),
			Expected:   map[int]int{13: 1},
			InExpected: map[int]int{13: 1},
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
