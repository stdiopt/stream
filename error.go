package stream

import (
	"errors"
	"fmt"
)

var ErrBreak = errors.New("break")

type strmError struct {
	name string
	file string
	line int
	err  error
}

func (e strmError) Unwrap() error {
	return e.err
}

func (e strmError) Error() string {
	return fmt.Sprintf("%s:%d [%s] %v",
		e.file,
		e.line,
		e.name,
		e.err,
	)
}
