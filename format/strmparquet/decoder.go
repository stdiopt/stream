package strmparquet

import (
	"os"

	goparquet "github.com/fraugster/parquet-go"
	"github.com/fraugster/parquet-go/floor"
	"github.com/fraugster/parquet-go/floor/interfaces"
	"github.com/fraugster/parquet-go/parquet"
	"github.com/fraugster/parquet-go/parquetschema"
	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/drow"
)

// DecodeFile receives a string path and outputs T.
// Reason for lack of Decode is that we need a ReadSeeker
/*func DecodeFile(sample interface{}) strm.Pipe {
	typ := reflect.TypeOf(sample)
	return strm.S(func(p strm.Sender, path string) error {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		pr, err := goparquet.NewFileReader(f)
		if err != nil {
			return err
		}
		fr := floor.NewReader(pr)
		for fr.Next() {
			v := reflect.New(typ)
			if err := fr.Scan(v.Interface()); err != nil {
				return err
			}
			if err := p.Send(v.Elem().Interface()); err != nil {
				return err
			}
		}
		return nil
	})
}*/

// DecodeFile receives a path as string and decodes parquet as a drow.Row
func DecodeFile() strm.Pipe {
	return strm.S(func(p strm.Sender, path string) error {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		pr, err := goparquet.NewFileReader(f)
		if err != nil {
			return err
		}
		def := pr.GetSchemaDefinition()
		fr := floor.NewReader(pr)
		for fr.Next() {
			r := drow.New()
			if err := fr.Scan(&drowUnmarshaller{def, &r}); err != nil {
				return err
			}
			if err := p.Send(r); err != nil {
				return err
			}
		}
		return nil
	})
}

type drowUnmarshaller struct {
	schema *parquetschema.SchemaDefinition
	row    *drow.Row
}

func (u *drowUnmarshaller) UnmarshalParquet(obj interfaces.UnmarshalObject) error {
	data := obj.GetData()
	for _, ch := range u.schema.RootColumn.Children {
		name := ch.SchemaElement.Name
		v, ok := data[name]
		if !ok {
			continue
		}
		// Check more types like time
		if *ch.SchemaElement.Type == parquet.Type_BYTE_ARRAY && *ch.SchemaElement.ConvertedType == parquet.ConvertedType_UTF8 {
			v = string(v.([]byte))
		}
		u.row.SetOrAdd(name, v)
	}
	return nil
}
