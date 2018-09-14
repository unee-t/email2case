package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/apex/log"
	"github.com/aws/aws-lambda-go/events"
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

	parts, err := h.inbox(email)
	if err != nil {
		log.WithError(err).Fatal("could not inbox")
	}

	fmt.Println(summarise(email, parts))

}

func Test_handler_comment(t *testing.T) {
	hostname, _ := os.Hostname()
	type args struct {
		from    string
		bugID   string
		comment string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Comment from Go",
			args: args{
				from:    "hendry@iki.fi",
				bugID:   "61825",
				comment: "From go test " + hostname,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := h.comment(tt.args.from, tt.args.bugID, tt.args.comment); (err != nil) != tt.wantErr {
				t.Errorf("handler.comment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
