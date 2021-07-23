package stream

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/bwmarrin/snowflake"
)

var flake *snowflake.Node = func() *snowflake.Node {
	r, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}
	return r
}()

type Meta struct {
	callers []string
	origin  snowflake.ID
	base    snowflake.ID
	id      snowflake.ID
	src     map[snowflake.ID]struct{}
	props   map[string]interface{}
}

func NewMeta() Meta {
	newID := flake.Generate()
	return Meta{
		origin: newID,
		id:     newID,
	}
}

func (m Meta) String() string {
	var propsBuf bytes.Buffer
	count := 0
	for k, v := range m.props {
		if k[0] == '_' {
			continue
		}
		if count != 0 {
			propsBuf.WriteString(", ")
		}
		count++
		fmt.Fprintf(&propsBuf, `%v:"%v"`, k, v)
	}

	var props string
	if propsBuf.Len() != 0 {
		props = "props: {" + propsBuf.String() + "}"
	}

	var srcBuf bytes.Buffer
	count = 0
	for k := range m.src {
		if count != 0 {
			srcBuf.WriteString(", ")
		} else {
			srcBuf.WriteString("src: ")
		}
		count++
		fmt.Fprintf(&srcBuf, "%s", sfToString(k))
	}

	return fmt.Sprintf(
		"(origin: %s) %s->%s callers: %v, %v %v",
		sfToString(m.origin),
		sfToString(m.base),
		sfToString(m.id),
		m.callers,
		srcBuf.String(),
		props,
	)
}

func (m Meta) OriginID() snowflake.ID {
	if m.origin == 0 {
		return m.id
	}
	return m.origin
}

func (m Meta) ID() snowflake.ID {
	return m.id
}

func (m Meta) Value(k string) interface{} {
	if m.props == nil {
		return nil
	}
	return m.props[k]
}

func (m Meta) MarshalJSON() ([]byte, error) {
	mm := struct {
		Callers []string                  `json:"caller,omitempty"`
		Src     map[snowflake.ID]struct{} `json:"src,omitempty"`
		Base    snowflake.ID              `json:"base,omitempty"`
		ID      snowflake.ID              `json:"id,omitempty"`
		Props   map[string]interface{}    `json:"props,omitempty"`
	}{
		Callers: m.callers,
		Src:     m.src,
		Base:    m.base,
		ID:      m.id,
		Props:   m.props,
	}
	return json.Marshal(mm)
}

func (m Meta) WithCaller(c string) Meta {
	r := m.clone()
	r.callers = append(r.callers, c)
	return r
}

func (m Meta) WithValue(k string, v interface{}) Meta {
	r := m.clone()
	r.props[k] = v
	return r
}

func (m Meta) Merge(o Meta) Meta {
	if _, ok := m.src[o.id]; ok {
		return m
	}

	r := m.clone()
	for k, v := range o.props {
		r.props[k] = v
	}
	for k := range o.src {
		r.src[k] = struct{}{}
	}
	log.Println("Setting src:", r.src)
	r.src[o.id] = struct{}{}

	r.callers = append(r.callers, o.callers...)

	return r
}

func (m Meta) clone() Meta {
	newID := flake.Generate()

	var callers []string
	callers = append(callers, m.callers...)

	props := map[string]interface{}{}
	for k, v := range m.props {
		props[k] = v
	}

	src := map[snowflake.ID]struct{}{}
	for k := range m.src {
		src[k] = struct{}{}
	}

	origin := m.OriginID()
	if origin == 0 {
		origin = newID
	}
	return Meta{
		callers: callers,
		origin:  origin,
		base:    m.id,
		id:      newID,
		src:     src,
		props:   props,
	}
}

var metaCtxKey = struct{ s string }{"metaKey"}

// Should this be merged?! just set?!
func ContextWithMeta(ctx context.Context, m Meta) context.Context {
	return context.WithValue(ctx, metaCtxKey, m)
}

// MetaFromContext returns a clone from context
func MetaFromContext(ctx context.Context) (Meta, bool) {
	m, ok := ctx.Value(metaCtxKey).(Meta)
	return m, ok
}

func ContextWithMetaValue(ctx context.Context, k string, v interface{}) context.Context {
	m, _ := MetaFromContext(ctx)
	return ContextWithMeta(ctx, m.WithValue(k, v))
}

func MetaValueFromContext(ctx context.Context, k string) interface{} {
	m, _ := MetaFromContext(ctx)
	return m.Value(k)
}

func sfToString(sf snowflake.ID) string {
	validCharacters := []byte("qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890")

	buf := bytes.Buffer{}
	quot := int64(sf)
	var rem int64
	for quot != 0 {
		rem = quot % int64(len(validCharacters))
		quot = quot / int64(len(validCharacters))
		buf.WriteByte(validCharacters[rem])
	}
	return buf.String()
}
