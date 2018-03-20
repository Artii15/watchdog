package s3

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/aws/client"
	"os"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Uploader struct {
	s3Uploader *s3manager.Uploader
	bucketName string
}

func New(awsSession client.ConfigProvider, bucketName string) *Uploader {
	s3Service := s3.New(awsSession)
	uploader := s3manager.NewUploaderWithClient(s3Service)

	return &Uploader{s3Uploader: uploader, bucketName: bucketName}
}

func (uploader *Uploader) Upload(file *os.File) error {
	objectKey := file.Name()
	_, err := uploader.s3Uploader.Upload(&s3manager.UploadInput{
		Bucket: &uploader.bucketName,
		Key: &objectKey,
		Body: file,
	})
	return err
}
