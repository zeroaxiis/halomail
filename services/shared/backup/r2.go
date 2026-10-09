package backup

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type ObjectStore interface {
	Put(context.Context, string, io.Reader, string) error
	Get(context.Context, string) (io.ReadCloser, error)
	List(context.Context, string) ([]string, error)
	Delete(context.Context, string) error
}

func (r *R2) Delete(ctx context.Context, key string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(r.bucket), Key: aws.String(key)})
	return err
}

type R2 struct {
	client *s3.Client
	bucket string
}

func NewR2(c Config) *R2 {
	client := s3.NewFromConfig(aws.Config{
		Region:                     c.Region,
		Credentials:                credentials.NewStaticCredentialsProvider(c.AccessKey, c.SecretKey, ""),
		HTTPClient:                 &http.Client{Timeout: 10 * time.Minute},
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	}, func(o *s3.Options) { o.BaseEndpoint = aws.String(c.Endpoint); o.UsePathStyle = true })
	return &R2{client: client, bucket: c.Bucket}
}

func (r *R2) Put(ctx context.Context, key string, body io.Reader, contentType string) error {
	_, err := transfermanager.New(r.client).UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket: aws.String(r.bucket), Key: aws.String(key), Body: body,
		ContentType: aws.String(contentType), CacheControl: aws.String("no-store"),
	})
	return err
}

func (r *R2) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := r.client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(r.bucket), Key: aws.String(key)})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (r *R2) List(ctx context.Context, prefix string) ([]string, error) {
	pages := s3.NewListObjectsV2Paginator(r.client, &s3.ListObjectsV2Input{Bucket: aws.String(r.bucket), Prefix: aws.String(prefix)})
	var keys []string
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, object := range page.Contents {
			keys = append(keys, aws.ToString(object.Key))
		}
	}
	return keys, nil
}
