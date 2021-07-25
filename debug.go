package stream

import (
	"encoding/json"
	"fmt"
	"io"
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

func DebugProc(w io.Writer, p *proc, v interface{}) {
	fmt.Fprintf(w, "[\033[01;37m%s\033[0m] ", p.dname)
	fmt.Fprintf(w, "\033[01;33m%T\033[0m ", v)
	fmt.Fprintf(w, "meta: \033[34m%v\033[0m\n\t", p.meta.values)

	switch v := v.(type) {
	case string:
		fmt.Fprint(w, v)
	case []byte:
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
