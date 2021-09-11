package strmsql

import "github.com/stdiopt/stream/x/strmdrow"

type Dialecter interface {
	QryDDL(name string, row strmdrow.Row) string
	QryBatch(qry string, nparams, nrows int) string
}
