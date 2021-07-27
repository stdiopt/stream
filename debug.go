package stream

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
)

type DebugOpt struct {
	output    io.Writer
	Processor bool
	Value     bool
}

func Debug(w io.Writer) Processor {
	return Func(func(p Proc) error {
		return p.Consume(func(v interface{}) error {
			p.MetaSet("_debug", w)
			return p.Send(v)
		})
	})
}

func DebugProc(w io.Writer, p Proc, v interface{}) {
	name := "<unknown>"
	if p, ok := p.(*proc); ok {
		name = p.name
	}
	fmt.Fprintf(w, "[\033[01;37m%s\033[0m] ", name)
	fmt.Fprintf(w, "\033[01;33m%T\033[0m ", v)
	fmt.Fprintf(w, "meta: \033[34m%v\033[0m\n\t", p.Meta())

	switch v := v.(type) {
	case string:
		fmt.Fprint(w, v)
	case []byte: // force []byte
		fmt.Fprint(w, v)
	default:
		data, err := json.Marshal(v)
		if err != nil {
			fmt.Fprintf(w, "<json err>")
			break
		}
		if _, err := w.Write(data); err != nil {
			fmt.Fprintf(w, "<err>")
			break
		}
	}
	fmt.Fprint(w, "\n\n")
}

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
