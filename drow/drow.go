// Package drow provides a type containing meta data information.
// like a row in a table with column meta data
package drow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

type Row struct {
	header *Header
	Values []interface{}
}

func New() Row {
	return Row{header: NewHeader()}
}

func NewWithHeader(hdr *Header) Row {
	return Row{
		header: hdr,
		Values: make([]interface{}, len(hdr.fields)),
	}
}

func (r Row) WithHeader(hdr *Header) Row {
	r.header = hdr
	return r
}

func (r Row) NumField() int {
	return len(r.header.fields)
}

func (r Row) Header(i int) Field {
	if i < 0 || i >= len(r.header.fields) {
		return Field{}
	}
	return r.header.fields[i]
}

func (r Row) HeaderByName(k string) Field {
	if r.header == nil || r.header.index == nil {
		return Field{}
	}
	i, ok := r.header.index[k]
	if !ok {
		return Field{}
	}

	if i < 0 || i >= len(r.header.fields) {
		return Field{}
	}
	return r.header.fields[i]
}

type FieldValue struct {
	Field
	Value interface{}
}

func (r Row) Meta(k string) *FieldValue {
	if r.header == nil || r.header.index == nil {
		return nil
	}
	i, ok := r.header.index[k]
	if !ok {
		return nil
	}
	return &FieldValue{
		Field: r.Header(i),
		Value: r.Value(i),
	}
}

func (r Row) Get(k string) interface{} {
	return r.Value(r.header.index[k])
}

func (r Row) GetInt(k string) int {
	i, _ := r.Get(k).(int)
	return i
}

func (r Row) GetString(k string) string {
	i, _ := r.Get(k).(string)
	return i
}

func (r Row) Value(i int) interface{} {
	if i < 0 || i >= len(r.Values) {
		return nil
	}
	return r.Values[i]
}

// Do not allow adding a new header unless flagged to
func (r *Row) SetI(i int, v interface{}) {
	if i < 0 || i > len(r.header.fields) {
		panic(fmt.Sprintf("field %d doesn't exists", i))
	}
	if t := reflect.TypeOf(v); t != r.header.fields[i].Type {
		panic(fmt.Sprintf("can't assign %v to %v", t, r.header.fields[i].Type))
	}
	r.Values[i] = v
}

// Do not allow adding a new header unless flagged to
func (r *Row) Set(k string, v interface{}) {
	i, ok := r.header.index[k]
	if !ok {
		panic(fmt.Sprintf("field %s doesn't exists", k))
	}
	if t := reflect.TypeOf(v); t != r.header.fields[i].Type {
		panic(fmt.Sprintf("can't assign %v to %v", t, r.header.fields[i].Type))
	}
	r.Values[i] = v
}

func (r *Row) SetOrAdd(k string, v interface{}) {
	if r.header == nil {
		r.header = NewHeader()
	}
	i, ok := r.header.index[k]
	if !ok {
		i = r.header.Add(Field{
			Name: k,
			Type: reflect.TypeOf(v),
		})
	}
	if t := reflect.TypeOf(v); t != r.header.fields[i].Type {
		r.header.fields[i] = Field{Name: k, Type: reflect.TypeOf(v)}
	}
	// Bound check
	if i <= len(r.Values) {
		n := make([]interface{}, i+1)
		copy(n, r.Values)
		r.Values = n
	}
	r.Values[i] = v
}

func (r Row) String() string {
	buf := &bytes.Buffer{}
	buf.WriteString("{")
	for i := 0; i < r.NumField(); i++ {
		h := r.Header(i)
		fmt.Fprintf(buf, "(%q:%v:%v) ", h.Name, h.Type, r.Value(i))
	}
	buf.WriteString("}")
	return buf.String()
}

func (r Row) MarshalJSON() ([]byte, error) {
	buf := bytes.Buffer{}

	buf.WriteString("{")
	for i, v := range r.Values {
		d, err := json.Marshal(r.Header(i).Name)
		if err != nil {
			return nil, err
		}
		buf.Write(d)
		buf.WriteString(":")
		d, err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
		buf.Write(d)
		if i < len(r.Values)-1 {
			buf.WriteString(",")
		}
	}
	buf.WriteString("}")
	return buf.Bytes(), nil
}

func (r *Row) UnmarshalJSON(d []byte) error {
	dec := json.NewDecoder(bytes.NewReader(d))
	// dec.UseNumber()

	t, err := dec.Token()
	if err != nil {
		return err
	}

	if delim, ok := t.(json.Delim); !ok || delim != '{' {
		return fmt.Errorf("expect JSON object open with '{'")
	}
	if err := r.parseobject(dec); err != nil {
		return err
	}

	return nil
}

func (r *Row) parseobject(dec *json.Decoder) error {
	var t json.Token
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			return err
		}

		key, ok := t.(string)
		if !ok {
			return fmt.Errorf("expecting JSON key should be always a string: %T: %v", t, t)
		}

		t, err = dec.Token()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		var value interface{}
		value, err = handledelim(t, dec)
		if err != nil {
			return err
		}

		r.SetOrAdd(key, value)
	}

	t, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '}' {
		return fmt.Errorf("expect JSON object close with '}'")
	}

	return nil
}

func parsearray(dec *json.Decoder) ([]interface{}, error) {
	var t json.Token
	var arr []interface{}
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			return nil, err
		}

		var value interface{}
		value, err = handledelim(t, dec)
		if err != nil {
			return nil, err
		}
		arr = append(arr, value)
	}
	t, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if delim, ok := t.(json.Delim); !ok || delim != ']' {
		err = fmt.Errorf("expect JSON array close with ']'")
		return nil, err
	}

	return arr, nil
}

func handledelim(t json.Token, dec *json.Decoder) (res interface{}, err error) {
	if delim, ok := t.(json.Delim); ok {
		switch delim {
		case '{':
			r := New()
			err = r.parseobject(dec)
			if err != nil {
				return
			}
			return r, nil
		case '[':
			var value []interface{}
			value, err = parsearray(dec)
			if err != nil {
				return
			}
			return value, nil
		default:
			return nil, fmt.Errorf("unexpected delimiter: %q", delim)
		}
	}
	return t, nil
}
