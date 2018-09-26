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
	"io/ioutil"
	"net/http"
	"net/mail"
	"os"
	"regexp"
	"strconv"
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
	_ "github.com/go-sql-driver/mysql"
	"github.com/jhillyerd/enmime"
	"github.com/pkg/errors"
	"github.com/unee-t/env"
)

func LambdaHandler(ctx context.Context, payload events.SNSEvent) (err error) {

	j, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "unable to marshal")
	}

	log.Infof("JSON payload %s", string(j))

	var email events.SimpleEmailService

	err = json.Unmarshal([]byte(payload.Records[0].SNS.Message), &email)
	if err != nil {
		return errors.Wrap(err, "bad JSON")
	}

	h, err := New()
	if err != nil {
		return errors.Wrap(err, "error setting configuration")
	}
	defer h.db.Close()

	parts, err := h.inbox(email)
	if err != nil {
		return errors.Wrap(err, "could not inbox")
	}

	log.Infof("Parts: %+v, TopicArn: %s", parts, h.Env.SNS("incomingreply", "us-west-2"))

	cfg, err := external.LoadDefaultAWSConfig(external.WithSharedConfigProfile("uneet-dev"))
	if err != nil {
		return errors.Wrap(err, "setting up credentials")
	}
	cfg.Region = endpoints.UsWest2RegionID

	snssvc := sns.New(cfg)
	snsreq := snssvc.PublishRequest(&sns.PublishInput{
		Message:  aws.String(fmt.Sprintf("%s", summarise(email, parts))),
		TopicArn: aws.String(h.Env.SNS("incomingreply", endpoints.UsWest2RegionID)),
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
		return
	}
	cfg.Region = endpoints.ApSoutheast1RegionID
	e, err := env.New(cfg)
	if err != nil {
		return
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
		return
	}

	return

}

