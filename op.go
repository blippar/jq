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

package jq

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Op defines a single transformation to be applied to a []byte
type Op interface {
	Apply(reflect.Value) (reflect.Value, error)
}

// OpFunc provides a convenient func type wrapper on Op
type OpFunc func(reflect.Value) (reflect.Value, error)

// Apply executes the transformation defined by OpFunc
func (fn OpFunc) Apply(in reflect.Value) (reflect.Value, error) {
	return fn(in)
}

func callChain(v reflect.Value, chainFun ...Op) (reflect.Value, error) {
	for _, f := range chainFun {
		var err error
		v, err = f.Apply(v)
		if err != nil {
			return v, err
		}
	}
	return v, nil
}

// Dot extract the specific key from the map provided; to extract a nested value, use the Dot Op in conjunction with the
// Chain Op
func Dot(key string, chainFun ...Op) OpFunc {
	key = strings.TrimSpace(key)
	if key == "" {
		return func(in reflect.Value) (reflect.Value, error) { return callChain(in, chainFun...) }
	}

	return func(in reflect.Value) (reflect.Value, error) {
		for in.Kind() == reflect.Interface || in.Kind() == reflect.Ptr {
			in = in.Elem()
		}
		switch in.Kind() {
		case reflect.Map:
			var err error
			mapVal := in.MapIndex(reflect.ValueOf(key))
			newMapVal := reflect.New(in.MapIndex(reflect.ValueOf(key)).Type())
			newMapVal.Elem().Set(mapVal)
			mapVal = newMapVal.Elem()
			mapVal, err = callChain(mapVal, chainFun...)
			if err != nil {
				return reflect.Value{}, err
			}
			in.SetMapIndex(reflect.ValueOf(key), mapVal)
			return mapVal, nil
		case reflect.Struct:
			var r reflect.Value

			if idx, ok := getJSONTag(in, key); ok {
				r = in.Field(idx)
			} else {
				r = in.FieldByName(strings.Title(key))
			}
			if r.Kind() == reflect.Invalid {
				break
			}
			return callChain(r, chainFun...)
		case reflect.Slice:
			return reflect.Value{}, errors.New("cannot access name field on slice")
		}
		return reflect.Value{}, errors.New("key not found")
	}
}

// Addition adds the val parameter to the provided interface{} (map/slice/struct)
func Addition(val interface{}, chainFun ...Op) OpFunc {
	valRef := reflect.ValueOf(val)

	return func(in reflect.Value) (reflect.Value, error) {
		// check if in is a pointer/interface
		// when val is a pointer, set the value of the pointer
		// when val is a value, set the underlying value of the pointer
		if (in.Kind() == reflect.Interface || in.Kind() == reflect.Ptr) &&
			(valRef.Kind() != reflect.Interface && valRef.Kind() != reflect.Ptr) {
			in = in.Elem()
		}
		if !in.CanSet() {
			return reflect.Value{}, ErrCannotSet
		}

		switch in.Kind() {
		case reflect.Slice:
			if v, ok := val.(json.RawMessage); ok {
				slcPtr := reflect.New(in.Type())
				d := json.NewDecoder(bytes.NewReader(v))
				d.DisallowUnknownFields()
				err := d.Decode(slcPtr.Interface())
				if err != nil {
					return reflect.Value{}, err
				}
				valRef = slcPtr.Elem()
			}

			if valRef.Kind() != reflect.Slice {
				return reflect.Value{}, fmt.Errorf("Addition: In is slice but val is: %v", valRef.Kind())
			}
			if in.Type() != valRef.Type() {
				return reflect.Value{}, fmt.Errorf("Addition: cannot add elem from slice of type %v to slice of type %v", valRef.Type(), in.Type())
			}
			in.Set(reflect.AppendSlice(in, valRef))
			return callChain(in, chainFun...)
		case reflect.Map:
			if v, ok := val.(json.RawMessage); ok {
				slcPtr := reflect.New(in.Type())
				d := json.NewDecoder(bytes.NewReader(v))
				d.DisallowUnknownFields()
				err := d.Decode(slcPtr.Interface())
				if err != nil {
					return reflect.Value{}, err
				}
				valRef = slcPtr.Elem()
			}

			if valRef.Kind() != reflect.Map {
				return reflect.Value{}, fmt.Errorf("Addition: In is map but val is: %v", valRef.Kind())
			}
			for _, k := range valRef.MapKeys() {
				v := valRef.MapIndex(k)
				if (v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr) &&
					(in.Kind() != reflect.Interface && in.Kind() != reflect.Ptr) {
					v = v.Elem()
				}
				in.SetMapIndex(k, v)
			}
			return callChain(in, chainFun...)
		case reflect.Struct:

			if v, ok := val.(json.RawMessage); ok {
				var buf map[string]json.RawMessage
				d := json.NewDecoder(bytes.NewReader(v))
				d.DisallowUnknownFields()
				err := d.Decode(&buf)
				if err != nil {
					return reflect.Value{}, err
				}

				for k, v := range buf {
					rv, err := Dot(k, Set(v))(in)
					if err != nil {
						return rv, err
					}
				}
				return callChain(in, chainFun...)
			}

			// TODO: handle struct to add values to another struct
			if valRef.Kind() != reflect.Map {
				return reflect.Value{}, fmt.Errorf("Addition: in is struct and val is not a map: %v", valRef.Kind())
			}
			if valRef.Type().Key().Kind() != reflect.String {
				return reflect.Value{}, fmt.Errorf("Addition: map of value to used with a struct is not a map with string keys: %v", valRef.Type().Key())
			}

			for _, k := range valRef.MapKeys() {
				// TODO: handle JSON values
				ks := strings.Title(k.String())

				v := valRef.MapIndex(k)
				if (v.Type().Kind() == reflect.Interface || v.Type().Kind() == reflect.Ptr) &&
					(valRef.Kind() != reflect.Interface && valRef.Kind() != reflect.Ptr) {
					v = v.Elem()
				}

				fieldRef := in.FieldByName(ks)
				if fieldRef.Kind() == reflect.Invalid {
					return reflect.Value{}, fmt.Errorf("Addition: Field \"%v\" does not exist", ks)
				}
				if v.Type() != fieldRef.Type() {
					return reflect.Value{}, fmt.Errorf("Addition: cannot set type %v in the field %s of type %v", v.Type(), k, fieldRef.Type())
				}
				in.FieldByName(ks).Set(v)
			}
			return callChain(in, chainFun...)
		}
		return reflect.Value{}, fmt.Errorf("Unsupported type (%v)", valRef.Type())
	}
}

