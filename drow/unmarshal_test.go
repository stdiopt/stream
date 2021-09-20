package drow

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stdiopt/stream/utils/strmtest"
)

type sample struct {
	ID        int
	Name      string
	Other     float64
	Data      []byte
	Unconvert string
	CreatedAt time.Time
}

type sampleJSON struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	No   string
}

// nolint: structcheck,unused
type sampleUnexported struct {
	id   int
	Name string `sql:"name"`
}

type sampleUnmarshaler struct {
	ID int
}

func (s *sampleUnmarshaler) UnmarshalDROW(row Row) error {
	s.ID = row.GetInt("id")
	return nil
}

func TestUnmarshal(t *testing.T) {
	now := time.Now()

	type args struct {
		r    Row
		v    interface{}
		opts []UnmarshalOpt
	}
	tests := []struct {
		name      string
		args      args
		want      interface{}
		wantErr   string
		wantPanic string
	}{

		{
			name: "unmarshal a struct",
			args: args{
				r: func() Row {
					row := New()
					row.SetOrAdd("ID", 7)
					row.SetOrAdd("Name", "name")
					row.SetOrAdd("Data", "some string")
					row.SetOrAdd("CreatedAt", now)
					return row
				}(),
				v: &sample{},
			},
			want: &sample{
				ID:        7,
				Name:      "name",
				Data:      []byte(`some string`),
				CreatedAt: now,
			},
		},
		{
			name: "unmarshal a struct with tagname",
			args: args{
				r: func() Row {
					row := New()
					row.SetOrAdd("id", 7)
					row.SetOrAdd("name", "name")
					row.SetOrAdd("no", "ok")
					return row
				}(),
				v: &sampleJSON{},
				opts: []UnmarshalOpt{
					UnmarshalOption.WithTag("json"),
				},
			},
			want: &sampleJSON{
				ID:   7,
				Name: "name",
			},
		},
		{
			name: "unmarshal a struct with unexported",
			args: args{
				r: func() Row {
					row := New()
					row.SetOrAdd("id", 7)
					row.SetOrAdd("name", "sqlname")
					return row
				}(),
				v: &sampleUnexported{},
				opts: []UnmarshalOpt{
					UnmarshalOption.WithTag("sql"),
				},
			},
			want: &sampleUnexported{
				Name: "sqlname",
			},
		},
		{
			name: "unmarshal an unmarshaller",
			args: args{
				r: func() Row {
					row := New()
					row.SetOrAdd("id", 7)
					return row
				}(),
				v: &sampleUnmarshaler{},
			},
			want: &sampleUnmarshaler{
				ID: 7,
			},
		},
		{
			name: "errors if type is not convertible",
			args: args{
				r: func() Row {
					row := New()
					row.SetOrAdd("Unconvert", 7)
					return row
				}(),
				v: &sample{},
			},
			want:    &sample{},
			wantErr: "'int' cannot be converted to 'string'",
		},
		{
			name: "panics if param is not pointer",
			args: args{
				r: func() Row {
					row := New()
					row.SetOrAdd("id", 7)
					return row
				}(),
				v: sample{},
			},
			wantPanic: "param should be a pointer to struct",
		},
		{
			name: "panics if param is not pointer struct",
			args: args{
				r: func() Row {
					row := New()
					row.SetOrAdd("id", 7)
					return row
				}(),
				v: new(int),
			},
			wantPanic: "param should be a pointer to struct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				p := recover()
				if !strmtest.MatchPanic(tt.wantPanic, p) {
					t.Errorf("Unmarshal() panic = %v, wantPanic %v", p, tt.wantPanic)
				}
			}()
			if err := Unmarshal(tt.args.r, tt.args.v, tt.args.opts...); !strmtest.MatchError(tt.wantErr, err) {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, tt.args.v, cmp.AllowUnexported(sampleUnexported{})); diff != "" {
				t.Errorf("Unmarshal() + want - got\n: %s", diff)
			}
		})
	}
}
