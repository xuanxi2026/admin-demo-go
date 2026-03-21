package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOS3Client struct {
	client        *minio.Client
	publicBucket  string
	privateBucket string
}

type MinIOS3Config struct {
	Endpoint      string
	AccessKey     string
	SecretKey     string
	UseSSL        bool
	PublicBucket  string
	PrivateBucket string
}

func NewMinIOS3(cfg MinIOS3Config) (*MinIOS3Client, error) {
	if strings.TrimSpace(cfg.Endpoint) == "" {
		cfg.Endpoint = "s3.amazonaws.com"
	}
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	c := &MinIOS3Client{
		client:        client,
		publicBucket:  cfg.PublicBucket,
		privateBucket: cfg.PrivateBucket,
	}
	if strings.TrimSpace(c.publicBucket) == "" {
		c.publicBucket = "admin-demo-public"
	}
	if strings.TrimSpace(c.privateBucket) == "" {
		c.privateBucket = "admin-demo-private"
	}
	if err = c.ensureBucket(context.Background(), c.publicBucket); err != nil {
		return nil, err
	}
	if err = c.ensureBucket(context.Background(), c.privateBucket); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *MinIOS3Client) List(ctx context.Context, area string) ([]ObjectInfo, error) {
	bucket := c.bucket(area)
	rows := make([]ObjectInfo, 0)
	for obj := range c.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{Recursive: true}) {
		if obj.Err != nil {
			return nil, obj.Err
		}
		rows = append(rows, ObjectInfo{
			Name:      obj.Key,
			Size:      obj.Size,
			UpdatedAt: obj.LastModified,
		})
	}
	return rows, nil
}

func (c *MinIOS3Client) Upload(ctx context.Context, area, filename string, reader io.Reader, size int64) (ObjectInfo, error) {
	name := safeObjectName(filename)
	if name == "" {
		return ObjectInfo{}, fmt.Errorf("invalid filename")
	}
	bucket := c.bucket(area)
	if _, err := c.client.StatObject(ctx, bucket, name, minio.StatObjectOptions{}); err == nil {
		name = fmt.Sprintf("%d_%s", time.Now().UnixMilli(), name)
	}
	info, err := c.client.PutObject(ctx, bucket, name, reader, size, minio.PutObjectOptions{})
	if err != nil {
		return ObjectInfo{}, err
	}
	return ObjectInfo{Name: name, Size: info.Size, UpdatedAt: time.Now()}, nil
}

func (c *MinIOS3Client) Download(ctx context.Context, area, filename string) (io.ReadCloser, error) {
	name := safeObjectName(filename)
	if name == "" {
		return nil, fmt.Errorf("invalid filename")
	}
	bucket := c.bucket(area)
	obj, err := c.client.GetObject(ctx, bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	if _, err = obj.Stat(); err != nil {
		return nil, err
	}
	return obj, nil
}

func (c *MinIOS3Client) Delete(ctx context.Context, area, filename string) error {
	name := safeObjectName(filename)
	if name == "" {
		return fmt.Errorf("invalid filename")
	}
	return c.client.RemoveObject(ctx, c.bucket(area), name, minio.RemoveObjectOptions{})
}

func (c *MinIOS3Client) bucket(area string) string {
	if area == AreaPrivate {
		return c.privateBucket
	}
	return c.publicBucket
}

func (c *MinIOS3Client) ensureBucket(ctx context.Context, name string) error {
	exists, err := c.client.BucketExists(ctx, name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return c.client.MakeBucket(ctx, name, minio.MakeBucketOptions{})
}

func safeObjectName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "\\", "/")
	name = strings.TrimPrefix(name, "/")
	name = strings.ReplaceAll(name, "..", "")
	return name
}
