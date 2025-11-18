package postgres

import (
	"context"
	"errors"

	"github.com/gwkeo/avito-test/internal/models"
	"github.com/gwkeo/avito-test/internal/storage"
	"github.com/jackc/pgx/v5"
)

func (s *Storage) CreatePullRequest(ctx context.Context, pullRequestID, pullRequestName, authorID string, assignedReviewers []string) (*models.PullRequest, error) {

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var pr models.PullRequest
	if err := tx.QueryRow(ctx,
		"INSERT INTO pull_requests (id, name, author_id, status, created_at) VALUES ($1, $2, $3, $4, NOW()) RETURNING id, name, author_id, status, created_at",
		pullRequestID,
		pullRequestName,
		authorID,
		storage.PullRequestStatusOpen,
	).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
	); err != nil {
		return nil, err
	}

	if len(assignedReviewers) == 0 {
		tx.Commit(ctx)
		return &pr, nil
	}

	for _, reviewer := range assignedReviewers {
		if _, err := tx.Exec(ctx, "INSERT INTO reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)", pullRequestID, reviewer); err != nil {
			return nil, err
		}
	}

	pr.AssignedReviewers = assignedReviewers

	tx.Commit(ctx)
	return &pr, nil
}
func (s *Storage) MergePullRequest(ctx context.Context, pullRequestID string) (*models.PullRequest, error) {
	status := ""
	err := s.db.QueryRow(ctx, "SELECT status FROM pull_requests WHERE id = $1", pullRequestID).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}

	if status == storage.PullRequestStatusMerged {
		return nil, storage.ErrPRMerged
	}

	var pr models.PullRequest
	if err := s.db.QueryRow(ctx,
		"UPDATE pull_requests SET status = $1, merged_at = NOW() WHERE id = $2 RETURNING id, name, author_id, status, created_at, merged_at",
		storage.PullRequestStatusMerged,
		pullRequestID,
	).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt,
	); err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, "SELECT reviewer_id FROM reviewers WHERE pull_request_id = $1", pullRequestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, id)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	pr.AssignedReviewers = reviewers
	return &pr, nil
}

func (s *Storage) ReassignPullRequest(ctx context.Context, pullRequestID, oldReviewerID, newReviewerID string) (*models.PullRequest, string, error) {
	_, err := s.db.Exec(ctx, "SELECT id FROM pull_requests WHERE id = $1", pullRequestID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", storage.ErrNotFound
		}
		return nil, "", err
	}

	_, err = s.db.Exec(
		ctx,
		"UPDATE reviewers SET reviewer_id = $1 WHERE reviewer_id = $2 AND pull_request_id = $3",
		newReviewerID,
		oldReviewerID,
		pullRequestID,
	)
	if err != nil {
		return nil, "", err
	}

	var pr models.PullRequest
	err = s.db.QueryRow(ctx, `
        SELECT 
            pr.id,
            pr.name,
            pr.author_id,
            pr.status,
            pr.created_at,
            pr.merged_at,
            COALESCE(
                ARRAY_AGG(r.reviewer_id) FILTER (WHERE r.reviewer_id IS NOT NULL),
                ARRAY[]::TEXT[]
            ) AS assigned_reviewers
        FROM pull_requests pr
        LEFT JOIN reviewers r ON pr.id = r.pull_request_id
        WHERE pr.id = $1
        GROUP BY pr.id, pr.name, pr.author_id, pr.status, pr.created_at, pr.merged_at
    `, pullRequestID).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt,
		&pr.AssignedReviewers,
	)

	if err != nil {
		return nil, "", err
	}

	return &pr, newReviewerID, nil
}
