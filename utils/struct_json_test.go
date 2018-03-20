package utils_test

import (
	"reflect"
	"testing"

	"github.com/blippar/jq/utils"
)

type testStruct struct {
	A string `json:"field_1"`
	B string `json:"field_2"`
	C string `json:"field_3"`
	D string `json:"field_4,omitempty"`
	E string `json:"field_5,omitempty"`
	F string `json:"-"`
	G string `json:",omitempty"`
}

type testStructBis struct {
	A string `json:"field_4"`
	B string `json:"field_3"`
	C string `json:"field_2"`
	D string `json:"field_1,omitempty"`
	G string `json:"-,"` // key is `-` for this case and `json:"-,omitempty"`
}

var noCompilOptiField int
var noCompilOptiOk bool

func BenchmarkSet(t *testing.B) {
	for i := 0; i < t.N; i++ {
		idx, ok := utils.GetByJSONTag(reflect.ValueOf(testStruct{}), "field_1")
		noCompilOptiField, noCompilOptiOk = idx, ok
	}
}

func TestGetByJSONTag(t *testing.T) {
	// To be truly tested the tests should be run with -race
	t.Parallel()
	tests := map[string]struct {
		inRef      reflect.Value
		inField    string
		ExpectedOk bool
		Expected   int
	}{
		"get field A":                   {reflect.ValueOf(testStruct{}), "field_1", true, 0},
		"get field B":                   {reflect.ValueOf(testStruct{}), "field_2", true, 1},
		"get field C":                   {reflect.ValueOf(testStruct{}), "field_3", true, 2},
		"get field D":                   {reflect.ValueOf(testStruct{}), "field_4", true, 3},
		"get field E":                   {reflect.ValueOf(testStruct{}), "field_5", true, 4},
		"get another nonexisting filed": {reflect.ValueOf(testStruct{}), "field_6", false, 0},
		"not in package reflect":        {reflect.ValueOf(testStruct{}), "-", false, 0},

		"other struct get field A":        {reflect.ValueOf(testStructBis{}), "field_4", true, 0},
		"other struct get field B":        {reflect.ValueOf(testStructBis{}), "field_3", true, 1},
		"other struct get field C":        {reflect.ValueOf(testStructBis{}), "field_2", true, 2},
		"other struct get field D":        {reflect.ValueOf(testStructBis{}), "field_1", true, 3},
		"other struct in package reflect": {reflect.ValueOf(testStructBis{}), "-", true, 4},
	}
	for label, tt := range tests {
		t.Run(label, func(t *testing.T) {
			idx, ok := utils.GetByJSONTag(tt.inRef, tt.inField)
			if ok != tt.ExpectedOk {
				t.Errorf("GetByJSONTag() ok = %v, want %v", ok, tt.ExpectedOk)
			}
			if idx != tt.Expected {
				t.Errorf("GetByJSONTag() idx = %v, want %v", idx, tt.Expected)
			}
		})
	}
}
