package strmsql

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmtest"
)

type sample struct {
	One string
	Two string
}

func TestInsertBatch(t *testing.T) {
	tests := []struct {
		name        string
		pipe        func(db *sql.DB) strm.Pipe
		send        []interface{}
		senderError error

		want        []interface{}
		wantQuery   func(mock sqlmock.Sqlmock)
		wantErrorRE string
	}{
		{
			name: "executes query",
			pipe: func(db *sql.DB) strm.Pipe {
				return PSQL.InsertBatch(
					db,
					2,
					"insert into table values",
					1, 2,
				)
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
			pipe: func(db *sql.DB) strm.Pipe {
				return PSQL.InsertBatch(
					db,
					2,
					"insert into table values",
					Field("One"),
					Field("Two"),
				)
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
			pipe: func(db *sql.DB) strm.Pipe {
				return MySQL.InsertBatch(
					db,
					2,
					"insert into table values",
					Field("One"),
					Field("Two"),
				)
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
			pipe: func(db *sql.DB) strm.Pipe {
				return PSQL.InsertBatch(
					db,
					1,
					"insert into table values",
					Field("not found"),
				)
			},
			send:        []interface{}{1},
			wantErrorRE: `strmsql.Dialect.InsertBatch.* invalid type 'int' for field: "not found"$`,
		},
		{
			name: "returns error when query errors",
			pipe: func(db *sql.DB) strm.Pipe {
				return PSQL.InsertBatch(
					db,
					1,
					"insert into table values",
					1,
				)
			},
			send: []interface{}{1},
			wantQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(`insert into table values($1)`)).
					WithArgs(1).
					WillReturnError(errors.New("query error"))
			},
			wantErrorRE: `strmsql.Dialect.InsertBatch.* query error`,
		},
		{
			name: "returns error when sender errors",
			pipe: func(db *sql.DB) strm.Pipe {
				return PSQL.InsertBatch(
					db,
					1,
					"insert into table values",
					1, 2,
				)
			},
			send:        []interface{}{1, 2, 3},
			senderError: errors.New("sender error"),

			wantQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(`insert into table values($1,$2)`)).
					WithArgs(1, 2).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErrorRE: "strmsql.Dialect.InsertBatch.* sender error$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}

			if tt.wantQuery != nil {
				tt.wantQuery(mock)
			}
			st := strmtest.New(t, tt.pipe(db))
			for _, s := range tt.send {
				st.Send(s).WithSenderError(tt.senderError)
			}
			st.ExpectFull(tt.want...).
				ExpectError(tt.wantErrorRE).
				Run()
		})
	}
}

func TestExec(t *testing.T) {
	tests := []struct {
		name        string
		pipe        func(db *sql.DB) strm.Pipe
		send        []interface{}
		senderError error

		want        []interface{}
		wantQuery   func(mock sqlmock.Sqlmock)
		wantErrorRE string
	}{
		{
			name: "executes query",
			pipe: func(db *sql.DB) strm.Pipe {
				return PSQL.Exec(
					db,
					"insert into table values($1,$2)", 1, 2,
				)
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
			pipe: func(db *sql.DB) strm.Pipe {
				return PSQL.Exec(
					db,
					"insert into table values($1, $2)",
					Field("not found"),
				)
			},
			send:        []interface{}{1},
			wantErrorRE: `strmsql.Dialect.Exec.* invalid type 'int' for field: "not found"$`,
		},
		{
			name: "returns error when query errors",
			pipe: func(db *sql.DB) strm.Pipe {
				return PSQL.Exec(
					db,
					"insert into table values($1)",
					1,
				)
			},
			send: []interface{}{1},
			wantQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(`insert into table values($1)`)).
					WithArgs(1).
					WillReturnError(errors.New("query error"))
			},
			wantErrorRE: `strmsql.Dialect.Exec.* query error`,
		},
		{
			name: "returns error when sender errors",
			pipe: func(db *sql.DB) strm.Pipe {
				return PSQL.Exec(
					db,
					"insert into table values($1,$2)",
					1, 2,
				)
			},
			send:        []interface{}{1, 2, 3},
			senderError: errors.New("sender error"),

			wantQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(`insert into table values($1,$2)`)).
					WithArgs(1, 2).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErrorRE: "strmsql.Dialect.Exec.* sender error$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}

			if tt.wantQuery != nil {
				tt.wantQuery(mock)
			}
			st := strmtest.New(t, tt.pipe(db))
			for _, s := range tt.send {
				st.Send(s).WithSenderError(tt.senderError)
			}
			st.ExpectFull(tt.want...).
				ExpectError(tt.wantErrorRE).
				Run()
		})
	}
}