func (h handler) inbox(email events.SimpleEmailService) (parts map[string]string, err error) {
	parts = make(map[string]string)
	bugid, userid, err := h.validReply(email.Mail.Destination[0])

	if err != nil {
		return parts, err
	}

	svc := s3.New(h.Env.Cfg)

	rawMessage := fmt.Sprintf("incoming/%s", email.Mail.MessageID)

	input := &s3.GetObjectInput{
		Bucket: aws.String(h.Env.Bucket("email")), // Goto env
		Key:    aws.String(rawMessage),
	}

	//fmt.Println(input)

	req := svc.GetObjectRequest(input)
	original, err := req.Send()
	if err != nil {
		return
	}
	// fmt.Println(original.Body)

	envelope, err := enmime.ReadEnvelope(original.Body)

	aclputparams := &s3.PutObjectAclInput{
		Bucket: aws.String(h.Env.Bucket("email")),
		Key:    aws.String(rawMessage),
		ACL:    s3.ObjectCannedACLPublicRead,
	}

	s3aclreq := svc.PutObjectAclRequest(aclputparams)
	_, err = s3aclreq.Send()
	if err != nil {
		return
	}

	textPartKey := time.Now().Format("2006-01-02") + "/" + email.Mail.MessageID + "/text"

	err = h.comment(userid, bugid, envelope.Text)
	if err != nil {
		return
	}

	putparams := &s3.PutObjectInput{
		Bucket:      aws.String(h.Env.Bucket("email")),
		Body:        bytes.NewReader([]byte(envelope.Text)),
		Key:         aws.String(textPartKey),
		ContentType: aws.String("text/plain; charset=UTF-8"),
		ACL:         s3.ObjectCannedACLPublicRead,
	}

	s3req := svc.PutObjectRequest(putparams)
	_, err = s3req.Send()
	if err != nil {
		return
	}

	htmlPart := time.Now().Format("2006-01-02") + "/" + email.Mail.MessageID + "/html"

	putparams = &s3.PutObjectInput{
		Bucket:      aws.String(h.Env.Bucket("email")),
		Body:        bytes.NewReader([]byte(envelope.HTML)),
		Key:         aws.String(htmlPart),
		ContentType: aws.String("text/html; charset=UTF-8"),
		ACL:         s3.ObjectCannedACLPublicRead,
	}

	s3req = svc.PutObjectRequest(putparams)
	_, err = s3req.Send()
	if err != nil {
		return
	}

	log.Infof("%+v", envelope)

	parts["orig"] = fmt.Sprintf("https://s3-ap-southeast-1.amazonaws.com/%s/incoming/%s",
		h.Env.Bucket("email"), email.Mail.MessageID)
	parts["text"] = fmt.Sprintf("https://s3-ap-southeast-1.amazonaws.com/%s/%s",
		h.Env.Bucket("email"), textPartKey)
	parts["html"] = fmt.Sprintf("https://s3-ap-southeast-1.amazonaws.com/%s/%s",
		h.Env.Bucket("email"), htmlPart)
	parts["bugURL"] = fmt.Sprintf("https://%s/case/%d", h.Env.Udomain("case"), bugid)
	parts["bugid"] = fmt.Sprintf("%d", bugid)
	parts["userid"] = fmt.Sprintf("%d", userid)

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

func cleanReply(comment string) (cleanedComment string, err error) {

	reg := regexp.MustCompile(`On .* wrote:`)
	parts := reg.Split(comment, 2)

	cleanedComment = strings.TrimSpace(parts[0])
	if cleanedComment == "" {
		return cleanedComment, fmt.Errorf("Empty reply")
	}

	return
}

func (h handler) comment(userid, bugid int, comment string) (err error) {
	log.Infof("userid: %d, BugID: %d, Comment: %s", userid, bugid, comment)

	comment, err = cleanReply(comment)
	if err != nil {
		return err
	}

	if bugid == 0 {
		return fmt.Errorf("Missing bug number")
	}

	url := fmt.Sprintf("https://%s/rest/bug/%d/comment", h.Env.Udomain("dashboard"), bugid)

	apikey, err := h.lookupAPIkey(userid)
	if err != nil {
		return err
	}

	payload, _ := json.Marshal(struct {
		APIkey  string `json:"api_key"`
		Comment string `json:"comment"`
	}{
		APIkey:  apikey,
		Comment: comment,
	})

	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	log.Infof("bz response: %v\n%s", res, string(body))

	return
}

func (h handler) validReply(toAddress string) (caseID, bzUserID int, err error) {
	log.Infof("Checking reply address is valid: %s", toAddress)

	e, err := mail.ParseAddress(toAddress)
	if err != nil {
		return
	}

	if !strings.HasPrefix(e.Address, "reply+") {
		return caseID, bzUserID, fmt.Errorf("missing reply+ prefix")
	}

	parts := strings.Split(e.Address, "-")
	fmt.Println("parts", parts, len(parts))

	if len(parts) < 3 {
		return caseID, bzUserID, fmt.Errorf("not in caseid-userid-signature structure")
	}

	replyParts := strings.Split(parts[0], "+")

	if len(replyParts) != 2 {
		return caseID, bzUserID, fmt.Errorf("missing caseid")
	}

	endParts := strings.Split(parts[2], "@")
	if len(endParts) != 2 {
		return caseID, bzUserID, fmt.Errorf("missing signature")
	}
	sig := endParts[0]

	log.Infof("parts: %+v\nreplyParts: %+v\nendParts: %+v", parts, replyParts, endParts)

	caseID, err = strconv.Atoi(replyParts[1])
	bzUserID, err = strconv.Atoi(parts[1])
	if err != nil {
		return
	}

	accessToken := h.Env.GetSecret("API_ACCESS_TOKEN")

	msg := fmt.Sprintf("%d%d", caseID, bzUserID)

	log.WithFields(log.Fields{
		"caseID":      caseID,
		"bzUserID":    bzUserID,
		"msg":         msg,
		"sig":         sig,
		"accessToken": accessToken,
	}).Info("checking mac")

	if !checkMAC(msg, sig, accessToken) {
		return caseID, bzUserID, fmt.Errorf("signature failed")
	}
	return

}

func checkMAC(message, messageMAC, key string) bool {
	fmt.Println(message, messageMAC, key)
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	expectedMAC := mac.Sum(nil)

	computedMAC, _ := hex.DecodeString(messageMAC)
	return hmac.Equal(computedMAC, expectedMAC)
}
