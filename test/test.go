package main

import (
	"fmt"

	"github.com/savaki/jq"
)

type testSub struct {
	S []string
}

type test struct {
	A     string
	Slice testSub
}

func main() {
	op, err := jq.Parse(".Slice.S.[0:2]") // create an Op
	if err != nil {
		panic(err)
	}
	v := test{
		A: "coucou",
		Slice: testSub{
			S: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
		},
	}
	value, err := op.Apply(v) // value == '"world"'
	if err != nil {
		panic(err)
	}
	fmt.Println(value)
}
