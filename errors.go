package stream

import (
	"fmt"
)

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

func (e TypeMismatchError) Error() string {
	return fmt.Sprintf("invalid type, want '%v' but got '%v'", e.want, e.got)
}

func wrapStrmError(name string, err error) error {
	if err == nil {
		return nil
	}
	return strmError{name, err}
}

func wrapProcFunc(name string, fn procFunc) procFunc {
	return func(p Proc) error {
		return wrapStrmError(name, fn(p))
	}
}
