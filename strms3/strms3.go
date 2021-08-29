package strms3

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmio"
)

type s3uploader interface {
	UploadWithContext(
		ctx aws.Context,
		input *s3manager.UploadInput,
		opts ...func(*s3manager.Uploader),
	) (*s3manager.UploadOutput, error)
}

// Upload uses s3manager from aws sdk and uploads into an s3 bucket
// url format: (s3://{bucket}/{key/key})
// https://pkg.go.dev/github.com/aws/aws-sdk-go@v1.40.30/service/s3/s3manager#NewUploader
func Upload(upl s3uploader, s3url string) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		pr := strmio.AsReader(p)
		defer pr.Close()
		r := io.TeeReader(pr, strmio.AsWriter(p))

		u, err := url.Parse(s3url)
		if err != nil {
			return err
		}
		if u.Scheme != "s3" || u.Host == "" || u.Path == "" {
			return fmt.Errorf("malformed url, should be in 's3://{bucket}/{key/key}' form")
		}
		path := strings.TrimPrefix(u.Path, "/")

		_, err = upl.UploadWithContext(
			p.Context(),
			&s3manager.UploadInput{
				Bucket: aws.String(u.Host),
				Key:    aws.String(path),
				Body:   r,
			},
		)
		return err
	})
}
