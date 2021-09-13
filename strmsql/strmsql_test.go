package strmsql

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/drow"
	"github.com/stdiopt/stream/strmtest"
)

func TestDB_BatchInsert(t *testing.T) {
	now := time.Now()
	type fields struct {
		dialect Dialecter
	}
	type args struct {
		qry  string
		opts []batchInsertOpt
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		send        []interface{}
		senderError error

		expect  func(sqlmock.Sqlmock)
		want    []interface{}
		wantErr string
	}{
		{
			name:   "execute batch insert with []interface{}",
			fields: fields{dialect: PSQL{}},
			args: args{
				qry: "insert into record values",
				opts: []batchInsertOpt{
					BatchOptions.WithSize(2),
				},
			},
			expect: func(s sqlmock.Sqlmock) {
				s.ExpectExec(`insert into record values\(\$1,\$2\)`).
					WithArgs(1, 2).
					WillReturnResult(driver.ResultNoRows)
			},
			send: []interface{}{
				[]interface{}{1, 2},
			},
			want: []interface{}{
				[]interface{}{1, 2},
			},
		},
		{
			name:   "execute batch insert with drow.Row",
			fields: fields{dialect: PSQL{}},
			args: args{
				qry: "insert into record values",
				opts: []batchInsertOpt{
					BatchOptions.WithSize(2),
				},
			},
			expect: func(s sqlmock.Sqlmock) {
				s.ExpectExec(`insert into record values\(\$1,\$2\)`).
					WithArgs(1, 2).
					WillReturnResult(driver.ResultNoRows)
			},
			send: []interface{}{
				*drow.New().
					SetOrAdd("field1", 1).
					SetOrAdd("field2", 2),
			},
			want: []interface{}{
				*drow.New().
					SetOrAdd("field1", 1).
					SetOrAdd("field2", 2),
			},
		},
		{
			name:   "execute batch insert with multiple drow.Row",
			fields: fields{dialect: PSQL{}},
			args: args{
				qry: "insert into record values",
				opts: []batchInsertOpt{
					BatchOptions.WithSize(2),
				},
			},
			expect: func(s sqlmock.Sqlmock) {
				s.ExpectExec(`insert into record values\(\$1,\$2\), \(\$3,\$4\)`).
					WithArgs(1, 2, 3, 4).
					WillReturnResult(driver.ResultNoRows)
			},
			send: []interface{}{
				*drow.New().
					SetOrAdd("field1", 1).
					SetOrAdd("field2", 2),
				*drow.New().
					SetOrAdd("field1", 3).
					SetOrAdd("field2", 4),
			},
			want: []interface{}{
				*drow.New().
					SetOrAdd("field1", 1).
					SetOrAdd("field2", 2),
				*drow.New().
					SetOrAdd("field1", 3).
					SetOrAdd("field2", 4),
			},
		},
		{
			name:   "generate DDL with multiple drow.Row",
			fields: fields{dialect: PSQL{}},
			args: args{
				qry: "insert into record values",
				opts: []batchInsertOpt{
					BatchOptions.WithSize(2),
					BatchOptions.WithAutoDDL("record"),
				},
			},
			expect: func(s sqlmock.Sqlmock) {
				s.ExpectExec(`create table if not exists "record".*"field1".*"field2"`).
					WillReturnResult(driver.ResultNoRows)
				s.ExpectExec(`insert into record values\(\$1,\$2,\$3,\$4,\$5,\$6\), .*`).
					WithArgs(
						1, 2, new(int), uint(4), "5", now,
						7, 8, new(int), uint(10), "11", now,
					).
					WillReturnResult(driver.ResultNoRows)
			},
			send: []interface{}{
				*drow.New().
					SetOrAdd("field1", 1).
					SetOrAdd("field2", 2).
					SetOrAdd("field3", new(int)).
					SetOrAdd("field4", uint(4)).
					SetOrAdd("field5", "5").
					SetOrAdd("field6", now),
				*drow.New().
					SetOrAdd("field1", 7).
					SetOrAdd("field2", 8).
					SetOrAdd("field3", new(int)).
					SetOrAdd("field4", uint(10)).
					SetOrAdd("field5", "11").
					SetOrAdd("field6", now),
			},
			want: []interface{}{
				*drow.New().
					SetOrAdd("field1", 1).
					SetOrAdd("field2", 2).
					SetOrAdd("field3", new(int)).
					SetOrAdd("field4", uint(4)).
					SetOrAdd("field5", "5").
					SetOrAdd("field6", now),
				*drow.New().
					SetOrAdd("field1", 7).
					SetOrAdd("field2", 8).
					SetOrAdd("field3", new(int)).
					SetOrAdd("field4", uint(10)).
					SetOrAdd("field5", "11").
					SetOrAdd("field6", now),
			},
		},
		{
			name:   "errors with autoDDL for wrong type",
			fields: fields{dialect: PSQL{}},
			args: args{
				qry: "insert into record values",
				opts: []batchInsertOpt{
					BatchOptions.WithSize(2),
					BatchOptions.WithAutoDDL("record"),
				},
			},
			expect: func(s sqlmock.Sqlmock) {},
			send: []interface{}{
				[]interface{}{1, 2},
			},
			wantErr: "strmsql.* auto DDL only supported on drow",
		},
		{
			name:   "errors when ddl qry errors",
			fields: fields{dialect: PSQL{}},
			args: args{
				qry: "insert into record values",
				opts: []batchInsertOpt{
					BatchOptions.WithSize(2),
					BatchOptions.WithAutoDDL("record"),
				},
			},
			expect: func(s sqlmock.Sqlmock) {
				s.ExpectExec(".*").WillReturnError(errors.New("query error"))
			},
			send: []interface{}{
				*drow.New().
					SetOrAdd("field1", 1).
					SetOrAdd("field2", 2),
			},
			wantErr: "strmsql.* query error",
		},
		{
			name:   "errors when insert qry errors",
			fields: fields{dialect: PSQL{}},
			args: args{
				qry: "insert into record values",
				opts: []batchInsertOpt{
					BatchOptions.WithSize(2),
				},
			},
			expect: func(s sqlmock.Sqlmock) {
				s.ExpectExec(".*").WillReturnError(errors.New("query error"))
			},
			send: []interface{}{
				[]interface{}{1, 2},
			},
			want: []interface{}{
				[]interface{}{1, 2},
			},
			wantErr: "strmsql.* query error",
		},
		{
			name:   "errors when invalid type sent",
			fields: fields{dialect: PSQL{}},
			args: args{
				qry: "insert into record values",
				opts: []batchInsertOpt{
					BatchOptions.WithSize(2),
				},
			},
			send: []interface{}{
				"invalid type",
			},
			wantErr: "strmsql.* invalid type string",
		},
		{
			name:   "errors when invalid type sent",
			fields: fields{dialect: PSQL{}},
			args: args{
				qry: "insert into record values",
				opts: []batchInsertOpt{
					BatchOptions.WithSize(2),
				},
			},
			send: []interface{}{
				[]interface{}{1, 2},
				"invalid type",
			},
			want: []interface{}{
				[]interface{}{1, 2},
			},
			wantErr: "strmsql.* invalid type string",
		},
		{
			name:   "errors when sender errors",
			fields: fields{dialect: PSQL{}},
			args: args{
				qry: "insert into record values",
				opts: []batchInsertOpt{
					BatchOptions.WithSize(2),
				},
			},
			expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(".*").WillReturnResult(driver.ResultNoRows)
			},
			send: []interface{}{
				[]interface{}{1, 2},
			},
			senderError: errors.New("sender error"),
			wantErr:     "strmsql.* sender error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			if tt.expect != nil {
				tt.expect(mock)
			}

			d := New(db, tt.fields.dialect)
			pp := d.BatchInsert(tt.args.qry, tt.args.opts...)
			if pp == nil {
				t.Errorf("BatchInsert() is nil = %v. want %v", pp == nil, false)
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

func TestDB_Exec(t *testing.T) {
	type fields struct {
		db      *sql.DB
		dialect Dialecter
	}
	type args struct {
		qry string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		send        []interface{}
		senderError error

		expect  func(sqlmock.Sqlmock)
		want    []interface{}
		wantErr string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			if tt.expect != nil {
				tt.expect(mock)
			}

			d := New(db, tt.fields.dialect)
			pp := d.Exec(tt.args.qry)
			if pp == nil {
				t.Errorf("Exec() is nil = %v. want %v", pp == nil, false)
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

func TestDB_Query(t *testing.T) {
	type fields struct {
		db      *sql.DB
		dialect Dialecter
	}
	type args struct {
		qry string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   strm.Pipe
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DB{
				db:      tt.fields.db,
				dialect: tt.fields.dialect,
			}
			if got := d.Query(tt.args.qry); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DB.Query() = %v, want %v", got, tt.want)
			}
		})
	}
}
