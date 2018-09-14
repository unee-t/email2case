package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/template"
	"github.com/apex/log"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jhillyerd/enmime"
	"github.com/unee-t/env"
)

func LambdaHandler(ctx context.Context, payload events.SNSEvent) (err error) {
	var email events.SimpleEmailService
	err = json.Unmarshal([]byte(payload.Records[0].SNS.Message), &email)
	if err != nil {
		log.WithError(err).Fatal("bad JSON")
		return
	}
	log.Infof("%+v", email)

	h, err := New()
	if err != nil {
		log.WithError(err).Fatal("error setting configuration")
	}
	defer h.db.Close()

	parts, err := h.inbox(email)
	if err != nil {
		log.WithError(err).Fatal("could not inbox")
	}

	parts["validReply"] = fmt.Sprintf("%t", h.validReply(email.Mail.Destination[0]))

	stssvc := sts.New(h.Env.Cfg)
	input := &sts.GetCallerIdentityInput{}

	req := stssvc.GetCallerIdentityRequest(input)
	result, err := req.Send()
	if err != nil {
		log.WithError(err).Fatal("unable to get STS")
		return
	}

	log.Infof("Parts: %+v", parts)
	snssvc := sns.New(h.Env.Cfg)
	snsreq := snssvc.PublishRequest(&sns.PublishInput{
		Message:  aws.String(fmt.Sprintf("%s", summarise(email, parts))),
		TopicArn: aws.String(fmt.Sprintf("arn:aws:sns:us-west-2:%s:incomingreply", aws.StringValue(result.Account))),
	})

	_, err = snsreq.Send()

	return

}

func main() {
	lambda.Start(LambdaHandler)
}

type handler struct {
	db  *sql.DB
	Env env.Env // Env.cfg for the AWS cfg
}

func New() (h handler, err error) {

	cfg, err := external.LoadDefaultAWSConfig(external.WithSharedConfigProfile("uneet-dev"))
	if err != nil {
		log.WithError(err).Fatal("setting up credentials")
		return
	}
	cfg.Region = endpoints.ApSoutheast1RegionID
	e, err := env.New(cfg)
	if err != nil {
		log.WithError(err).Warn("error getting unee-t env")
	}

	var mysqlhost string
	val, ok := os.LookupEnv("MYSQL_HOST")
	if ok {
		log.Infof("MYSQL_HOST overridden by local env: %s", val)
		mysqlhost = val
	} else {
		mysqlhost = e.Udomain("auroradb")
	}

	h = handler{Env: e}

	h.db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/bugzilla?multiStatements=true&sql_mode=TRADITIONAL",
		e.GetSecret("MYSQL_USER"),
		e.GetSecret("MYSQL_PASSWORD"),
		mysqlhost))
	if err != nil {
		log.WithError(err).Fatal("error opening database")
	}

	return

}

