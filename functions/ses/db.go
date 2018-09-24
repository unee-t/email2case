package main

import (
	"fmt"

	"github.com/pkg/errors"
)

func (h handler) lookupID(email string) (UserID int, err error) {
	err = h.db.QueryRow("SELECT userid FROM profiles WHERE login_name=?", email).Scan(&UserID)
	if err != nil {
		return UserID, errors.Wrap(err, fmt.Sprintf("email %s does not exist", email))
	}
	return
}

func (h handler) lookupAPIkey(UserID int) (APIkey string, err error) {
	err = h.db.QueryRow("SELECT api_key FROM user_api_keys WHERE user_id=?", UserID).Scan(&APIkey)
	if err != nil {
		return APIkey, errors.Wrap(err, fmt.Sprintf("user %d does not have an API key", UserID))
	}
	return
}
