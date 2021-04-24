package strmutil

import "github.com/stdiopt/stream"

// End just a type used in several group streams to send the group
var End = struct{}{}

type (
	ProcFunc = stream.ProcFunc
	Proc     = stream.Proc
)
