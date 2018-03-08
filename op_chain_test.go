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
	"testing"

	"github.com/savaki/jq"
)

type BenchStruct struct {
	A ABenchStruct
}
type ABenchStruct struct {
	B string
}

func BenchmarkChain(t *testing.B) {
	op := jq.Chain(jq.Dot("A"), jq.Dot("B"))
	data := BenchStruct{
		A: ABenchStruct{
			B: "value",
		},
	}

	for i := 0; i < t.N; i++ {
		_, err := op.Apply(data)
		if err != nil {
			t.FailNow()
			return
		}
	}
}

func TestChain(t *testing.T) {
	testCases := map[string]struct {
		In       interface{}
		Op       jq.Op
		Expected string
		HasError bool
	}{
		"simple": {
			In:       struct{ Hello string }{Hello: "world"}, //`{"hello":"world"}`
			Op:       jq.Chain(jq.Dot("Hello")),
			Expected: "world", //`"world"`,
		},
		"nested": {
			In:       BenchStruct{A: ABenchStruct{B: "world"}}, // `{"a":{"b":"world"}}`,
			Op:       jq.Chain(jq.Dot("A"), jq.Dot("B")),
			Expected: "world", //`"world"`,
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			data, err := tc.Op.Apply(tc.In)
			if (err == nil) == tc.HasError {
				t.Errorf("Expected an error (%v) , got %v ", tc.HasError, err)
				t.FailNow()
			} else {
				if v, ok := data.(string); !ok || v != tc.Expected {
					t.Errorf("Expected %v (%T), got %v (%T)", tc.Expected, tc.Expected, data, data)
					t.FailNow()
				}
				if err != nil {
					t.Errorf("Expected %v (%T), got %v (%T)", tc.Expected, tc.Expected, data, data)
					t.FailNow()
				}
			}
		})
	}
}
