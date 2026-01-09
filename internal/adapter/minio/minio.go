package minio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/url"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const MaxRetries = 5

type Config struct {
	MinioRootUser     string
	MinioRootPassword string
	MinioURL          string
	MinioBucketName   string
	MinioRegion       string
	MinioUseSSL       bool
}

type Minio struct {
	client     *minio.Client
	bucketName string
	region     string
}

func New(ctx context.Context, cfg *Config) (*Minio, error) {
	const operation = "minio.New"

	minioClient, err := minio.New(cfg.MinioURL, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioRootUser, cfg.MinioRootPassword, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	m := &Minio{
		client:     minioClient,
		bucketName: cfg.MinioBucketName,
		region:     cfg.MinioRegion,
	}

	if err = m.CreateBucket(ctx); err != nil {
		slog.Error(err.Error())
	}

	return m, nil
}

func (m *Minio) CreateBucket(ctx context.Context) error {
	const operation = "minio.CreateBucket"

	exists, err := m.client.BucketExists(ctx, m.bucketName)
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}

	if !exists {
		err = m.client.MakeBucket(ctx, m.bucketName, minio.MakeBucketOptions{
			Region: m.region,
		})
		if err != nil {
			return fmt.Errorf("%s: %w", operation, err)
		}
	}

	return nil
}

func (m *Minio) Get(path string) ([]byte, error) {
	const operation = "minio.Get"

	ctx := context.Background()

	decodedPath, err := url.PathUnescape(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	object, err := m.client.GetObject(ctx, m.bucketName, decodedPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return data, nil
}

func (m *Minio) Upload(path string, data []byte) error {
	const operation = "minio.Upload"

	ctx := context.Background()

	decodedPath, err := url.PathUnescape(path)
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}

	reader := bytes.NewReader(data)
	_, err = m.client.PutObject(ctx, m.bucketName, decodedPath, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}

	return nil
}

func (m *Minio) Delete(path string) error {
	const operation = "minio.Delete"

	ctx := context.Background()

	decodedPath, err := url.PathUnescape(path)
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}

	err = m.client.RemoveObject(ctx, m.bucketName, decodedPath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}

	return nil
}

func (m *Minio) Check(path string) error {
	const operation = "minio.Check"

	ctx := context.Background()

	decodedPath, err := url.PathUnescape(path)
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}

	_, err = m.client.StatObject(ctx, m.bucketName, decodedPath, minio.StatObjectOptions{})
	if err != nil {
		return fmt.Errorf("%s: %w, path - %s", operation, err, path)
	}

	return nil
}
