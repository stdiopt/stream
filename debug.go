package stream

import (
	"fmt"
	"path/filepath"
	"runtime"
)

type callerInfo struct {
	where string
	name  string
}

func (ci callerInfo) String() string {
	return fmt.Sprintf("%s(%s)", ci.name, ci.where)
}

func procName() callerInfo {
	var name string

	{ // Function name
		pc, _, _, _ := runtime.Caller(2)
		fi := runtime.FuncForPC(pc)
		name = fi.Name()
		_, name = filepath.Split(name)
	}

	// Where it was called
	_, f, l, _ := runtime.Caller(3)
	_, file := filepath.Split(f)

	return callerInfo{
		where: fmt.Sprintf("%s:%d", file, l),
		name:  name,
	}
}
