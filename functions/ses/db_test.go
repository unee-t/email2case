package main

import (
	"os"
	"testing"
)

var h handler

func TestMain(m *testing.M) {
	h, _ = New()
	defer h.db.Close()
	os.Exit(m.Run())
}

func Test_handler_lookupIDwithEmail(t *testing.T) {
	type args struct {
		email string
	}
	tests := []struct {
		name       string
		args       args
		wantUserID int
		wantErr    bool
	}{
		{
			name: "",
			args: args{
				email: "hendry@iki.fi",
			},
			wantUserID: 86,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUserID, err := h.lookupID(tt.args.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("handler.lookupIDwithEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUserID != tt.wantUserID {
				t.Errorf("handler.lookupIDwithEmail() = %v, want %v", gotUserID, tt.wantUserID)
			}
		})
	}
}

func Test_handler_lookupAPIkey(t *testing.T) {
	type args struct {
		UserID int
	}
	tests := []struct {
		name       string
		args       args
		wantAPIkey string
		wantErr    bool
	}{
		{
			name: "hendry",
			args: args{
				UserID: 86,
			},
			wantAPIkey: "onm6xtyyytjxW438",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAPIkey, err := h.lookupAPIkey(tt.args.UserID)
			if (err != nil) != tt.wantErr {
				t.Errorf("handler.lookupAPIkey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotAPIkey != tt.wantAPIkey {
				t.Errorf("handler.lookupAPIkey() = %v, want %v", gotAPIkey, tt.wantAPIkey)
			}
		})
	}
}
