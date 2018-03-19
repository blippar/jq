package jq

import (
	"reflect"
	"strings"
	"sync"
)

var jsonNameToIdx = sync.Map{} // map[reflect.Type]map[string]int{}

func getJSONTag(v reflect.Value, field string) (int, bool) {
	v = reflect.Indirect(v)
	if v.Kind() == reflect.Invalid || v.Kind() != reflect.Struct {
		return 0, false
	}

	vType := v.Type()
	// When it is the first time
	if _, ok := jsonNameToIdx.Load(vType); !ok {
		buf := map[string]int{}
		// For each field get the json tag and store the
		// first of the element (the name) in the map
		for i := 0; i < vType.NumField(); i++ {
			f := vType.Field(i)
			t := f.Tag.Get("json")
			strs := strings.Split(t, ",")
			if strs[0] != "" {
				buf[strs[0]] = i
			}
		}
		jsonNameToIdx.LoadOrStore(vType, buf)
	}
	buf, _ := jsonNameToIdx.Load(vType)
	fieldM := buf.(map[string]int)
	val, ok := fieldM[field]
	return val, ok
}
