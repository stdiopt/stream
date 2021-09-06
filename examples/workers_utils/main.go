package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmjson"
	"github.com/stdiopt/stream/strmrefl"
	"github.com/stdiopt/stream/x/strmhttp"
)

var ctxKey string = "k"

func main() {
	err := stream.Run(
		strmhttp.Get("https://randomuser.me/api/?results=100"),
		strmjson.Decode(nil),
		strmrefl.Extract("results"),
		strmrefl.Unslice(),
		strmrefl.Extract("picture.thumbnail"),
		stream.Workers(32,
			strmhttp.GetResponse(),
			stream.Func(HTTPDownload),
		),
		strmjson.Dump(os.Stdout),
	)
	if err != nil {
		log.Println("err:", err)
	}
}

type HTTPDownloadOutput struct {
	URL  string
	Data string
}

func HTTPDownload(p stream.Proc) error {
	return p.Consume(func(v interface{}) error {
		res, ok := v.(*http.Response)
		if !ok {
			return fmt.Errorf("needs a *http.Response but got: %T", v)
		}
		defer res.Body.Close()
		url := res.Request.URL.String()
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		b64data := base64.StdEncoding.EncodeToString(data)
		m := mime.TypeByExtension(filepath.Ext(url))

		return p.Send(HTTPDownloadOutput{
			URL:  url,
			Data: fmt.Sprintf("data:%s;base64,%s", m, b64data),
		})
	})
}
