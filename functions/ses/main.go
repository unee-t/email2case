package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var invokeCount = 0
var myObjects []*s3.Object

func init() {
	svc := s3.New(session.New(), aws.NewConfig().WithRegion("us-east-1"))
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String("dev-email-unee-t"),
	}
	result, err := svc.ListObjectsV2(input)
	if err != nil {
		panic(err)
	}
	myObjects = result.Contents
}

func LambdaHandler(ctx context.Context, payload json.RawMessage) {
	fmt.Println("here")
	log.Println(payload)

	json, err := json.MarshalIndent(payload, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(json))
}

func main() {
	lambda.Start(LambdaHandler)
}
