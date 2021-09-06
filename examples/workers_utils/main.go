package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"os"
	"path/filepath"

	"github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmio"
	"github.com/stdiopt/stream/strmjson"
	"github.com/stdiopt/stream/strmrefl"
	"github.com/stdiopt/stream/strmutil"
	"github.com/stdiopt/stream/x/strmhttp"
)

var ctxKey string = "k"

func main() {
	err := stream.Run(
		strmhttp.Get("https://randomuser.me/api/?results=100"),
		strmjson.Decode(nil),
		strmrefl.Extract("results"),
		strmrefl.Unslice(),
		strmrefl.Extract("picture", "thumbnail"),
		stream.Workers(32,
			stream.S(func(s stream.Sender, url string) error {
				return stream.RunWithContext(s.Context(),
					strmhttp.Get(url),
					HTTPDownload(url),
					strmutil.Pass(s),
				)
			}),
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

func HTTPDownload(url string) stream.Pipe {
	return stream.Func(func(p stream.Proc) error {
		r := strmio.AsReader(p)
		data, err := ioutil.ReadAll(r)
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
