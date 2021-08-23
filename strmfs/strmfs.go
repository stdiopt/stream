package strmfs

/*func solvePath(v interface{}, ps ...interface{}) (string, error) {
	parts := []string{}
	for _, p := range ps {
		switch p := p.(type) {
		case string:
			parts = append(parts, p)
		case strmutil.ParamFunc:
			r, err := p(v)
			if err != nil {
				return "", err
			}
			parts = append(parts, fmt.Sprint(r))
		}
	}
	return filepath.Join(parts...), nil
}*/

// What if we use a single solvable param that produces a single path
// either from one or another field?
/*
	type My struct {
		Name string
		Field string
	}
	strm.Run(
		strmutil.Value(My{...})
		strmio.ListFiles(strmrefl.Input("Name"), strmrefl.Input("Field")),
	)


*/
