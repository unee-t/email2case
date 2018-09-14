package main

func (h handler) lookupID(email string) (UserID int, err error) {
	err = h.db.QueryRow("SELECT userid FROM profiles WHERE login_name=?", email).Scan(&UserID)
	return
}

func (h handler) lookupAPIkey(UserID int) (APIkey string, err error) {
	err = h.db.QueryRow("SELECT api_key FROM user_api_keys WHERE user_id=?", UserID).Scan(&APIkey)
	return
}
