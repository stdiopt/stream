package strmutil

import (
	"errors"
	"io/ioutil"
	"net/http"
)

// HTTPGet receives a stream of urls performs a call and sends the content as []byte
func HTTPGet(hdr http.Header) ProcFunc {
	return func(p Proc) error {
		return p.Consume(func(v interface{}) error {
			s, ok := v.(string)
			if !ok {
				return errors.New("needs a string")
			}
			res, err := http.Get(s)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			data, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return err
			}
			return p.Send(data)
		})
	}
}
