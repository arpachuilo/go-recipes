package main

import (
	"fmt"
	"reflect"
)

type Registration interface{}
type RegisterFn func() Registration

var registerFnType RegisterFn = func() Registration {
	return nil
}

type Registerable[H Registration] interface {
	Register(H)
}

func RegisterMethods[H Registration, T Registerable[H]](registerable T) {
	st := reflect.TypeOf(registerable)
	sv := reflect.ValueOf(registerable)
	for i := 0; i < st.NumMethod(); i++ {
		mtype := st.Method(i)
		mvalue := sv.Method(i)

		register := mvalue.Type().AssignableTo(reflect.TypeOf(registerFnType))
		if register {
			fmt.Println("registering", mtype.Name)
			sh := mtype.Func.Call([]reflect.Value{sv})
			registerable.Register(sh[0].Interface().(H))
		}
	}
}
