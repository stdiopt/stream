package strmrefl

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stdiopt/stream/strmtest"
)

func TestExtract(t *testing.T) {
	type sub struct {
		Sub string
	}
	type sample struct {
		Slice []sub
	}
	type args struct {
		f []interface{}
	}
	tests := []struct {
		name        string
		args        args
		send        []interface{}
		senderError error

		want      []interface{}
		wantErr   string
		wantPanic string
	}{
		{
			name: "extracts field",
			args: args{[]interface{}{"Slice", 0, "Sub"}},
			send: []interface{}{
				sample{Slice: []sub{{Sub: "sub test string"}}},
				sample{Slice: []sub{{Sub: "sub test string 2"}}},
			},
			want: []interface{}{
				"sub test string",
				"sub test string 2",
			},
		},
		{
			name: "returns error on field invalid field",
			args: args{[]interface{}{"Test"}},
			send: []interface{}{
				sample{Slice: []sub{{Sub: "sub test string"}}},
				sample{Slice: []sub{{Sub: "sub test string 2"}}},
			},
			wantErr: "struct: field invalid",
		},
		{
			name: "returns error when sender errors",
			args: args{[]interface{}{"Slice"}},
			send: []interface{}{
				sample{Slice: []sub{{Sub: "sub test string"}}},
				sample{Slice: []sub{{Sub: "sub test string 2"}}},
			},
			senderError: errors.New("sender error"),
			wantErr:     "sender error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				p := recover()
				if !strmtest.MatchPanic(tt.wantPanic, p) {
					t.Errorf("Extract() panic = %v, wantPanic %v", p, tt.wantPanic)
				}
			}()

			pp := Extract(tt.args.f...)
			if pp == nil {
				t.Errorf("Extract() is nil = %v, want %v", pp == nil, false)
			}

			st := strmtest.New(t, pp)
			for _, s := range tt.send {
				st.Send(s).WithSenderError(tt.senderError)
			}
			st.ExpectFull(tt.want...).
				ExpectError(tt.wantErr).
				Run()
		})
	}
}

func TestStructMap(t *testing.T) {
	type sample struct {
		String string
		Int    int
		Slice  []int
	}
	type args struct {
		target interface{}
		fm     FMap
	}
	tests := []struct {
		name        string
		args        args
		send        []interface{}
		senderError error

		want      []interface{}
		wantErr   string
		wantPanic string
	}{
		{
			name: "maps a struct",
			args: args{
				target: sample{},
				fm: FMap{
					"String": F(),
				},
			},
			send: []interface{}{"a"},
			want: []interface{}{
				sample{
					String: "a",
				},
			},
		},
		{
			name: "maps a struct",
			args: args{
				target: sample{},
				fm: FMap{
					"String": F(0),
					"Int":    F(1),
					"Slice":  F(2),
				},
			},
			send: []interface{}{
				[]interface{}{
					"some string",
					7,
					[]int{1, 2, 3},
				},
			},
			want: []interface{}{
				sample{
					String: "some string",
					Int:    7,
					Slice:  []int{1, 2, 3},
				},
			},
		},
		{
			name: "return error on invalid target field",
			args: args{
				target: sample{},
				fm: FMap{
					"Test": F(),
				},
			},
			send:    []interface{}{"a"},
			wantErr: `field not found "Test" in strmrefl.sample`,
		},
		{
			name: "return error on invalid map field",
			args: args{
				target: sample{},
				fm: FMap{
					"String": F("b"),
				},
			},
			send:    []interface{}{"a"},
			wantErr: `invalid type`,
		},
		{
			name: "panics on nil target",
			args: args{
				target: nil,
			},
			send:      []interface{}{"a"},
			wantPanic: "target value is nil",
		},
		{
			name: "returns error on sender error",
			args: args{
				target: sample{},
			},
			send:        []interface{}{"a"},
			senderError: errors.New("sender error"),
			wantErr:     "sender error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				p := recover()
				if !strmtest.MatchPanic(tt.wantPanic, p) {
					t.Errorf("StructMap() panic = %v, wantPanic %v", p, tt.wantPanic)
				}
			}()

			pp := StructMap(tt.args.target, tt.args.fm)
			if pp == nil {
				t.Errorf("StructMap() is nil = %v, want %v", pp == nil, false)
			}

			st := strmtest.New(t, StructMap(tt.args.target, tt.args.fm))
			for _, s := range tt.send {
				st.Send(s).WithSenderError(tt.senderError)
			}
			st.ExpectFull(tt.want...).
				ExpectError(tt.wantErr).
				Run()
		})
	}
}

func TestFieldOf(t *testing.T) {
	type key struct{ k string }
	type sample struct {
		Comp []map[interface{}][]interface{}
	}
	type args struct {
		v  interface{}
		ff []interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr string
	}{
		{
			name: "returns the passed value",
			args: args{v: "field"},
			want: "field",
		},
		{
			name: "returns slice element",
			args: args{
				v:  []int{1, 2, 3},
				ff: []interface{}{1},
			},
			want: 2,
		},
		{
			name: "returns struct field",
			args: args{
				v:  struct{ Field string }{"field"},
				ff: []interface{}{"Field"},
			},
			want: "field",
		},
		{
			name: "returns map element",
			args: args{
				v:  map[string]interface{}{"key 1": "map field"},
				ff: []interface{}{"key 1"},
			},
			want: "map field",
		},
		{
			name: "returns deep field",
			args: args{
				v: sample{
					Comp: []map[interface{}][]interface{}{
						{
							key{"1"}: []interface{}{
								"deep",
							},
						},
					},
				},
				ff: []interface{}{"Comp", 0, key{"1"}, 0},
			},
			want: "deep",
		},
		{
			name: "returns error if invalid",
			args: args{
				v:  nil,
				ff: []interface{}{"Test"},
			},
			wantErr: `invalid type <nil>`,
		},
		{
			name: "returns error if invalid struct field",
			args: args{
				v:  sample{},
				ff: []interface{}{"Test"},
			},
			wantErr: `struct: field invalid: "Test" of strmrefl.sample`,
		},
		{
			name: "returns error if invalid slice index",
			args: args{
				v:  []int{1, 2, 3},
				ff: []interface{}{"Test"},
			},
			want:    nil,
			wantErr: `slice: field invalid: "Test" of \[\]int`,
		},
		{
			name: "returns error if sub field invalid",
			args: args{
				v:  map[string]interface{}{"stuff": 1},
				ff: []interface{}{"stuff1", "err"},
			},
			wantErr: `invalid type `,
		},
		{
			name: "returns error if field invalid",
			args: args{
				v:  1,
				ff: []interface{}{"stuff1"},
			},
			wantErr: `invalid type `,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FieldOf(tt.args.v, tt.args.ff...)
			if !strmtest.MatchError(tt.wantErr, err) {
				t.Errorf("FieldOf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FieldOf() = %v, want %v", got, tt.want)
			}
		})
	}
}
