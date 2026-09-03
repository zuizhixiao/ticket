package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type mos struct {
	endpoint string
	bucket   string
	client   *minio.Client
}

// newMos MinIO / S3 兼容(如云厂商 MOS)。
func newMos(endpoint, bucket, ak, sk string) (Store, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(ak, sk, ""),
		Secure: true,
	})
	if err != nil {
		return nil, err
	}
	return &mos{endpoint: endpoint, bucket: bucket, client: client}, nil
}

func (m *mos) Put(ctx context.Context, object string, r io.Reader, size int64, contentType string) (string, error) {
	_, err := m.client.PutObject(ctx, m.bucket, object, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	return m.PublicURL(object), nil
}

func (m *mos) Delete(ctx context.Context, object string) error {
	return m.client.RemoveObject(ctx, m.bucket, object, minio.RemoveObjectOptions{})
}

func (m *mos) PublicURL(object string) string {
	return fmt.Sprintf("https://%s/%s/%s", m.endpoint, m.bucket, object)
}
