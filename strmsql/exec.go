package strmsql

import "database/sql"

func Exec(db *sql.DB, qry string) strm.Pipe {
	return strm.S(func(_ strm.Sender, params []interface{}) error {
		_, err := db.Exec(qry, params...)
		return err
	})
}
