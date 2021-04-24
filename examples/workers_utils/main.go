package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmutil"
)

func main() {
	err := stream.Run(
		stream.Line(
			strmutil.Value("https://randomuser.me/api/?results=100"), // just sends the string
			strmutil.HTTPGet(nil),
			strmutil.JSONParse(nil),
			strmutil.Field("results"),
			strmutil.Unslice(),
			strmutil.Field("picture.thumbnail"),
			// Download profile pictures in parallel
			stream.Workers(32, HTTPDownload(nil)),
			strmutil.JSONDump(os.Stdout),
		),
	)
	if err != nil {
		log.Println("err:", err)
	}
}

type URLOutput struct {
	URL  string
	Data string
}

func HTTPDownload(hdr http.Header) stream.ProcFunc {
	return func(p stream.Proc) error {
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
			b64data := base64.StdEncoding.EncodeToString(data)
			m := mime.TypeByExtension(filepath.Ext(s))

			return p.Send(URLOutput{
				URL:  s,
				Data: fmt.Sprintf("data:%s;base64,%s", m, b64data),
			})
		})
	}
}
