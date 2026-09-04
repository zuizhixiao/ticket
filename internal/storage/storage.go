// Package storage 抽象对象存储(MinIO/S3 兼容 与 腾讯云 COS)。
package storage

import (
	"context"
	"fmt"
	"io"
)

// Store 对象存储接口。object 形如 "product/20260220/1739...png"(不含前导斜杠)。
type Store interface {
	// Put 上传并返回公网可访问 URL。
	Put(ctx context.Context, object string, r io.Reader, size int64, contentType string) (string, error)
	// Delete 按对象 key 删除(对象不存在返回 nil)。
	Delete(ctx context.Context, object string) error
	// PublicURL 依据 object 拼出公网 URL。
	PublicURL(object string) string
	// DeleteByURL 依据公网 URL 反解对象 key 后删除。
	DeleteByURL(ctx context.Context, rawURL string) error
}

// Default 全局存储客户端,由 main 在启动时注入。
var Default Store

// New 依据 type(mos/cos)创建存储客户端。
func New(typ, accessKeyID, accessKeySecret, endpoint, bucket string) (Store, error) {
	switch typ {
	case "mos":
		return newMos(endpoint, bucket, accessKeyID, accessKeySecret)
	case "cos":
		return newCos(endpoint, bucket, accessKeyID, accessKeySecret)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", typ)
	}
}
