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

	"github.com/blippar/jq"
)

var noCompilOptiParse jq.Op

func BenchmarkParse(t *testing.B) {
	for i := 0; i < t.N; i++ {
		op, err := jq.Parse(".[1].+= \t\n{\"hello\":[\"\\\"w.o.r.l.d\\\"\"],\"field2\":\"val2=+=+=\"}")
		noCompilOptiParse = op
		if err != nil {
			t.Errorf("%v", err)
			t.FailNow()
			return
		}
	}
}

func TestParse(t *testing.T) {
	testCases := map[string]struct {
		In            interface{}
		Op            string
		OpArg         []interface{}
		Expected      interface{}
		InExpected    interface{}
		HasError      bool
		HasParseError bool
	}{
		"simple": {
			In:         struct{ Hello string }{Hello: "world"}, // `{"hello":"world"}`,
			Op:         ".Hello",
			Expected:   "world",
			InExpected: struct{ Hello string }{Hello: "world"},
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
		"addition nested": {
			In: &struct {
				Abc string
				Def []string
			}{Abc: "-", Def: []string{"a", "b", "c"}},
			Op:       ".Def+=%v",
			OpArg:    []interface{}{[]string{"d"}},
			Expected: []string{"a", "b", "c", "d"},
		},
		"set nested": {
			In: &struct {
				Abc string
				Def []string
			}{Abc: "-", Def: []string{"a", "b", "c"}},
			Op:       ".Def=%v",
			OpArg:    []interface{}{[]string{"d"}},
			Expected: []string{"d"},
		},
		"set no args": {
			In: &struct {
				Abc string
				Def []string
			}{Abc: "-", Def: []string{"a", "b", "c"}},
			Op:            ".Def=%v",
			OpArg:         []interface{}{},
			Expected:      nil,
			HasParseError: true,
		},
		"addition no args": {
			In: &struct {
				Abc string
				Def []string
			}{Abc: "-", Def: []string{"a", "b", "c"}},
			Op:            ".Def+=%v",
			OpArg:         []interface{}{},
			Expected:      nil,
			HasParseError: true,
		},
		"set nested json": {
			In: &struct {
				Abc string
				Def []string
			}{Abc: "-", Def: []string{"a", "b", "c"}},
			Op:       `.Def=["d"]`,
			OpArg:    []interface{}{},
			Expected: []string{"d"},
			InExpected: struct {
				Abc string
				Def []string
			}{Abc: "-", Def: []string{"d"}},
		},
		"add nested slice json": {
			In: &struct {
				Abc string
				Def []string
			}{Abc: "-", Def: []string{"a", "b", "c"}},
			Op:       `.Def+=["d"]`,
			OpArg:    []interface{}{},
			Expected: []string{"a", "b", "c", "d"},
			InExpected: struct {
				Abc string
				Def []string
			}{Abc: "-", Def: []string{"a", "b", "c", "d"}},
		},
		"add nested map json": {
			In: &struct {
				Abc string
				Def map[string]string
			}{Abc: "-", Def: map[string]string{"a": "a", "b": "b", "c": "c"}},
			Op:       `.Def+={"d":"d"}`,
			OpArg:    []interface{}{},
			Expected: map[string]string{"a": "a", "b": "b", "c": "c", "d": "d"},
			InExpected: struct {
				Abc string
				Def map[string]string
			}{Abc: "-", Def: map[string]string{"a": "a", "b": "b", "c": "c", "d": "d"}},
		},
		"add nested struct json": {
			In: &struct {
				Abc string
				Def struct {
					A string
					B string
					C string
					D string
				}
			}{Abc: "-", Def: struct {
				A string
				B string
				C string
				D string
			}{A: "a", B: "b", C: "c"}},
			Op:    `.Def+={"d":"d"}`,
			OpArg: []interface{}{},
			Expected: struct {
				A string
				B string
				C string
				D string
			}{A: "a", B: "b", C: "c", D: "d"},
			InExpected: struct {
				Abc string
				Def struct {
					A string
					B string
					C string
					D string
				}
			}{Abc: "-", Def: struct {
				A string
				B string
				C string
				D string
			}{A: "a", B: "b", C: "c", D: "d"}},
		},
		"json struct tag json": {
			In: &struct {
				Abc string
				def int
				Def struct {
					N int `json:"world"`
				} `json:"hello"`
			}{Abc: "-", def: 1, Def: struct {
				N int `json:"world"`
			}{42}},
			Op:       `.hello.world`,
			OpArg:    []interface{}{},
			Expected: 42,
			InExpected: struct {
				Abc string
				def int
				Def struct {
					N int `json:"world"`
				} `json:"hello"`
			}{Abc: "-", def: 1, Def: struct {
				N int `json:"world"`
			}{42}},
		},
		"parse string json set with dot": {
			In:         &[]string{"a", "b", "c"},
			Op:         `.[1]="a.b.c.d"`,
			Expected:   "a.b.c.d",
			InExpected: []string{"a", "a.b.c.d", "c"},
		},
		"parse array add json": {
			In:         &[]string{"a", "b", "c"},
			Op:         `.+=["a.b.c.d"]`,
			Expected:   []string{"a", "b", "c", "a.b.c.d"},
			InExpected: []string{"a", "b", "c", "a.b.c.d"},
		},
		"parse plus equal string": {
			In: &[]map[string]interface{}{
				map[string]interface{}{"a": "b"},
				map[string]interface{}{"field0": "val0"},
			},
			Op: ".[ 1 ] .  += \t\n{   \"hello\" \t \n : [ \"\\\"w.o.r.l.d\\\"\"],\n\t\r\"field2\":\"val2=+=+=\"}",
			Expected: map[string]interface{}{
				"field0": "val0",
				"field2": "val2=+=+=",
				"hello":  []interface{}{"\"w.o.r.l.d\""},
			},
			InExpected: []map[string]interface{}{
				map[string]interface{}{"a": "b"},
				map[string]interface{}{
					"field0": "val0",
					"field2": "val2=+=+=",
					"hello":  []interface{}{"\"w.o.r.l.d\""},
				},
			},
		},
		"can modify value in map": {
			In: &map[string][]string{
				"hello": []string{"hello"},
			},
			Op:       ".hello+=[\"world\"]",
			Expected: []string{"hello"},
			InExpected: map[string][]string{
				"hello": []string{"hello", "world"},
			},
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			op, err := jq.Parse(tc.Op, tc.OpArg...)
			if tc.HasParseError {
				if err == nil {
					t.Errorf("Expected an error got %v, %v", op, err)
					t.FailNow()
				}
				t.Skip()
			} else {
				if err != nil {
					t.Errorf("Expected no error got %v, %v", op, err)
					t.FailNow()
				}
			}

			data, err := op.Apply(reflect.ValueOf(tc.In))
			if tc.HasError {
				if err == nil {
					t.Errorf("Expected an error got %v, %v", data, err)
					t.FailNow()
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error got %v, %v", data, err)
					t.FailNow()
				}
				if data.Kind() == reflect.Invalid {
					t.Errorf("Expected %v (%T), got %v (%T)", tc.Expected, tc.Expected, data, data)
					t.FailNow()
				}
				if data.Type() != reflect.TypeOf(tc.Expected) {
					t.Errorf("Expected %v (%T), got %v (%T)", tc.Expected, tc.Expected, data, data)
					t.FailNow()
				}
				if data.Kind() == reflect.Slice {
					for i := 0; i < data.Len() && i < reflect.ValueOf(tc.Expected).Len(); i++ {
						if data.Index(i).Interface() != reflect.ValueOf(tc.Expected).Index(i).Interface() {
							t.Errorf("Expected %v (%T), got %v (%T)", tc.Expected, tc.Expected, data, data)
							t.FailNow()
						}
					}
				}
				if tc.InExpected != nil {
					in := reflect.ValueOf(tc.In)
					if in.Kind() == reflect.Ptr {
						if !reflect.DeepEqual(tc.InExpected, in.Elem().Interface()) {
							t.Errorf("Expected %v (%T), got %v (%T)", tc.InExpected, tc.InExpected, in.Elem().Interface(), in.Elem().Interface())
							t.FailNow()
						}
					} else {
						if !reflect.DeepEqual(tc.InExpected, tc.In) {
							t.Errorf("Expected %v (%T), got %v (%T)", tc.InExpected, tc.InExpected, tc.In, tc.In)
							t.FailNow()
						}
					}
				}
			}
		})
	}
}
