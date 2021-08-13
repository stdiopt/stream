package dframe

import "reflect"

type DFrame struct {
	columns    []string
	columnType []reflect.Type // go based time?
}
