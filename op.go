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
	"errors"
	"reflect"
	"strings"
)

// Op defines a single transformation to be applied to a []byte
type Op interface {
	Apply(interface{}) (interface{}, error)
}

// OpFunc provides a convenient func type wrapper on Op
type OpFunc func(interface{}) (interface{}, error)

// Apply executes the transformation defined by OpFunc
func (fn OpFunc) Apply(in interface{}) (interface{}, error) {
	return fn(in)
}

// Dot extract the specific key from the map provided; to extract a nested value, use the Dot Op in conjunction with the
// Chain Op
func Dot(key string) OpFunc {
	key = strings.TrimSpace(key)
	if key == "" {
		return func(in interface{}) (interface{}, error) { return in, nil }
	}
	key = strings.Title(key)
	return func(in interface{}) (rtn interface{}, err error) {
		if reflect.TypeOf(in).Kind() == reflect.Map {
			return reflect.ValueOf(in).MapIndex(reflect.ValueOf(key)).Interface(), nil
		}

		if v := reflect.ValueOf(in).FieldByName(key); v.Kind() != reflect.Invalid {
			defer func() {
				if r := recover(); r != nil {
					rtn = nil
					err = errors.New("panic :(")
				}
			}()
			rtn = v.Interface()
			return
		}
		return nil, errors.New("key not found")
	}
}

// Chain executes a series of operations in the order provided
func Chain(filters ...Op) OpFunc {
	return func(in interface{}) (interface{}, error) {
		if filters == nil {
			return in, nil
		}

		var err error
		data := in
		for _, filter := range filters {
			data, err = filter.Apply(data)
			if err != nil {
				return nil, err
			}
		}

		return data, nil
	}
}

// Index extracts a specific element from the array provided
func Index(index int) OpFunc {
	if index < 0 {
		return func(interface{}) (interface{}, error) {
			return nil, errors.New("Index needs to be supperior or equal to 0")
		}
	}
	return func(in interface{}) (interface{}, error) {
		if reflect.TypeOf(in).Kind() != reflect.Array && reflect.TypeOf(in).Kind() != reflect.Slice {
			return nil, errors.New("Not an array or a slice")
		}
		if reflect.ValueOf(in).Len() < index {
			return nil, errors.New("out of bound")
		}
		return reflect.ValueOf(in).Index(index).Interface(), nil
	}
}

// Range extracts a selection of elements from the array provided, inclusive
func Range(from, to int) OpFunc {
	if from < 0 {
		return func(interface{}) (interface{}, error) {
			return nil, errors.New("from needs to be supperior or equal to 0")
		}
	}
	if from > to {
		return func(interface{}) (interface{}, error) {
			return nil, errors.New("from needs to be inferior than to")
		}
	}

	return func(in interface{}) (interface{}, error) {
		if reflect.TypeOf(in).Kind() != reflect.Array && reflect.TypeOf(in).Kind() != reflect.Slice {
			return nil, errors.New("Not an array or a slice")
		}
		if reflect.ValueOf(in).Len() <= to {
			return nil, errors.New("out of bound")
		}

		slc := reflect.MakeSlice(reflect.TypeOf(in), 0, to-from)
		res := reflect.New(slc.Type())
		res.Elem().Set(slc)
		for i := from; i <= to; i++ {
			res.Elem().Set(
				reflect.Append(
					res.Elem(),
					reflect.ValueOf(in).Index(i),
				),
			)
		}
		return interface{}(res.Elem().Interface()), nil
	}
}
