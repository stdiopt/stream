package strmutil

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"reflect"
)

// JSONParse parses the []byte input as json and send the object
// interface can the type of object that will parse and send
//		type MyType struct {
//			Name string
//		}
//		JSONParse(&MyType{})
// if type is nil it will decode the input into an &interface{} which might
// produce different types map[string]interface{}, []interface{}, string, float64
// as the regular native json.Unmarshal
// if the input is not bytes it will error and cancel the pipeline
func JSONParse(v interface{}) ProcFunc {
	if v == nil {
		var l interface{}
		v = &l
	}
	typ := reflect.Indirect(reflect.ValueOf(v)).Type()
	return func(p Proc) error {
		rd := AsReader(p)
		dec := json.NewDecoder(rd)
		for {
			v := reflect.New(typ).Interface()
			err := dec.Decode(v)
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return rd.CloseWithError(err)
			}
			// deref
			if vv, ok := v.(*interface{}); ok {
				v = *vv
			}
			if err := p.Send(v); err != nil {
				return rd.CloseWithError(err)
			}
		}
	}
}

// JSONDump encodes the input as json into the writer
// TODO: {lpf} rename, this was meant for debug but might be good for general use
func JSONDump(w io.Writer) ProcFunc {
	return func(p Proc) error {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return p.Consume(func(v interface{}) error {
			if err := enc.Encode(v); err != nil {
				log.Println("err:", err)
				return err
			}
			fmt.Fprintln(w)
			return p.Send(v)
		})
	}
}
