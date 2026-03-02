package minio

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	mc            *minio.Client
	bucket        string
	endpoint      string
	useSSL        bool
	publicBaseURL string
}

type FileInfo struct {
	Data        []byte
	ContentType string
	Size        int64
}

func NewClient(endpoint, accessKey, secretKey, bucket string, useSSL bool, publicBaseURL ...string) (*Client, error) {
	mc, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio new client: %w", err)
	}

	ctx := context.Background()
	exists, err := mc.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("minio check bucket: %w", err)
	}
	if !exists {
		if err := mc.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("minio create bucket: %w", err)
		}
	}

	var pubURL string
	if len(publicBaseURL) > 0 {
		pubURL = publicBaseURL[0]
	}

	return &Client{
		mc:            mc,
		bucket:        bucket,
		endpoint:      endpoint,
		useSSL:        useSSL,
		publicBaseURL: pubURL,
	}, nil
}

func (c *Client) Upload(ctx context.Context, folder string, data []byte, contentType string) (string, error) {
	objectName := folder + "/" + uuid.New().String()

	_, err := c.mc.PutObject(ctx, c.bucket, objectName, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("minio put object: %w", err)
	}

	return objectName, nil
}

func (c *Client) Get(ctx context.Context, objectName string) (*FileInfo, error) {
	obj, err := c.mc.GetObject(ctx, c.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("minio get object: %w", err)
	}
	defer obj.Close()

	stat, err := obj.Stat()
	if err != nil {
		return nil, fmt.Errorf("minio stat object: %w", err)
	}

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("minio read object: %w", err)
	}

	return &FileInfo{
		Data:        data,
		ContentType: stat.ContentType,
		Size:        stat.Size,
	}, nil
}

func (c *Client) Delete(ctx context.Context, objectName string) error {
	err := c.mc.RemoveObject(ctx, c.bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("minio remove object: %w", err)
	}
	return nil
}

func (c *Client) PublicURL(objectName string) string {
	if c.publicBaseURL != "" {
		return c.publicBaseURL + "/" + objectName
	}
	scheme := "http"
	if c.useSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, c.endpoint, c.bucket, objectName)
}
