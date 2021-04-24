package main

import (
	"context"
	"encoding/base64"
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

var ctxKey string = "k"

func main() {
	err := stream.Run(
		strmutil.Value("https://randomuser.me/api/?results=100"),
		contextWithValue(ctxKey, "first"),

		strmutil.HTTPGetResponse(nil),
		contextWithValue(ctxKey, "after response"),
		strmutil.Field("Body"),
		contextWithValue(ctxKey, "after Field"),
		strmutil.IOWithReader(),
		strmutil.JSONParse(nil),
		contextWithValue(ctxKey, "in reader"),
		strmutil.Field("results"),
		contextWithValue(ctxKey, "after results"),
		strmutil.Unslice(),
		strmutil.Field("picture.thumbnail"),
		stream.Workers(32,
			strmutil.HTTPGetResponse(nil),
			HTTPDownload,
		),
		strmutil.JSONDump(os.Stdout),
	)
	if err != nil {
		log.Println("err:", err)
	}
}

func contextTrace(p stream.Proc) error {
	return p.Consume(func(ctx context.Context, v interface{}) error {
		log.Println("  Parent: ", p.Context().Value(ctxKey))
		log.Println("  Message:", ctx.Value(ctxKey))
		return p.Send(ctx, v)
	})
}

func contextWithValue(ck, cv interface{}) stream.ProcFunc {
	return func(p stream.Proc) error {
		return p.Consume(func(ctx context.Context, v interface{}) error {
			ctx = context.WithValue(ctx, ck, cv)
			return p.Send(ctx, v)
		})
	}
}

type HTTPDownloadOutput struct {
	URL  string
	Data string
}

func HTTPDownload(p stream.Proc) error {
	return p.Consume(func(ctx context.Context, v interface{}) error {
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

		return p.Send(ctx, HTTPDownloadOutput{
			URL:  url,
			Data: fmt.Sprintf("data:%s;base64,%s", m, b64data),
		})
	})
}
