// Package drow provides a type containing meta data information.
// like a row in a table with column meta data
package strmdrow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	strm "github.com/stdiopt/stream"
)

func init() {
	type consumerType = func(Row) error
	type senderType = func(strm.Sender, Row) error
	// Registers this type on consumer registry for a fast path
	strm.ConsumerRegistry[reflect.TypeOf(consumerType(nil))] = func(f interface{}) strm.ConsumerFunc {
		fn := f.(consumerType)
		return func(v interface{}) error {
			r, ok := v.(Row)
			if !ok {
				return strm.NewTypeMismatchError(Row{}, v)
			}
			return fn(r)
		}
	}
	strm.SRegistry[reflect.TypeOf(senderType(nil))] = func(f interface{}) strm.ProcFunc {
		fn := f.(senderType)
		return func(p strm.Proc) error {
			return p.Consume(func(r Row) error {
				return fn(p, r)
			})
		}
	}
}

type Field struct {
	Name string
	Type reflect.Type
	Tag  reflect.StructTag
}

type Header struct {
	fields []Field
	index  map[string]int
}

func NewHeader(fields ...Field) Header {
	hdr := Header{}
	for _, f := range fields {
		hdr.Add(f)
	}
	return hdr
}

func (h *Header) Fields() []Field {
	return h.fields
}

func (h *Header) Add(f Field) int {
	if h.index == nil {
		h.index = map[string]int{}
	}
	_, ok := h.index[f.Name]
	if ok {
		panic("header already exists")
	}
	h.fields = append(h.fields, f)

	i := len(h.fields) - 1
	h.index[f.Name] = i
	return i
}

func (h *Header) Len() int {
	return len(h.fields)
}

type Row struct {
	header Header
	Values []interface{}
}

func New() Row {
	return Row{}
}

func NewWithHeader(hdr Header) Row {
	return (Row{}).WithHeader(hdr)
}

func (r Row) WithHeader(hdr Header) Row {
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

func (r Row) HeaderByName(i int) Field {
	if i < 0 || i >= len(r.header.fields) {
		return Field{}
	}
	return r.header.fields[i]
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

func (r *Row) Set(k string, v interface{}) {
	i, ok := r.header.index[k]
	if !ok {
		i = r.header.Add(Field{
			Name: k,
			Type: reflect.TypeOf(v),
		})
	}
	// Reset header
	// r.header.fields[i] = Field{Name: k, Type: reflect.TypeOf(v)}
	if t := reflect.TypeOf(v); t != r.header.fields[i].Type {
		panic(fmt.Sprintf("can't assign %v to %v", t, r.header.fields[i].Type))
	}
	// Bound check
	if i <= len(r.Values) {
		n := make([]interface{}, i+1)
		copy(n, r.Values)
		r.Values = n
	}
	r.Values[i] = v
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
	dec.UseNumber()

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

		r.Set(key, value)
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
			return nil, fmt.Errorf("Unexpected delimiter: %q", delim)
		}
	}
	return t, nil
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
