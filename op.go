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
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/blippar/jq/utils"
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
	var err error
	for _, f := range chainFun {
		if v, err = f.Apply(v); err != nil {
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
			if mapVal.Kind() == reflect.Interface {
				mapVal = mapVal.Elem()
			}
			newMapVal := reflect.New(mapVal.Type())
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

			if idx, ok := utils.GetByJSONTag(in, key); ok {
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
	valRefTyp := valRef.Type()
	valRefKind := valRef.Kind()

	return func(in reflect.Value) (reflect.Value, error) {
		// check if in is a pointer/interface
		// when val is a pointer, set the value of the pointer
		// when val is a value, set the underlying value of the pointer
		in = reflect.Indirect(in)
		inTyp := in.Type()
		if !in.CanSet() {
			return reflect.Value{}, ErrCannotSet
		}

		switch in.Kind() {
		case reflect.Slice:
			if ok, res, err := valueFromJSONWithAlloc(val, inTyp); ok {
				if err != nil {
					return reflect.Value{}, err
				}
				valRef = res
				valRefTyp = inTyp
				valRefKind = valRef.Kind()
			}

			if valRefKind != reflect.Slice {
				return reflect.Value{}, fmt.Errorf("Addition: In is slice but val is: %v", valRefKind)
			}
			if inTyp != valRefTyp {
				return reflect.Value{}, fmt.Errorf("Addition: cannot add elem from slice of type %v to slice of type %v", valRefTyp, inTyp)
			}
			in.Set(reflect.AppendSlice(in, valRef))
			return callChain(in, chainFun...)
		case reflect.Map:
			if ok, res, err := valueFromJSONWithAlloc(val, inTyp); ok {
				if err != nil {
					return reflect.Value{}, err
				}
				valRef = res
				valRefTyp = valRef.Type()
				valRefKind = valRef.Kind()
			}

			if valRefKind != reflect.Map {
				return reflect.Value{}, fmt.Errorf("Addition: In is map but val is: %v", valRefKind)
			}
			for _, k := range valRef.MapKeys() {
				v := reflect.Indirect(valRef.MapIndex(k))
				in.SetMapIndex(k, v)
			}
			return callChain(in, chainFun...)
		case reflect.Struct:
			if v, ok := val.(json.RawMessage); ok {
				var buf map[string]json.RawMessage
				err := decodeJSON(v, &buf)
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
			if valRefKind != reflect.Map {
				return reflect.Value{}, fmt.Errorf("Addition: in is struct and val is not a map: %v", valRefKind)
			}
			if valRefTyp.Key().Kind() != reflect.String {
				return reflect.Value{}, fmt.Errorf("Addition: map of value to used with a struct is not a map with string keys: %v", valRefTyp.Key())
			}

			for _, k := range valRef.MapKeys() {
				// TODO: handle JSON values
				ks := strings.Title(k.String())

				v := reflect.Indirect(valRef.MapIndex(k))
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
		return reflect.Value{}, fmt.Errorf("Unsupported type (%v)", valRefTyp)
	}
}

// Set change the val parameter to the provided interface{} (map/slice/struct)
func Set(val interface{}, chainFun ...Op) OpFunc {
	valRef := reflect.ValueOf(val)
	valRefTyp := valRef.Type()

	return func(in reflect.Value) (reflect.Value, error) {
		in = reflect.Indirect(in)
		inTyp := in.Type()
		if !in.CanSet() {
			return reflect.Value{}, ErrCannotSet
		}

		inAddr := in.Addr()
		if ok, err := valueFromJSON(val, inAddr.Interface()); ok {
			if err != nil {
				return reflect.Value{}, err
			}
			return callChain(in, chainFun...)
		}

		if valRefTyp != inTyp {
			return reflect.Value{}, fmt.Errorf("Different type: %v, %v", valRefTyp, inTyp)
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
		in = reflect.Indirect(in)
		if in.Kind() != reflect.Array && in.Kind() != reflect.Slice {
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
		if k := in.Kind(); k != reflect.Array &&
			k != reflect.Slice &&
			k != reflect.String {
			return reflect.Value{}, errors.New("Not an array, a slice or a string")
		}
		if in.Len() <= to {
			return reflect.Value{}, errors.New("out of bound")
		}
		return callChain(in.Slice(from, to), chainFun...)
	}
}
