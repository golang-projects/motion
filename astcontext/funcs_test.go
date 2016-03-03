package astcontext

import (
	"fmt"
	"testing"
)

func TestEnclosingFunc(t *testing.T) {
	var src = `package main

var bar = func() {}

func foo() error {
	_ = func() {
		// -------
	}
	return nil
}
`

	testPos := []struct {
		offset     int
		funcOffset int
	}{
		{25, 24},
		{32, 24}, // var bar = func {}
		{35, 35}, // func foo() error {
		{53, 35}, // func foo() error {
		{67, 59}, // _ = func() {
		{70, 59}, // _ = func() {
		{85, 35}, // func foo() error {
		{96, 35}, // func foo() error {
	}

	parser, err := NewParser().SetOptions(nil).ParseSrc([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	funcs := parser.Funcs()

	for _, pos := range testPos {
		fn, err := funcs.EnclosingFunc(pos.offset)
		if err != nil {
			fmt.Printf("err = %+v\n", err)
			continue
		}

		if fn.FuncPos.Offset != pos.funcOffset {
			t.Errorf("offset %d should belong to func with offset: %d, got: %d",
				pos.offset, pos.funcOffset, fn.FuncPos.Offset)
		}
	}
}

func TestNextFuncComment(t *testing.T) {
	var src = `package main

// Comment foo
// Comment bar
func foo() error {
	_ = func() {
		// -------
	}
	return nil
}

func bar() error {
	return nil
}`

	testPos := []struct {
		start int
		want  int
	}{
		{start: 14, want: 108},
		{start: 29, want: 108},
	}

	opts := &ParserOptions{ParseComments: true}
	parser, err := NewParser().SetOptions(opts).ParseSrc([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	funcs := parser.Funcs().Declarations()

	for _, pos := range testPos {
		fn, _ := funcs.NextFunc(pos.start)

		if fn.FuncPos.Offset != pos.want {
			t.Errorf("offset %d should pick func with offset: %d, got: %d",
				pos.start, pos.want, fn.FuncPos.Offset)
		}
	}
}

func TestFunc_Signature(t *testing.T) {
	var src = `package main

var a = func() { fmt.Println("tokyo") }

func foo(
	a int,
	b string,
	c bool,
) (
	bool,
	error,
) {
	return false, nil
}

func foo(a, b int, foo string) (string, error) {
	_ = func() {
		// -------
	}
	return nil
}

func (q *qaz) example(x,y,z int) error {
	_ = func(foo int) error {
		return nil
	}
}

func example() {}

func variadic(x ...string) {}

func bar(x int) error {
	return nil
}`

	testFuncs := []struct {
		want string
	}{
		{want: "func()"},
		{want: "func foo(a int, b string, c bool) (bool, error)"},
		{want: "func foo(a, b int, foo string) (string, error)"},
		{want: "func()"},
		{want: "func (q *qaz) example(x, y, z int) error"},
		{want: "func(foo int) error"},
		{want: "func example()"},
		{want: "func variadic(x ...string)"},
		{want: "func bar(x int) error"},
	}

	parser, err := NewParser().ParseSrc([]byte(src))
	if err != nil {
		t.Fatal(err)
	}

	funcs := parser.Funcs()

	for i, fn := range funcs {
		fmt.Printf("[%d] %s\n", i, fn.Signature)
		if fn.Signature != testFuncs[i].want {
			t.Errorf("function signatures\n\twant: %s\n\tgot : %s",
				testFuncs[i].want, fn.Signature)
		}
	}
}
