package s3

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/aws/client"
	"os"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/aws"
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
	defer file.Close()
	_, err := uploader.s3Uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(uploader.bucketName),
		Key: aws.String(file.Name()),
		Body: file,
	})
	return err
}
