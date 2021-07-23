package stream

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

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

func DebugCtx(ctx context.Context, prefix string, v interface{}) string {
	buf := &bytes.Buffer{}

	meta, _ := MetaFromContext(ctx)

	fmt.Fprintf(buf, "[\033[01;37m%s\033[0m] ", prefix)
	fmt.Fprintf(buf, "\033[01;33m%T\033[0m ", v)
	fmt.Fprintf(buf, "meta: \033[34m%v\033[0m\n\t", meta)

	switch v := v.(type) {
	case string:
		fmt.Fprint(buf, v)
	case []byte:
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

	return buf.String()
}
