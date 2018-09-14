package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/apex/log"
	"github.com/aws/aws-lambda-go/events"
	"github.com/davecgh/go-spew/spew"
	_ "github.com/go-sql-driver/mysql"
)

func TestIntegration(t *testing.T) {

	content, err := ioutil.ReadFile("ses.json")
	if err != nil {
		log.WithError(err).Fatal("could not open ses.json")
	}

	h, err := New()
	if err != nil {
		log.WithError(err).Fatal("error setting configuration")
	}
	defer h.db.Close()

	var email events.SimpleEmailService
	err = json.Unmarshal([]byte(content), &email)
	if err != nil {
		log.WithError(err).Fatal("could not unmarshall json")
	}
	//fmt.Printf("%+v\n", email)
	spew.Dump(email)

	parts, err := h.inbox(email)
	if err != nil {
		log.WithError(err).Fatal("could not inbox")
	}

	fmt.Println(summarise(email, parts))

}
