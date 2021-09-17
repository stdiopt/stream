package strms3

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/go-cmp/cmp"
	"github.com/stdiopt/stream/utils/strmtest"
)

func TestUpload(t *testing.T) {
	type args struct {
		s3url string
	}
	tests := []struct {
		name string
		args args

		send          []interface{}
		senderError   error
		uploaderError error

		want []interface{}

		wantBucket  string
		wantKey     string
		wantBody    []byte
		wantErrorRE string
	}{

		{
			name: "upload data",
			args: args{s3url: "s3://bucket/key"},
			send: []interface{}{
				[]byte("test"),
			},

			want: []interface{}{
				[]byte("test"),
			},
			wantBucket: "bucket",
			wantKey:    "key",
			wantBody:   []byte("test"),
		},
		{
			name: "upload data from multiple sends",
			args: args{s3url: "s3://bucket/key"},
			send: []interface{}{
				[]byte("test 1"),
				[]byte("test 2"),
			},

			want: []interface{}{
				[]byte("test 1"),
				[]byte("test 2"),
			},
			wantBucket: "bucket",
			wantKey:    "key",
			wantBody:   []byte("test 1test 2"),
		},
		{
			name: "returns error on malformed url",
			args: args{s3url: "bucket/key"},
			send: []interface{}{[]byte("test 1")},

			wantErrorRE: "strms3.Upload.* malformed url, should be in 's3://{bucket}/{key/key}' form$",
		},
		{
			name: "returns error on invalid send type",
			args: args{s3url: "s3://bucket/key"},
			send: []interface{}{1},

			wantBucket:  "bucket",
			wantKey:     "key",
			wantErrorRE: `strms3.Upload.* invalid type, want '\[\]uint8' but got 'int'$`,
		},
		{
			name: "returns error on parsing url",
			args: args{s3url: "%20://test//--"},
			send: []interface{}{[]byte("test 1")},

			wantErrorRE: "strms3.Upload.* first path segment in URL cannot contain colon$",
		},
		{
			name:          "returns error on uploader error",
			args:          args{s3url: "s3://bucket/key"},
			send:          []interface{}{[]byte("test 1")},
			uploaderError: errors.New("uploader test error"),

			wantErrorRE: "strms3.Upload.* uploader test error$",
		},
		{
			name:        "returns error on sender error",
			args:        args{s3url: "s3://bucket/key"},
			send:        []interface{}{[]byte("test 1")},
			senderError: errors.New("sender test error"),

			wantErrorRE: "strms3.Upload.* sender test error$",
			wantBucket:  "bucket",
			wantKey:     "key",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upl := s3mock{uploaderError: tt.uploaderError}

			pp := Upload(&upl, tt.args.s3url)
			if pp == nil {
				t.Errorf("Limit() is nil = %v, want %v", pp == nil, false)
			}
			st := strmtest.New(t, pp)
			for _, s := range tt.send {
				st.Send(s).WithSenderError(tt.senderError)
			}
			st.ExpectFull(tt.want...).
				ExpectError(tt.wantErrorRE).
				Run()

			if upl.bucket != tt.wantBucket {
				t.Errorf("wrong bucket\nwant: %v\n got: %v\n", tt.wantBucket, upl.bucket)
			}

			if upl.key != tt.wantKey {
				t.Errorf("wrong key\nwant: %v\n got: %v\n", tt.wantKey, upl.key)
			}

			if diff := cmp.Diff(tt.wantBody, upl.body); diff != "" {
				t.Error("wrong body output\n- want + got\n", diff)
			}
		})
	}
}

type s3mock struct {
	uploaderError error

	bucket string
	key    string
	body   []byte
}

func (s *s3mock) UploadWithContext(
	ctx context.Context,
	input *s3manager.UploadInput,
	opts ...func(*s3manager.Uploader),
) (*s3manager.UploadOutput, error) {

	if s.uploaderError != nil {
		return nil, s.uploaderError
	}
	s.bucket = aws.StringValue(input.Bucket)
	s.key = aws.StringValue(input.Key)

	data, err := io.ReadAll(input.Body)
	if err != nil {
		return nil, err
	}
	s.body = data

	return &s3manager.UploadOutput{}, nil
}
