package jq

import (
	"encoding/json"
	"reflect"
)

func valueFromJSONWithAlloc(raw interface{}, typ reflect.Type) (bool, reflect.Value, error) {
	v, ok := raw.(json.RawMessage)
	if !ok {
		return false, reflect.Value{}, nil
	}
	// Need to allocate
	to := reflect.New(typ)

	err := decodeJSON(v, to.Interface())
	if err != nil {
		return true, reflect.Value{}, err
	}
	return true, to.Elem(), nil
}

func valueFromJSON(raw, to interface{}) (bool, error) {
	in, ok := raw.(json.RawMessage)
	if !ok {
		return false, nil
	}
	return true, decodeJSON(in, to)
}
