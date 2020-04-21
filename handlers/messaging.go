package handlers

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/sns"
	"log"
	"os"
)

const (
	CharSet = "UTF-8"
)

func createSession() (*session.Session, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
	})
	if err != nil {
		return nil, err
	}

	return sess, err
}

func (t *MethodInterface) SendMail(recipient string, subject string, textBody string) error {
	sess, err := createSession()
	if err != nil {
		return err
	}

	svc := ses.New(sess)

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				/*Html: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(HtmlBody),
				},*/
				Text: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(textBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(CharSet),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(os.Getenv("AWS_SES_SENDER")),
		// Uncomment to use a configuration set
		//ConfigurationSetName: aws.String(ConfigurationSet),
	}

	// Attempt to send the email.
	result, err := svc.SendEmail(input)
	log.Println(result)

	return nil
}

func (t *MethodInterface) SendSms(topicArn string, endpoint string, message string) error {
	sess, err := createSession()
	if err != nil {
		return err
	}

	svc := sns.New(sess)

	result, err := svc.Subscribe(&sns.SubscribeInput{
		Endpoint:              &endpoint,
		Protocol:              aws.String("sms"),
		ReturnSubscriptionArn: aws.Bool(true), // Return the ARN, even if user has yet to confirm
		TopicArn:              &topicArn,
	})
	if err != nil {
		return err
	}

	log.Println(*result.SubscriptionArn)

	input := &sns.PublishInput{
		Message:     aws.String(message),
		PhoneNumber: aws.String(endpoint),
	}

	res, err := svc.Publish(input)
	if err != nil {
		return err
	}

	log.Println(res)

	return nil
}
