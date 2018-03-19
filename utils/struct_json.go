// Package utils exposes functions that might be useful as standalone.
package utils

import (
	"reflect"
	"strings"
	"sync"
)

// jsonNameToIdx is a map used to cache json tag lookup.
var jsonNameToIdx = sync.Map{} // map[reflect.Type]map[string]int{}

// GetByJSONTag takes a reflect value and a string, it will try to get
// the field index by searching for the name in the json part of the struct tag.
// Return (0, false) if no field has a json tag name matching the field.
// We use a map to cache type already looked up. (safe for concurency)
func GetByJSONTag(v reflect.Value, field string) (int, bool) {
	v = reflect.Indirect(v)
	if v.Kind() != reflect.Struct {
		return 0, false
	}

	vType := v.Type()
	buf, ok := jsonNameToIdx.Load(vType)

	// When it is the first time
	if !ok {
		bufMap := map[string]int{}
		// For each field get the json tag and store the
		// first of the element (the name) in the map
		for i := 0; i < vType.NumField(); i++ {
			f := vType.Field(i)
			t := f.Tag.Get("json")
			strs := strings.Split(t, ",")
			if (strs[0] != "" && strs[0] != "-") || (strs[0] == "-" && len(strs) > 1) {
				bufMap[strs[0]] = i
			}
		}
		buf, _ = jsonNameToIdx.LoadOrStore(vType, bufMap)
	}

	// When Cached
	fieldM := buf.(map[string]int)
	val, ok := fieldM[field]
	return val, ok
}
