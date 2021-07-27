package stream

import (
	"errors"
	"fmt"
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
