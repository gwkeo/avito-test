package models

import "time"

type PullRequest struct {
	Id        int64
	Name      string
	AuthorId  int64
	Status    string
	CreatedAt time.Time
}