func (h handler) inbox(email events.SimpleEmailService) (parts map[string]string, err error) {

	svc := s3.New(h.Env.Cfg)

	rawMessage := fmt.Sprintf("incoming/%s", email.Mail.MessageID)

	input := &s3.GetObjectInput{
		Bucket: aws.String("dev-email-unee-t"), // Goto env
		Key:    aws.String(rawMessage),
	}

	//fmt.Println(input)

	req := svc.GetObjectRequest(input)
	original, err := req.Send()
	if err != nil {
		log.WithError(err).Fatal("could not fetch")
		return
	}
	// fmt.Println(original.Body)

	envelope, err := enmime.ReadEnvelope(original.Body)

	aclputparams := &s3.PutObjectAclInput{
		Bucket: aws.String("dev-email-unee-t"),
		Key:    aws.String(rawMessage),
		ACL:    s3.ObjectCannedACLPublicRead,
	}

	s3aclreq := svc.PutObjectAclRequest(aclputparams)
	_, err = s3aclreq.Send()
	if err != nil {
		log.WithError(err).Fatal("making rawMessage readable")
		return
	}

	textPart := time.Now().Format("2006-01-02") + "/" + email.Mail.MessageID + "/text"

	putparams := &s3.PutObjectInput{
		Bucket:      aws.String("dev-email-unee-t"),
		Body:        bytes.NewReader([]byte(envelope.Text)),
		Key:         aws.String(textPart),
		ContentType: aws.String("text/plain; charset=UTF-8"),
		ACL:         s3.ObjectCannedACLPublicRead,
	}

	s3req := svc.PutObjectRequest(putparams)
	_, err = s3req.Send()
	if err != nil {
		log.WithError(err).Fatal("putting text part")
		return
	}

	htmlPart := time.Now().Format("2006-01-02") + "/" + email.Mail.MessageID + "/html"

	putparams = &s3.PutObjectInput{
		Bucket:      aws.String("dev-email-unee-t"),
		Body:        bytes.NewReader([]byte(envelope.HTML)),
		Key:         aws.String(htmlPart),
		ContentType: aws.String("text/html; charset=UTF-8"),
		ACL:         s3.ObjectCannedACLPublicRead,
	}

	s3req = svc.PutObjectRequest(putparams)
	_, err = s3req.Send()
	if err != nil {
		log.WithError(err).Fatal("putting html part")
		return
	}

	log.Infof("%+v", envelope)

	parts = make(map[string]string)
	parts["orig"] = fmt.Sprintf("https://s3-ap-southeast-1.amazonaws.com/dev-email-unee-t/incoming/%s", email.Mail.MessageID)
	parts["text"] = fmt.Sprintf("https://s3-ap-southeast-1.amazonaws.com/dev-email-unee-t/%s", textPart)
	parts["html"] = fmt.Sprintf("https://s3-ap-southeast-1.amazonaws.com/dev-email-unee-t/%s", htmlPart)

	return
}

func summarise(email events.SimpleEmailService, parts map[string]string) string {

	// https://github.com/aws/aws-lambda-go/blob/master/events/ses.go
	tmpl, err := template.New("").Parse(`
To: {{.Mail.CommonHeaders.To}}
From: {{.Mail.CommonHeaders.From}}
Date: {{.Mail.CommonHeaders.Date}}
Subject: {{.Mail.CommonHeaders.Subject}}
MessageID: {{.Mail.MessageID}}
	`)
	if err != nil {
		panic(err)
	}
	var output bytes.Buffer
	err = tmpl.Execute(&output, email)
	if err != nil {
		panic(err)
	}

	tmpl, err = template.New("").Parse(`{{range $k,$v := .}}
{{ $k }} {{$v}}{{end}}`)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(&output, parts)
	if err != nil {
		panic(err)
	}

	return output.String()
}

func (h handler) validReply(toAddress string) bool {
	log.Infof("Checking reply address is valid: %s", toAddress)

	e, err := mail.ParseAddress(toAddress)
	if err != nil {
		return false
	}

	if !strings.HasPrefix(e.Address, "reply+") {
		return false
	}

	parts := strings.Split(e.Address, "-")
	// fmt.Println("parts", parts, len(parts))

	if len(parts) < 2 {
		return false
	}

	// fmt.Println("parts", parts)
	replyParts := strings.Split(parts[0], "+")

	if len(replyParts) != 2 {
		return false
	}

	endParts := strings.Split(parts[1], "@")

	if len(endParts) != 2 {
		return false
	}

	accessToken := h.Env.GetSecret("API_ACCESS_TOKEN")
	log.Infof("API_ACCESS_TOKEN", accessToken)
	return checkMAC(replyParts[1], endParts[0], accessToken)

}

func checkMAC(message, messageMAC, key string) bool {
	// fmt.Println(message, messageMAC, key)
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	expectedMAC := mac.Sum(nil)

	computedMAC, _ := hex.DecodeString(messageMAC)
	return hmac.Equal(computedMAC, expectedMAC)
}
