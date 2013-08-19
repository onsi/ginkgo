package gospec_tests

import (
	. "github.com/onsi/godescribe"
	// "fmt"
	// "reflect"
	"testing"
)

func Test(t *testing.T) {
	RunSpecs(t, "GoDescribe's Example Suite")
}

// func TestReflection(t *testing.T) {
// 	a := func() {
// 		println("A")
// 	}

// 	b := func(c chan string) {
// 		println(<-c)
// 	}

// 	fmt.Println("=============================")

// 	fmt.Println(reflect.TypeOf(a), reflect.TypeOf(b))
// 	fmt.Println(reflect.ValueOf(a).Type(), reflect.ValueOf(b).Type())
// 	fmt.Println(reflect.TypeOf(a).Kind(), reflect.TypeOf(b).Kind())

// 	fmt.Println(reflect.TypeOf(a).NumIn(), reflect.TypeOf(b).NumIn())
// 	fmt.Println(reflect.TypeOf(b).In(0).Kind() == reflect.Chan)
// 	fmt.Println(reflect.TypeOf(b).In(0).Elem())

// 	var f func(chan string)

// 	fmt.Println(reflect.TypeOf(b) == reflect.TypeOf(f))
// }
