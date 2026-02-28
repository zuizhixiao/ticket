package fileFactory

import (
	"errors"
	"ticket/package/fileFactory/cos"
	"ticket/package/fileFactory/mos"
)

type FileFactory interface {
	PutFileWithBody(saveName string, bufferBytes []byte) (string, error)
}

func SetUp(accessKeyId, accessKeySecret, endpoint, bucket, sType string) (FileFactory, error) {
	if sType == "cos" {
		return cos.SetUp(accessKeyId, accessKeySecret, endpoint, bucket), nil
	} else if sType == "mos" {
		return mos.SetUp(accessKeyId, accessKeySecret, endpoint, bucket), nil
	}
	return nil, errors.New("暂时不支持" + sType)
}
