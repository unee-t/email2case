package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/apex/log"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func LambdaHandler(ctx context.Context, payload events.SNSEvent) (err error) {
	var email events.SimpleEmailService
	err = json.Unmarshal([]byte(payload.Records[0].SNS.Message), &email)
	if err != nil {
		log.WithError(err).Fatal("bad JSON")
		return
	}
	log.Infof("%+v", email)

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return
	}

	stssvc := sts.New(cfg)
	input := &sts.GetCallerIdentityInput{}

	req := stssvc.GetCallerIdentityRequest(input)
	result, err := req.Send()
	if err != nil {
		return
	}

	snssvc := sns.New(cfg)
	snsreq := snssvc.PublishRequest(&sns.PublishInput{
		Message:  aws.String(fmt.Sprintf("%s", email.Mail.CommonHeaders.Subject)),
		TopicArn: aws.String(fmt.Sprintf("arn:aws:sns:us-west-2:%s:incomingreply", aws.StringValue(result.Account))),
	})

	_, err = snsreq.Send()

	return

}

func main() {
	lambda.Start(LambdaHandler)
}
