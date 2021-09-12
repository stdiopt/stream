package strmdrow

import "fmt"

type Header struct {
	fields []Field
	index  map[string]int
}

func NewHeader(fields ...Field) *Header {
	hdr := &Header{
		fields: fields,
	}
	hdr.index = map[string]int{}
	for i, f := range fields {
		hdr.index[f.Name] = i
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
		panic(fmt.Sprintf(`header "%s" already exists`, f.Name))
	}
	h.fields = append(h.fields, f)

	i := len(h.fields) - 1
	h.index[f.Name] = i
	return i
}

func (h *Header) Len() int {
	return len(h.fields)
}
