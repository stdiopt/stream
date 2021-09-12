package drow

import "reflect"

func StructType(d Row) (reflect.Type, error) {
	fields := []reflect.StructField{}

	for i := range d.Values {
		h := d.Header(i)
		colName, err := normalizeGoField(h.Name)
		if err != nil {
			return nil, err
		}
		fields = append(fields, reflect.StructField{
			Name: colName,
			Type: h.Type,
			Tag:  h.Tag,
		})
	}

	return reflect.StructOf(fields), nil
}
