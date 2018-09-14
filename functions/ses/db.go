package main

func (h handler) lookupIDwithEmail(email string) (UserID int, err error) {
	err = h.db.QueryRow("SELECT userid FROM profiles WHERE login_name=?", email).Scan(&UserID)
	return
}
