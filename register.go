package main

import (
	"fmt"
	"reflect"
)

// Registration holds info to be registered and used within your code.
// Good place to store any information requied by the Registerable.
type Registration interface{}

// RegisterFn signature of a Registerable method.
type RegisterFn func() Registration

// used to detect return type of methods
var registerFnType RegisterFn = func() Registration {
	return nil
}

// Registerable make a method Registerable.
// Good place to store any dependencies for use within your Registerable methods.
type Registerable[H Registration] interface {
	Register(H)
}

// RegisterMethods Register methods on Registerable T to be ran.
// Uses reflection to find methods that return Registration H.
// That Registration is then passed to and called by Registerable T's Register method.
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
