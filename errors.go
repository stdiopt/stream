package stream

import (
	"errors"
	"fmt"
	"reflect"
)

var ErrBreak = errors.New("break")

type strmError struct {
	pname string
	err   error
}

func (e strmError) Unwrap() error {
	return e.err
}

func (e strmError) Error() string {
	return fmt.Sprintf("%s %v", e.pname, e.err)
}

type TypeMismatchError struct {
	want string
	got  string
}

func NewTypeMismatchError(w, g interface{}) TypeMismatchError {
	var ws string
	var gs string
	switch w := w.(type) {
	case reflect.Type:
		ws = w.String()
	default:
		ws = fmt.Sprintf("%T", w)
	}

	switch g := g.(type) {
	case reflect.Type:
		gs = g.String()
	default:
		gs = fmt.Sprintf("%T", g)
	}
	return TypeMismatchError{
		want: ws,
		got:  gs,
	}
}

func (e TypeMismatchError) Error() string {
	return fmt.Sprintf("invalid type, want '%v' but got '%v'", e.want, e.got)
}

func wrapStrmError(name string, err error) error {
	if err == nil {
		return nil
	}
	return strmError{name, err}
}

func wrapProcFunc(name string, fn ProcFunc) ProcFunc {
	return func(p Proc) error {
		return wrapStrmError(name, fn(p))
	}
}
