package stream

import (
	"fmt"
	"log"
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
	log.Println("File is:", f, runtime.GOROOT())
	_, file := filepath.Split(f)

	return callerInfo{
		where: fmt.Sprintf("%s:%d", file, l),
		name:  name,
	}
}
