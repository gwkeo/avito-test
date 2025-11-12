package models

type User struct {
	Id       string
	Username string
	TeamId   int64
	IsActive bool
}
