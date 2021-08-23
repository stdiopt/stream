package strms3

import (
	"io"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmio"
)

func Uploader(s3url string, conf *aws.Config) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		sess, err := session.NewSession()
		if err != nil {
			return err
		}
		cli := s3.New(sess, conf)
		upl := s3manager.NewUploaderWithClient(cli)

		r := io.TeeReader(strmio.AsReader(p), strmio.AsWriter(p))

		u, err := url.Parse(s3url)
		if err != nil {
			return err
		}

		_, err = upl.UploadWithContext(
			p.Context(),
			&s3manager.UploadInput{
				Bucket: aws.String(u.Host),
				Key:    aws.String(u.Path),
				Body:   r,
			},
		)
		return err
	})
}