// Set change the val parameter to the provided interface{} (map/slice/struct)
func Set(val interface{}, chainFun ...Op) OpFunc {
	valRef := reflect.ValueOf(val)

	return func(in reflect.Value) (reflect.Value, error) {
		if (in.Kind() == reflect.Interface || in.Kind() == reflect.Ptr) &&
			(valRef.Kind() != reflect.Interface && valRef.Kind() != reflect.Ptr) {
			in = in.Elem()
		}
		if !in.CanSet() {
			return reflect.Value{}, ErrCannotSet
		}

		// special case to unmashall json
		if v, ok := val.(json.RawMessage); ok {
			d := json.NewDecoder(bytes.NewReader(v))
			d.DisallowUnknownFields()
			err := d.Decode(in.Addr().Interface())
			if err != nil {
				return reflect.Value{}, err
			}
			return callChain(in, chainFun...)
		}

		if valRef.Type() != in.Type() {
			return reflect.Value{}, fmt.Errorf("Different type: %v, %v", valRef.Type(), in.Type())
		}
		in.Set(valRef)
		return callChain(in, chainFun...)
	}
}

// Chain executes a series of operations in the order provided
func Chain(filters ...OpFunc) OpFunc {
	return func(in reflect.Value) (reflect.Value, error) {
		if filters == nil {
			return in, nil
		}

		var err error
		data := in
		for _, filter := range filters {
			data, err = filter.Apply(data)
			if err != nil {
				return reflect.Value{}, err
			}
		}

		return data, nil
	}
}

// Index extracts a specific element from the array provided
func Index(index int, chainFun ...Op) OpFunc {
	if index < 0 {
		return func(reflect.Value) (reflect.Value, error) {
			return reflect.Value{}, errors.New("Index needs to be supperior or equal to 0")
		}
	}
	return func(in reflect.Value) (reflect.Value, error) {
		if in.Kind() == reflect.Interface || in.Kind() == reflect.Ptr {
			in = in.Elem()
		}
		if in.Type().Kind() != reflect.Array && in.Type().Kind() != reflect.Slice {
			return reflect.Value{}, errors.New("Not an array or a slice")
		}
		if in.Len() < index {
			return reflect.Value{}, errors.New("out of bound")
		}
		return callChain(in.Index(index), chainFun...)
	}
}

// Range extracts a selection of elements from the array provided, inclusive
func Range(from, to int, chainFun ...Op) OpFunc {
	if from < 0 {
		return func(reflect.Value) (reflect.Value, error) {
			return reflect.Value{}, errors.New("from needs to be supperior or equal to 0")
		}
	}
	if from > to {
		return func(reflect.Value) (reflect.Value, error) {
			return reflect.Value{}, errors.New("from needs to be inferior than to")
		}
	}

	return func(in reflect.Value) (reflect.Value, error) {
		if in.Kind() != reflect.Array &&
			in.Kind() != reflect.Slice &&
			in.Kind() != reflect.String {
			return reflect.Value{}, errors.New("Not an array, a slice or a string")
		}
		if in.Len() <= to {
			return reflect.Value{}, errors.New("out of bound")
		}
		return callChain(in.Slice(from, to), chainFun...)
	}
}
