package models

type User struct {
	ID       int64
	Name     string
	Email    string
	PassHash []byte
}

type UsersInfo struct {
	Names  []string
	Emails []string
}
