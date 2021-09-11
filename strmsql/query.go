package strmsql

import (
	"database/sql"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/x/strmdrow"
)

func Query(db *sql.DB, qry string) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		var hdr *strmdrow.Header
		return p.Consume(func(params []interface{}) error {
			rows, err := db.QueryContext(p.Context(), qry, params...)
			if err != nil {
				return err
			}

			if hdr == nil {
				nhdr, err := drowHeader(rows)
				if err != nil {
					return err
				}
				hdr = nhdr
			}
			for rows.Next() {
				row, err := drowScan(hdr, rows)
				if err != nil {
					return err
				}
				if err := p.Send(row); err != nil {
					return err
				}
			}
			return nil
		})
	})
}
