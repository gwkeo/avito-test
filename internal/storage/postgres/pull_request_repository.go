package postgres

import (
	"context"
	"errors"

	"github.com/gwkeo/avito-test/internal/models"
	"github.com/gwkeo/avito-test/internal/storage"
	"github.com/jackc/pgx/v5"
)

func (s *Storage) PullRequest(ctx context.Context, pullRequestID string) (*models.PullRequest, error) {
	var pr models.PullRequest
	if err := s.db.QueryRow(ctx, "SELECT * FROM pull_requests WHERE id = ?", pullRequestID).Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}

	rows, err := s.db.Query(ctx, "SELECT * FROM reviewers WHERE pull_request_id = ?", pullRequestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		id := ""
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, id)
	}

	pr.AssignedReviewers = reviewers

	return &pr, nil
}

func (s *Storage) CreatePullRequest(ctx context.Context, pullRequestID, pullRequestName, authorID string, assignedReviewers []string) (*models.PullRequest, error) {

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var pr models.PullRequest
	if err := tx.QueryRow(ctx,
		"INSERT INTO pull_requests VALUES id = ?, name = ?, author_id = ?, status = ?",
		pullRequestID,
		pullRequestName,
		authorID,
		storage.PullRequestStateOpen,
	).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
	); err != nil {
		return nil, err
	}

	for _, reviewer := range assignedReviewers {
		if _, err := tx.Query(ctx, "INSERT INTO reviewers VALUES pull_request_id = $1, reviewer_id = $2", pullRequestID, reviewer); err != nil {
			return nil, err
		}
	}

	pr.AssignedReviewers = assignedReviewers

	tx.Commit(ctx)

	return &pr, nil
}

func (s *Storage) MergePullRequest(ctx context.Context, pullRequestID string) (*models.PullRequest, error) {
	if _, err := s.db.Exec(ctx,
		"UPDATE pull_requests SET state = $1, merged_at = NOW() WHERE id = $2",
		storage.PullRequestStateMerged,
		pullRequestID,
	); err != nil {
		return nil, err
	}

	pr, err := s.PullRequest(ctx, pullRequestID)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *Storage) ReassignPullRequest(ctx context.Context, pullRequestID, oldReviewerID, newReviewerID string) (*models.PullRequest, string, error) {

	_, err := s.db.Exec(
		ctx,
		"UPDATE reviewers SET reviewer_id = $1 WHERE reviewer_id = $2 AND pull_request_id = $3",
		newReviewerID,
		oldReviewerID,
		pullRequestID,
	)
	if err != nil {
		return nil, "", err
	}

	pr, err := s.PullRequest(ctx, pullRequestID)
	if err != nil {
		return nil, "", err
	}

	return pr, newReviewerID, nil
}
