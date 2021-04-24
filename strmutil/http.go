package strmutil

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// HTTPGetResponse receives url as string, performs a get request and sends the
// response
func HTTPGetResponse(hdr http.Header) ProcFunc {
	return func(p Proc) error {
		return p.Consume(func(ctx context.Context, v interface{}) error {
			url, ok := v.(string)
			if !ok {
				return errors.New("input should be a string")
			}
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				return fmt.Errorf("HTTPGetResponse: %w", err)
			}
			req.Header = hdr
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("HTTPGetResponse: %w", err)
			}
			return p.Send(ctx, res)
		})
	}
}

// HTTPGet receives a stream of urls performs a get request and sends the
// content as []byte
func HTTPGet(hdr http.Header) ProcFunc {
	return func(p Proc) error {
		return p.Consume(func(ctx context.Context, v interface{}) error {
			url, ok := v.(string)
			if !ok {
				return errors.New("input should be a string")
			}
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				return fmt.Errorf("HTTPGet: %w", err)
			}

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			data, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return fmt.Errorf("HTTPGet: %w", err)
			}
			return p.Send(ctx, data)
		})
	}
}
