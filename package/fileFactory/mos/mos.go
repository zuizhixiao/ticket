package mos

import (
	"bytes"
	"context"
	"net/http"
	"path"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Mos struct {
	Bucket          string
	Endpoint        string
	AccessKeyId     string
	AccessKeySecret string
	Client          *minio.Client
}

func SetUp(accessKeyId, accessKeySecret, endpoint, bucket string) *Mos {
	minioClient, _ := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyId, accessKeySecret, ""),
		Secure: true,
	})
	return &Mos{
		Bucket:          bucket,
		Endpoint:        endpoint,
		AccessKeyId:     accessKeyId,
		AccessKeySecret: accessKeySecret,
		Client:          minioClient,
	}
}

// PutFileWithBody 上传文件到指定文件夹
func (m *Mos) PutFileWithBody(saveName string, bufferBytes []byte) (string, error) {
	ctx := context.Background()
	contentType := http.DetectContentType(bufferBytes)
	ext := path.Ext(saveName)
	if ext == ".css" {
		contentType = "text/css"
	}
	if ext == ".js" {
		contentType = "application/javascript"
	}
	_, err := m.Client.PutObject(ctx, m.Bucket, saveName, bytes.NewReader(bufferBytes),
		int64(len(bufferBytes)), minio.PutObjectOptions{
			ContentType: contentType,
		})
	return "https://" + m.Endpoint + "/" + m.Bucket + "/" + saveName, err
}
