package jq

import (
	"reflect"
	"strings"
)

var structToIdx map[string]map[string]jsonToIdx

type jsonToIdx struct {
	ok  bool
	val int
}

func init() {
	structToIdx = make(map[string]map[string]jsonToIdx)
}

func getJsonTag(v reflect.Value, field string) (int, bool) {
	if v.Kind() == reflect.Invalid || v.Kind() != reflect.Struct {
		return 0, false
	}

	if structToIdx[v.Type().Name()] == nil || v.Type().Name() == "" {
		structToIdx[v.Type().Name()] = make(map[string]jsonToIdx)
		for i := 0; i < v.Type().NumField(); i++ {
			f := v.Type().Field(i)
			t := f.Tag.Get("json")
			strs := strings.Split(t, ",")
			if strs[0] != "" {
				structToIdx[v.Type().Name()][strs[0]] = jsonToIdx{true, i}
			}
		}
	}

	r := structToIdx[v.Type().Name()][field]
	return r.val, r.ok
}
