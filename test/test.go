package main

import (
	"fmt"
	"log"
	"reflect"

	"github.com/savaki/jq"
)

type testSub struct {
	S []string
}

type test struct {
	A     string
	Slice testSub
	B     map[string]testSub
}

func main() {
	op, err := jq.Parse(". += %v", map[string]interface{}{"A": "wesh"}) // create an Op
	if err != nil {
		panic(err)
	}
	v := test{
		A: "coucou",
		Slice: testSub{
			S: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
		},
		B: map[string]testSub{
			"coucou": testSub{
				S: []string{"0"},
			},
		},
	}
	value, err := op.Apply(reflect.ValueOf(&v))
	if err != nil {
		panic(err)
	}
	log.Printf("value: %v, err %v\n\n", v, err)

	value, err = jq.Chain(jq.Dot("B"), jq.Addition(map[string]testSub{
		"coucou": testSub{
			S: []string{"11"},
		},
	})).Apply(reflect.ValueOf(&v))
	if err != nil {
		panic(err)
	}
	fmt.Println(value.Interface())
}
