package main

import "fmt"

func assertEqual(a, b interface{}) {
	if a != b {
		e1, _ := a.(error)
		e2, _ := b.(error)
		if e1 != nil && e2 != nil {
			if e1.Error() == e2.Error() {
				return
			}
		}
		panic(fmt.Sprintf("a:%#v != b:%#v", a, b))
	}
}

func assertNil(a interface{}) {
	if a != nil {
		panic(fmt.Sprintf("a:%v != nil", a))
	}
}
