package strmsql

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmtest"
)

func TestDialect_InsertBatch(t *testing.T) {
	type args struct {
		// db        *sql.DB
		batchSize int
		qry       string
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
			name: "executes PSQL batch query",
			d:    PSQL,
			args: args{
				batchSize: 2,
				qry:       "insert into table values",
			},
			send: []interface{}{
				[]interface{}{1, 2},
			},
			wantQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(`insert into table values($1,$2)`)).
					WithArgs(1, 2).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name: "executes MYSQL batch insert",
			d:    MySQL,
			args: args{
				batchSize: 2,
				qry:       "insert into table values",
			},
			send: []interface{}{
				[]interface{}{"1.1", "2.1"},
				[]interface{}{"2.1", "2.2"},
				[]interface{}{"3.1", "2.3"},
				[]interface{}{"4.1", "2.4"},
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
			name: "returns error when query errors",
			d:    PSQL,
			args: args{
				batchSize: 1,
				qry:       "insert into table values",
			},
			send: []interface{}{
				[]interface{}{1},
			},
			wantQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(`insert into table values($1)`)).
					WithArgs(1).
					WillReturnError(errors.New("query error"))
			},
			wantErr: `strmsql.Dialect.InsertBatch.* query error`,
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
		db  *sql.DB
		qry string
	}
	tests := []struct {
		name string
		d    Dialect
		args args
		want strm.Pipe
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.Exec(tt.args.db, tt.args.qry); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Dialect.Exec() = %v, want %v", got, tt.want)
			}
		})
	}
}
