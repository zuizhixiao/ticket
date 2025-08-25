package cos

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

type Cos struct {
	Bucket   string
	Endpoint string
	Client   *cos.Client
	// 添加文件夹路径配置
	PosterFolder  string // 海报图片存储文件夹
	ProductFolder string // 成品图片存储文件夹
}

func SetUp(accessKeyId, accessKeySecret, endpoint, bucket string) *Cos {
	u, _ := url.Parse("https://" + bucket + "." + endpoint)
	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  os.Getenv(accessKeyId),
			SecretKey: os.Getenv(accessKeySecret),
		},
	})

	return &Cos{
		Bucket:        bucket,
		Endpoint:      endpoint,
		Client:        client,
		PosterFolder:  "images/poster",  // 海报图片存储文件夹
		ProductFolder: "images/product", // 成品图片存储文件夹
	}
}

// PutFileWithBody 上传文件到指定文件夹
func (m *Cos) PutFileWithBody(saveName string, bufferBytes []byte) error {
	ctx := context.Background()
	_, err := m.Client.Object.Put(ctx, saveName, bytes.NewReader(bufferBytes), nil)
	return err
}

// PutPosterFile 上传海报图片
func (m *Cos) PutPosterFile(filename string, bufferBytes []byte) error {
	// 海报图片存储路径：images/poster/日期/文件名
	dateFormat := getCurrentDate()
	savePath := filepath.Join(m.PosterFolder, dateFormat, filename)
	return m.PutFileWithBody(savePath, bufferBytes)
}

// PutProductFile 上传成品图片
func (m *Cos) PutProductFile(filename string, bufferBytes []byte) error {
	// 成品图片存储路径：images/product/日期/文件名
	dateFormat := getCurrentDate()
	savePath := filepath.Join(m.ProductFolder, dateFormat, filename)
	return m.PutFileWithBody(savePath, bufferBytes)
}

// getCurrentDate 获取当前日期字符串
func getCurrentDate() string {
	return time.Now().Format("20060102")
}
