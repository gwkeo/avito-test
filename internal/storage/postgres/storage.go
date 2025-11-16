package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
)

/*

users			team
-----			-----
user_id			id
username		name
team_id
is_active

pull_request	reviewers
-------------	----------
id				pr_id
name			reviewers_id
author_id
status
created_at
merged_at

*/

type UserRepository interface {

}

type Storage struct {
	db *pgx.Conn
}

func New(ctx context.Context, connectionString string) (*Storage, error) {
	db, err := pgx.Connect(ctx, connectionString)
	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}
