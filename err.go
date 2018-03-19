package jq

import "errors"

var (
	// ErrCannotSet indicates the value is not settable, it will happend
	// when a value is passed by copy
	ErrCannotSet = errors.New("Cannot set value")
	// ErrUnknownField indicates that we try to access a field not present in
	// a struct
)
