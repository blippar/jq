# jq

Based on https://github.com/savaki/jq

Use `reflect` package and jq syntax to modify golang struct


## Example

```go
package main

import (
	"fmt"
	"reflect"

	"github.com/blippar/jq"
)

func main() {
	data := struct {
		A map[string][]int `json:"field"`
	}{
		A: map[string][]int{
			"here": []int{1},
		},
	}
  
	// create an Op
	// op := Dot("field", Dot("here", Addition(json.RawMessage("[3]"))))
	op, _ := jq.Parse(".field.here+=[3]")
  
	// Apply the op to data
	value, _ := op.Apply(reflect.ValueOf(&data))
	fmt.Printf("data: %v\n", data)
	fmt.Printf("value: %v\n", value)
}
// value is of type reflect.Value,
// value.Interface() will return underlying type
//
// output:
// data: {map[here:[1 3]]}
// value: [1 3]
```

## Syntax

The initial goal is to support all the selectors the original jq command line supports.

| syntax | meaning|
| :--- | :--- |
| . |  unchanged input |
| .foo |  value at key |
| .foo.bar |  value at nested key |
| .[0] | value at specified element of array | 
| .[0:1] | array of specified elements of array, inclusive |
| .foo.[0] | nested value |
| .= | set value |
| .+= | add value to string and slice |

## Examples

### Data
```go
struct {
  String string
  Number float32
  Simple []string
  Mixed  []interface{}
  Struct struct {
    A string
    B []int
  }
  Map         map[string]interface{}
  WithJSONTag int `json:"tagged"`
}{
  String: "a",
  Number: 1.23,
  Simple: []string{"a", "b", "c"},
  Mixed: []interface{}{
    "d",
    2,
    map[string]string{"hello": "world"},
  },
  Struct: struct {
    A string
    B []int
  }{
    A: "e",
    B: []int{3, 4, 5},
  },
  Map:         map[string]interface{}{"f": []int{6, 7, 8}},
  WithJSONTag: 9,
}
```

| syntax | value |
| :--- | :--- |
| .string | "a" |
| .number | 1.23 |
| .simple | ["a", "b", "c"] |
| .simple.[0] | "a" |
| .simple = ["d"] | ["d"] |
| .simple += ["d"] | ["a", "b", "c", "d"] |
| .simple = "d" | \<Error\> |
| .simple.[0:1] | ["a","b"] |
| .mixed.[2].hello | "world" |
| .Struct.a | "e" |
| .tagged | 9 |

## Addition and Set

For these two, the parameter reflect.Value can be set ( `val.CanSet() == true` )
or it is an interface / ptr, result of Elem can be set  ( `val.Elem().CanSet() == true` )

Set `=` will assign the value to another, if they are not of the same type it will return an error.

Addition `+=` will concat the value to the one provided, if they are not of the same type it will return an error

