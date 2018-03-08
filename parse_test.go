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
	"reflect"
	"testing"

	"github.com/savaki/jq"
)

func TestParse(t *testing.T) {
	testCases := map[string]struct {
		In       interface{}
		Op       string
		Expected interface{}
		HasError bool
	}{
		"simple": {
			In:       struct{ Hello string }{Hello: "world"}, // `{"hello":"world"}`,
			Op:       ".Hello",
			Expected: "world",
		},
		"lowercase": {
			In:       struct{ Hello string }{Hello: "world"}, // `{"hello":"world"}`,
			Op:       ".hello",
			Expected: "world",
		},
		"nested": {
			In:       struct{ A struct{ B string } }{A: struct{ B string }{"world"}}, //`{"a":{"b":"world"}}`,
			Op:       ".A.B",
			Expected: "world", // `"world"`
		},
		"index": {
			In:       []string{"a", "b", "c"}, //`["a","b","c"]`,
			Op:       ".[1]",
			Expected: "b", // `"b"`
		},
		"range": {
			In:       []string{"a", "b", "c"}, //`["a","b","c"]`,
			Op:       ".[1:2]",
			Expected: []string{"b", "c"}, //`["b","c"]`,
		},
		"nested index": {
			In: struct {
				Abc string
				Def []string
			}{Abc: "-", Def: []string{"a", "b", "c"}}, //`{"abc":"-","def":["a","b","c"]}`,
			Op:       ".Def.[1]",
			Expected: "b", //`"b"`,
		},
		"nested range": {
			In: struct {
				Abc string
				Def []string
			}{Abc: "-", Def: []string{"a", "b", "c"}}, //`{"abc":"-","def":["a","b","c"]}`,
			Op:       ".Def.[1:2]",
			Expected: []string{"b", "c"}, //`["b","c"]`,
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			op, err := jq.Parse(tc.Op)
			if err != nil {
				t.FailNow()
			}

			data, err := op.Apply(tc.In)
			if tc.HasError {
				if err == nil {
					t.Errorf("Expected an error got %v, %v", data, err)
					t.FailNow()
				}
			} else {
				if ty := reflect.TypeOf(data); ty != reflect.TypeOf(tc.Expected) {
					t.Errorf("ZExpected %v (%T), got %v (%T)", tc.Expected, tc.Expected, data, data)
					t.FailNow()
				}
				if reflect.TypeOf(data).Kind() == reflect.Slice {
					for i := 0; i < reflect.ValueOf(data).Len() && i < reflect.ValueOf(tc.Expected).Len(); i++ {
						if reflect.ValueOf(data).Index(i).Interface() != reflect.ValueOf(tc.Expected).Index(i).Interface() {
							t.Errorf("AExpected %v (%T), got %v (%T)", tc.Expected, tc.Expected, data, data)
							t.FailNow()
						}
					}
				}
				if err != nil {
					t.Errorf("EExpected no error got %v, %v", data, err)
					t.FailNow()
				}
			}
		})
	}
}
