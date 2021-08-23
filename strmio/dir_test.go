package strmio

import (
	"testing"

	"github.com/stdiopt/stream/strmtest"
)

func TestListFiles(t *testing.T) {
	t.Run("list files", strmtest.Case{
		PipeFunc: ListFiles("*"),
		Sends: []strmtest.Send{
			{
				Value: "./testdata",
				Want: []interface{}{
					"testdata/01/test.txt",
					"testdata/01/test2.txt",
					"testdata/02/subdir/test.csv",
					"testdata/02/test.json",
				},
			},
		},
	}.Test)
}
