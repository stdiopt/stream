package strmsql

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stdiopt/stream/strmtest"
)

type sample struct {
	One string
	Two string
}

func TestDialect_InsertBatch(t *testing.T) {
	type args struct {
		batchSize int
		qry       string
		params    []interface{}
	}
	tests := []struct {
		name        string
		d           Dialect
		args        args
		send        []interface{}
		senderError error

		want      []interface{}
		wantQuery func(mock sqlmock.Sqlmock)
		wantErr   string
	}{

		{
			name: "executes query",
			d:    PSQL,
			args: args{
				batchSize: 2,
				qry:       "insert into table values",
				params:    []interface{}{1, 2},
			},
			send: []interface{}{1},
			want: []interface{}{1},
			wantQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(`insert into table values($1,$2)`)).
					WithArgs(1, 2).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name: "executes PSQL batch insert",
			d:    PSQL,
			args: args{
				batchSize: 2,
				qry:       "insert into table values",
				params: []interface{}{
					Field("One"),
					Field("Two"),
				},
			},
			send: []interface{}{
				sample{One: "1.1", Two: "2.1"},
				sample{One: "2.1", Two: "2.2"},
				sample{One: "3.1", Two: "2.3"},
				sample{One: "4.1", Two: "2.4"},
			},
			want: []interface{}{
				sample{One: "1.1", Two: "2.1"},
				sample{One: "2.1", Two: "2.2"},
				sample{One: "3.1", Two: "2.3"},
				sample{One: "4.1", Two: "2.4"},
			},
			wantQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(
					"insert into table values($1,$2),\n($3,$4)")).
					WithArgs(
						"1.1", "2.1",
						"2.1", "2.2",
					).
					WillReturnResult(sqlmock.NewResult(1, 2))
				mock.ExpectExec(regexp.QuoteMeta(
					"insert into table values($1,$2),\n($3,$4)\n")).
					WithArgs(
						"3.1", "2.3",
						"4.1", "2.4",
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name: "executes MYSQL batch insert",
			d:    MySQL,
			args: args{
				batchSize: 2,
				qry:       "insert into table values",
				params: []interface{}{
					Field("One"),
					Field("Two"),
				},
			},
			send: []interface{}{
				sample{One: "1.1", Two: "2.1"},
				sample{One: "2.1", Two: "2.2"},
				sample{One: "3.1", Two: "2.3"},
				sample{One: "4.1", Two: "2.4"},
			},
			want: []interface{}{
				sample{One: "1.1", Two: "2.1"},
				sample{One: "2.1", Two: "2.2"},
				sample{One: "3.1", Two: "2.3"},
				sample{One: "4.1", Two: "2.4"},
			},
			wantQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(
					"insert into table values(?,?),\n(?,?)")).
					WithArgs(
						"1.1", "2.1",
						"2.1", "2.2",
					).
					WillReturnResult(sqlmock.NewResult(1, 2))
				mock.ExpectExec(regexp.QuoteMeta(
					"insert into table values(?,?),\n(?,?)\n")).
					WithArgs(
						"3.1", "2.3",
						"4.1", "2.4",
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name: "returns error when param errors",
			d:    PSQL,
			args: args{
				batchSize: 1,
				qry:       "insert into table values",
				params: []interface{}{
					Field("not found"),
				},
			},
			send:    []interface{}{1},
			wantErr: `strmsql.Dialect.InsertBatch.* invalid type 'int' for field: "not found"$`,
		},
		{
			name: "returns error when query errors",
			d:    PSQL,
			args: args{
				batchSize: 1,
				qry:       "insert into table values",
				params: []interface{}{
					1,
				},
			},
			send: []interface{}{1},
			wantQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(`insert into table values($1)`)).
					WithArgs(1).
					WillReturnError(errors.New("query error"))
			},
			wantErr: `strmsql.Dialect.InsertBatch.* query error`,
		},
		{
			name: "returns error when sender errors",
			d:    PSQL,
			args: args{
				batchSize: 1,
				qry:       "insert into table values",
				params: []interface{}{
					1, 2,
				},
			},
			send:        []interface{}{1, 2, 3},
			senderError: errors.New("sender error"),

			wantQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(`insert into table values($1,$2)`)).
					WithArgs(1, 2).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: "strmsql.Dialect.InsertBatch.* sender error$",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}

			pp := tt.d.InsertBatch(
				db,
				tt.args.batchSize,
				tt.args.qry,
				tt.args.params...,
			)
			if pp == nil {
				t.Errorf("InsertBatch() is nil = %v, want %v", pp == nil, false)
			}

			if tt.wantQuery != nil {
				tt.wantQuery(mock)
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

func TestDialect_Exec(t *testing.T) {
	type args struct {
		qry    string
		params []interface{}
	}
	tests := []struct {
		name        string
		d           Dialect
		args        args
		send        []interface{}
		senderError error

		want      []interface{}
		wantQuery func(mock sqlmock.Sqlmock)
		wantErr   string
	}{

		{
			name: "executes query",
			args: args{
				qry:    "insert into table values($1,$2)",
				params: []interface{}{1, 2},
			},
			send: []interface{}{1},
			want: []interface{}{1},
			wantQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(`insert into table values($1,$2)`)).
					WithArgs(1, 2).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name: "returns error when param errors",
			args: args{
				qry:    "insert into table values($1, $2)",
				params: []interface{}{Field("not found")},
			},
			send:    []interface{}{1},
			wantErr: `strmsql.Dialect.Exec.* invalid type 'int' for field: "not found"$`,
		},
		{
			name: "returns error when query errors",
			args: args{
				qry:    "insert into table values($1)",
				params: []interface{}{1},
			},
			send: []interface{}{1},
			wantQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(`insert into table values($1)`)).
					WithArgs(1).
					WillReturnError(errors.New("query error"))
			},
			wantErr: `strmsql.Dialect.Exec.* query error`,
		},
		{
			name: "returns error when sender errors",

			args: args{
				qry:    "insert into table values($1,$2)",
				params: []interface{}{1, 2},
			},
			send:        []interface{}{1, 2, 3},
			senderError: errors.New("sender error"),

			wantQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(`insert into table values($1,$2)`)).
					WithArgs(1, 2).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: "strmsql.Dialect.Exec.* sender error$",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}

			pp := tt.d.Exec(
				db,
				tt.args.qry,
				tt.args.params...,
			)
			if pp == nil {
				t.Errorf("Exec() is nil = %v, want %v", pp == nil, false)
			}

			if tt.wantQuery != nil {
				tt.wantQuery(mock)
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
