package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func Test_cleanReply(t *testing.T) {

	replyText, err := ioutil.ReadFile("reply.txt")
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		comment string
	}
	tests := []struct {
		name               string
		args               args
		wantCleanedComment string
		wantErr            bool
	}{
		{
			name: "empty",
			args: args{
				comment: fmt.Sprintf(" \n \n"),
			},
			wantCleanedComment: "",
			wantErr:            true,
		},
		{
			name: "spaced",
			args: args{
				comment: " Howdy! ",
			},
			wantCleanedComment: "Howdy!",
			wantErr:            false,
		},
		{
			name: "Remove quote",
			args: args{
				comment: string(replyText),
			},
			wantCleanedComment: "Replied to...",
			wantErr:            false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCleanedComment, err := cleanReply(tt.args.comment)
			if (err != nil) != tt.wantErr {
				t.Errorf("cleanReply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotCleanedComment != tt.wantCleanedComment {
				t.Errorf("cleanReply() = %v, want %v", gotCleanedComment, tt.wantCleanedComment)
			}
		})
	}
}

func Test_checkMAC(t *testing.T) {

	h, err := New()
	if err != nil {
		log.WithError(err).Fatal("error setting configuration")
	}
	defer h.db.Close()

	type args struct {
		message    string
		messageMAC string
		key        string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "dev",
			args: args{
				message:    "61825107",
				messageMAC: "c7a1609c4a839b7e1eae86a353ffd975e96a11e5f0f9e7bb52f4cd3010d9eb35",
				key:        "O6I9svDTizOfLfdVA5ri",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkMAC(tt.args.message, tt.args.messageMAC, tt.args.key); got != tt.want {
				t.Errorf("checkMAC() = %v, want %v", got, tt.want)
			}
		})
	}
}
