package stream

import (
	"fmt"
	"path/filepath"
	"runtime"
)

func procName() string {
	var name string
	{
		pc, _, _, _ := runtime.Caller(2)
		fi := runtime.FuncForPC(pc)
		name = fi.Name()
		_, name = filepath.Split(name)
	}

	// The function that calls the builder func
	_, f, l, _ := runtime.Caller(3)
	_, file := filepath.Split(f)

	return fmt.Sprintf("%s(%s:%d)", name, file, l)
}
