package strmhttp

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmio"
)

func WithHeader(k, v string) RequestFunc {
	return func(r *http.Request) {
		r.Header.Add(k, v)
	}
}

type RequestFunc func(r *http.Request)

// GetResponse receives url as string, performs a get request and sends the
// response
func GetResponse(reqFunc ...RequestFunc) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		return p.Consume(func(v interface{}) error {
			url, ok := v.(string)
			if !ok {
				return errors.New("input should be a string")
			}
			req, err := http.NewRequestWithContext(p.Context(), http.MethodGet, url, nil)
			if err != nil {
				return fmt.Errorf("GetResponse: %w", err)
			}

			for _, fn := range reqFunc {
				fn(req)
			}

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("GetResponse: %w", err)
			}
			return p.Send(res)
		})
	})
}

// Get receives a stream of urls performs a get request and sends the
// content as []byte returns error on status < 200 || >= 400
func Get(reqFunc ...RequestFunc) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		return p.Consume(func(url string) error {
			req, err := http.NewRequestWithContext(p.Context(), http.MethodGet, url, nil)
			if err != nil {
				return fmt.Errorf("Get: %w", err)
			}

			for _, fn := range reqFunc {
				fn(req)
			}

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			if res.StatusCode < 200 || res.StatusCode >= 400 {
				return fmt.Errorf("http status code: %d - %v", res.StatusCode, res.Status)
			}

			w := strmio.AsWriter(p)
			_, err = io.Copy(w, res.Body)
			return err
		})
	})
}
