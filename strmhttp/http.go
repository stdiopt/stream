package strmhttp

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/stdiopt/stream"
)

func WithHeader(k, v string) RequestFunc {
	return func(r *http.Request) {
		r.Header.Add(k, v)
	}
}

type RequestFunc func(r *http.Request)

// GetResponse receives url as string, performs a get request and sends the
// response
func GetResponse(reqFunc ...RequestFunc) stream.PipeFunc {
	return stream.Func(func(p stream.Proc) error {
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
// content as []byte
func Get(reqFunc ...RequestFunc) stream.PipeFunc {
	return stream.Func(func(p stream.Proc) error {
		return p.Consume(func(v interface{}) error {
			url, ok := v.(string)
			if !ok {
				return errors.New("input should be a string")
			}
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
			data, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return fmt.Errorf("Get: %w", err)
			}
			return p.Send(data)
		})
	})
}
