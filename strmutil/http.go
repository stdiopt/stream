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
			url, ok := v.(string)
			if !ok {
				return errors.New("needs a string")
			}
			req, err := http.NewRequestWithContext(p.Context(), "GET", url, nil)
			if err != nil {
				return err
			}

			res, err := http.DefaultClient.Do(req)
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
