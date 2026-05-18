package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/devrapture/omni/internal/config"
)

type R2Storage struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
}

func NewR2Storage(ctx context.Context, cfg *config.Config) (*R2Storage, error) {
	config, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.R2_ACCESS_KEY, cfg.R2_SECRET_KEY, "")),
		awsconfig.WithRegion("auto"),
	)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(config, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2_ACCOUNT_ID))
	})
	return &R2Storage{
		client:        client,
		presignClient: s3.NewPresignClient(client),
		bucket:        cfg.R2_BUCKET_NAME,
	}, nil
}

func (s *R2Storage) PresignPutObject(ctx context.Context, key string, expiresIn time.Duration) (string, error) {
	result, err := s.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(po *s3.PresignOptions) {
		po.Expires = expiresIn
	})
	if err != nil {
		return "", err
	}

	return result.URL, nil
}

func (s *R2Storage) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return result.Body, nil
}

func (s *R2Storage) DeleteObject(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	return nil
}
