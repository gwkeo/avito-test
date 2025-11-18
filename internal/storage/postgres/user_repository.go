package postgres

import (
	"context"
	"errors"

	"github.com/gwkeo/avito-test/internal/models"
	"github.com/gwkeo/avito-test/internal/storage"
	"github.com/jackc/pgx/v5"
)

func (s *Storage) SetIsUserActive(ctx context.Context, userID string, isActive bool) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(ctx, "UPDATE users SET is_active = $1 WHERE id = $2 RETURNING id, username, is_active, team_name", isActive, userID).Scan(&user.UserID, &user.Username, &user.IsActive, &user.TeamName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (s *Storage) UsersReviews(ctx context.Context, userID string) ([]models.PullRequestShort, error) {
	rows, err := s.db.Query(
		ctx,
		"SELECT pr.id AS pull_request_id, pr.name AS pull_request_name, pr.author_id, ps.status FROM pull_requests pr JOIN reviewers r ON pr.id = r.pull_request_id WHERE reviewer_id = $1",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.PullRequestShort
	for rows.Next() {
		var prs models.PullRequestShort
		if err = rows.Scan(&prs.PullRequestID, &prs.PullRequestName, &prs.AuthorID, &prs.Status); err != nil {
			return nil, err
		}
		result = append(result, prs)
	}

	return result, nil
}

func (s *Storage) User(ctx context.Context, userID string) (*models.User, error) {
	rows, err := s.db.Query(ctx, "SELECT * FROM users WHERE id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, storage.ErrNotFound
	}

	var user models.User
	if err = rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
		return nil, err
	}

	return &user, nil
}
