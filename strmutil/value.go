package strmutil

// Value returns a ProcFunc that sends a single value v.
func Value(v interface{}) ProcFunc {
	return func(p Proc) error {
		return p.Send(p.Context(), v)
	}
}
