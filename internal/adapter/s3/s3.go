package s3

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const MaxRetries = 5

type Config struct {
	S3RootUser     string
	S3RootPassword string
	S3URL          string
	S3BucketName   string
	S3Region       string
}

type S3 struct {
	s3Client   *s3.S3
	bucketName string
}

func New(cfg *Config) (*S3, error) {
	const operation = "s3.NewS3"

	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(cfg.S3Region),
		Endpoint:         aws.String(cfg.S3URL),
		Credentials:      credentials.NewStaticCredentials(cfg.S3RootUser, cfg.S3RootPassword, ""),
		S3ForcePathStyle: aws.Bool(true),
		MaxRetries:       aws.Int(MaxRetries),
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	s3Client := &S3{
		s3Client:   s3.New(sess),
		bucketName: cfg.S3BucketName,
	}
	if err := s3Client.CreateBucket(); err != nil {
		slog.Error(err.Error())
	}
	return s3Client, nil
}

func (s *S3) CreateBucket() error {
	const operation = "s3.CreateBucket"

	_, err := s.s3Client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(s.bucketName),
	})
	if err != nil {
		_, err = s.s3Client.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(s.bucketName),
		})
		if err != nil {
			return fmt.Errorf("%s: %w", operation, err)
		}
	}

	return nil
}

func (s *S3) Get(path string) ([]byte, error) {
	const operation = "s3.Get"

	output, err := s.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	defer output.Body.Close()

	data, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return data, nil
}

func (s *S3) Upload(path string, data []byte) error {
	const operation = "s3.Upload"

	_, err := s.s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(path),
		Body:   aws.ReadSeekCloser(bytes.NewReader(data)),
	})
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}

	return nil
}

func (s *S3) Delete(path string) error {
	const operation = "s3.Delete"

	_, err := s.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(path),
	})
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}

	return nil
}
