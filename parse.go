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
	"regexp"
	"strconv"
	"strings"
)

var (
	reArray = regexp.MustCompile(`^\s*\[\s*(\d+)(\s*:\s*(\d+))?\s*]\s*$`)
)

// Must is a convenience method similar to template.Must
func Must(op Op, err error) Op {
	if err != nil {
		panic(fmt.Errorf("unable to parse selector; %v", err.Error()))
	}

	return op
}

// see unicode.IsSpace
func isSpace(b byte) bool {
	switch b {
	// see unicode.isSpace
	case '\t', '\n', '\v', '\f', '\r', ' ', 0x85, 0xA0:
		return true
	}
	return false
}

func getIdxAfterJSON(s string, idx int) (int, error) {
	var (
		brace   int
		bracket int
		quotes  bool
	)

	for ; idx < len(s) && isSpace(s[idx]); idx++ {
	}

	for idx < len(s) {
		switch s[idx] {
		case '{':
			brace++
		case '"':
			quotes = !quotes
		case '[':
			bracket++
		}

		if brace == 0 && bracket == 0 && !quotes {
			idx++
			return idx, nil
		}
		if quotes && s[idx] == '\\' {
			idx += 2
			continue
		}

		switch s[idx] {
		case '}':
			brace--
		case ']':
			bracket--
		}
		idx++
	}
	if brace == 0 && bracket == 0 && !quotes {
		return idx, nil
	}
	return -1, errors.New("Not proper JSON")
}

func escapeJSON(selector string, args []interface{}) (string, []interface{}, error) {
	idx := 0
	argsIdx := 0
	selectors := []string{}

	for idx < len(selector) {
		for isSpace(selector[idx]) {
			idx++
		}

		if selector[idx] == '=' {
			idx++
			for isSpace(selector[idx]) {
				idx++
			}
			if selector[idx] == '%' &&
				(idx+1) < len(selector) && selector[idx+1] == 'v' {
				argsIdx++
				idx += 2
				continue
			} else {
				nIdx, err := getIdxAfterJSON(selector, idx)
				if err != nil {
					return "", nil, err
				}
				selectors = append(selectors, selector[:idx]+"%v")
				args = append(args, json.RawMessage(selector[idx:nIdx]))
				selector = selector[nIdx:]
				idx = 0
				continue
			}
		}
		idx++
	}
	return strings.Join(selectors, "=") + selector, args, nil
}

// Parse takes a string representation of a selector and returns the corresponding Op definition
// TODO: move the parsing logic to another package
func Parse(selector string, args ...interface{}) (Op, error) {
	var err error
	selector, args, err = escapeJSON(selector, args)
	if err != nil {
		return nil, err
	}
	segments := strings.Split(selector, ".")
	ops := make([]func(op ...Op) Op, 0, len(segments))
	for _, segment := range segments {
		var callAddition bool
		var callSet bool

		key := strings.TrimSpace(segment)
		if key == "" {
			continue
		}

		keys := strings.Split(key, "=")
		if len(keys) == 2 {
			key = strings.TrimSpace(keys[0])
			if strings.HasSuffix(key, "+") {
				key = strings.TrimSpace(strings.TrimSuffix(key, "+"))
				callAddition = true
			} else {
				callSet = true
			}
		} else if len(keys) > 2 {
			return nil, fmt.Errorf("Invalid argument: %v", keys)
		}

		if op, ok := parseArray(key); ok {
			ops = append(ops, op)
		} else {
			ops = append(ops, parseDot(key))
		}

		if callSet {
			if op, ok, err := parseSetAdd(keys[1], &args, parseSet); err != nil {
				return nil, err
			} else if ok {
				ops = append(ops, op)
				continue
			}
		} else if callAddition {
			if op, ok, err := parseSetAdd(keys[1], &args, parseAddition); err != nil {
				return nil, err
			} else if ok {
				ops = append(ops, op)
				continue
			}
		}
	}
	var f Op
	f = OpFunc(func(in reflect.Value) (reflect.Value, error) {
		return in, nil
	})
	for i := len(ops) - 1; i >= 0; i-- {
		f = ops[i](f)
	}
	return f, nil
}

func parseSetAdd(key string, args *[]interface{}, f func(string, interface{}) (func(op ...Op) Op, bool)) (func(op ...Op) Op, bool, error) {
	s := strings.TrimSpace(key)
	if s == "%v" {
		if len(*args) < 1 {
			return nil, false, errors.New("Not enough argument provided for addition function")
		}
		o, b := f(strings.TrimSpace(key), (*args)[0])
		*args = (*args)[1:]
		return o, b, nil
	}
	o, b := f(strings.TrimSpace(key), json.RawMessage(s))
	return o, b, nil
}

func parseSet(key string, arg interface{}) (func(op ...Op) Op, bool) {
	return func(op ...Op) Op {
		return Set(arg, op...)
	}, true
}

func parseAddition(key string, arg interface{}) (func(op ...Op) Op, bool) {
	return func(op ...Op) Op {
		return Addition(arg, op...)
	}, true
}

func parseArray(key string) (func(op ...Op) Op, bool) {
	match := reArray.FindAllStringSubmatch(key, -1)
	if len(match) != 1 {
		return nil, false
	}

	fromStr := match[0][1]
	from, err := strconv.Atoi(fromStr)
	if err != nil {
		return nil, false
	}

	toStr := match[0][3]
	if toStr == "" {
		return func(op ...Op) Op {
			return Index(from, op...)
		}, true
	}

	to, err := strconv.Atoi(toStr)
	if err != nil {
		return nil, false
	}

	return func(op ...Op) Op {
		return Range(from, to, op...)
	}, true
}

func parseDot(key string) func(...Op) Op {
	return func(op ...Op) Op {
		return Dot(key, op...)
	}
}
