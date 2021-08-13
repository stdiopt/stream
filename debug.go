package stream

import (
	"bytes"
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

func Debug(w io.Writer) ProcFunc {
	return Func(func(p Proc) error {
		return p.Consume(func(v interface{}) error {
			return p.Send(v)
		})
	})
}

func DebugProc(w io.Writer, pp Proc, v interface{}) {
	name := "<unknown>"
	if p, ok := pp.(*proc); ok {
		name = p.name
	}
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "[\033[01;37m%s\033[0m] ", name)
	fmt.Fprintf(buf, "\033[01;33m%T\033[0m ", v)

	switch v := v.(type) {
	case string:
		fmt.Fprint(buf, v)
	case []byte: // force []byte
		fmt.Fprint(buf, v)
	default:
		data, err := json.Marshal(v)
		if err != nil {
			fmt.Fprintf(buf, "<json err>")
			break
		}
		if _, err := buf.Write(data); err != nil {
			fmt.Fprintf(buf, "<err>")
			break
		}
	}
	fmt.Fprint(buf, "\n\n")
	fmt.Fprint(w, buf.String())
	// io.Copy(w, buf)
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
