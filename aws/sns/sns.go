package sns

import (
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws"
)

type Notifier struct {
	topic string
	snsService *sns.SNS
}

func New(awsSession client.ConfigProvider, topic string) *Notifier {
	snsService := sns.New(awsSession)
	return &Notifier{topic:topic, snsService:snsService}
}

func (notifier *Notifier) Notify(message string) error {
	snsServiceParams := &sns.PublishInput{
		Message: aws.String(message),
		TopicArn: aws.String(notifier.topic),
	}

	_, err := notifier.snsService.Publish(snsServiceParams)
	return err
}
