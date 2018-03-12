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

var noCompilOptiDot reflect.Value

func BenchmarkDot(t *testing.B) {
	op := jq.Dot("Hello")
	data := reflect.ValueOf(struct{ Hello string }{Hello: "world"}) // {"hello":"world"}

	for i := 0; i < t.N; i++ {
		rv, err := op.Apply(data)
		noCompilOptiDot = rv
		if err != nil {
			t.FailNow()
			return
		}
	}
}

func TestDot(t *testing.T) {
	testCases := map[string]struct {
		In       interface{}
		Key      string
		Expected interface{}
		HasError bool
	}{
		"simple": {
			In:       struct{ Hello string }{Hello: "world"}, //`{"hello":"world"}`,
			Key:      "Hello",
			Expected: `"world"`,
		},
		"key not found": {
			In:       struct{ Hello string }{Hello: "world"}, // `{"hello":"world"}`,
			Key:      "junk",
			HasError: true,
		},
		// "unclosed value": {
		// 	In:       `{"hello":"world`,
		// 	Key:      "hello",
		// 	HasError: true,
		// },
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			op := jq.Dot(tc.Key)
			data, err := op.Apply(reflect.ValueOf(tc.In))
			if tc.HasError {
				if err == nil {
					t.Errorf("Expected an error got %v, %v", data, err)
					t.FailNow()
				}
			} else {
				if reflect.TypeOf(data) == reflect.TypeOf(tc.Expected) && data == tc.Expected {
					t.Errorf("Expected %v (%T), got %v (%T)", tc.Expected, tc.Expected, data, data)
					t.FailNow()
				}
				if err != nil {
					t.Errorf("Expected no error got %v, %v", data, err)
					t.FailNow()
				}
			}
		})
	}
}
