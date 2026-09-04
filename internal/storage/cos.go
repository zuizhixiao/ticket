package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/tencentyun/cos-go-sdk-v5"
)

type cosStore struct {
	bucket   string
	endpoint string
	client   *cos.Client
}

// newCos 腾讯云 COS。endpoint 形如 cos.ap-shanghai.myqcloud.com。
func newCos(endpoint, bucket, secretID, secretKey string) (Store, error) {
	bucketURL, err := url.Parse("https://" + bucket + "." + endpoint)
	if err != nil {
		return nil, err
	}
	client := cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretID,
			SecretKey: secretKey,
		},
	})
	return &cosStore{bucket: bucket, endpoint: endpoint, client: client}, nil
}

func (c *cosStore) Put(ctx context.Context, object string, r io.Reader, size int64, contentType string) (string, error) {
	_, err := c.client.Object.Put(ctx, object, r, &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{ContentType: contentType},
	})
	if err != nil {
		return "", err
	}
	return c.PublicURL(object), nil
}

func (c *cosStore) Delete(ctx context.Context, object string) error {
	_, err := c.client.Object.Delete(ctx, object)
	return err
}

// DeleteByURL 从公网 URL 反解对象 key 后删除。
func (c *cosStore) DeleteByURL(ctx context.Context, rawURL string) error {
	prefix := fmt.Sprintf("https://%s.%s/", c.bucket, c.endpoint)
	if !strings.HasPrefix(rawURL, prefix) {
		return fmt.Errorf("URL 不属于当前存储桶:%s", rawURL)
	}
	return c.Delete(ctx, strings.TrimPrefix(rawURL, prefix))
}

func (c *cosStore) PublicURL(object string) string {
	return fmt.Sprintf("https://%s.%s/%s", c.bucket, c.endpoint, object)
}
