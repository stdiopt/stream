package strms3

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/utils/strmio"
)

// Upload uses s3manager from aws sdk and uploads into an s3 bucket
// url format: (s3://{bucket}/{key/key})
// https://pkg.go.dev/github.com/aws/aws-sdk-go@v1.40.30/service/s3/s3manager#NewUploader
func Upload(svc s3iface.S3API, s3url string) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		ctx := p.Context()
		bucket, key, err := s3urlParse(s3url)
		if err != nil {
			return err
		}
		upl := s3manager.NewUploaderWithClient(svc)

		pr := strmio.AsReader(p)
		defer pr.Close()
		r := io.TeeReader(pr, strmio.AsWriter(p))

		uploadInput := &s3manager.UploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   r,
		}
		_, err = upl.UploadWithContext(ctx, uploadInput)
		return err
	})
}

// Download receives a s3url as string and produces []byte
func Download(svc s3iface.S3API) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		dn := s3manager.NewDownloaderWithClient(svc)
		ctx := p.Context()
		return p.Consume(func(s3url string) error {
			bucket, key, err := s3urlParse(s3url)
			if err != nil {
				return err
			}

			getObjectInput := &s3.GetObjectInput{
				Bucket: aws.String(bucket),
				Key:    aws.String(key),
			}

			buf := aws.WriteAtBuffer{}
			if _, err := dn.DownloadWithContext(ctx, &buf, getObjectInput); err != nil {
				return err
			}

			rd := bytes.NewReader(buf.Bytes())
			wr := strmio.AsWriter(p)

			_, err = io.Copy(wr, rd)
			return err
		})
	})
}

// List receives a s3url as string and produces a string as s3urls
func List(s3cli s3iface.S3API) strm.Pipe {
	return strm.S(func(s strm.Sender, s3url string) error {
		ctx := s.Context()
		bucket, key, err := s3urlParse(s3url)
		if err != nil {
			return err
		}
		var contToken *string
		for {
			listObjects := &s3.ListObjectsV2Input{
				Bucket:            aws.String(bucket),
				Prefix:            aws.String(key),
				ContinuationToken: contToken,
			}
			res, err := s3cli.ListObjectsV2WithContext(ctx, listObjects)
			if err != nil {
				return err
			}
			for _, r := range res.Contents {
				objurl := fmt.Sprintf("s3://%s/%s", bucket, *r.Key)
				if err := s.Send(objurl); err != nil {
					return err
				}
			}
			if res.ContinuationToken == nil {
				break
			}
			contToken = res.ContinuationToken
		}
		return nil
	})
}

func s3urlParse(s3url string) (bucket string, ket string, err error) {
	u, err := url.Parse(s3url)
	if err != nil {
		return "", "", err
	}
	if u.Scheme != "s3" || u.Host == "" {
		return "", "", fmt.Errorf("malformed url, should be in 's3://{bucket}/{key/...}' form")
	}
	return u.Host, strings.TrimPrefix(u.Path, "/"), nil
}
