// Package drow provides a type containing meta data information.
// like a row in a table with column meta data
package drow

import "reflect"

type Column struct {
	Name  string
	Type  reflect.Type // ???
	Value interface{}
}

type Row struct {
	Name    string
	Columns []Column
}
