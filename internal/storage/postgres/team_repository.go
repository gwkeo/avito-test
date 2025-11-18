package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/gwkeo/avito-test/internal/models"
	"github.com/gwkeo/avito-test/internal/storage"
	"github.com/jackc/pgx/v5"
)

func (s *Storage) AddTeam(ctx context.Context, team *models.Team) (*models.Team, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	defer tx.Rollback(ctx)

	if _, err = tx.Exec(ctx, "INSERT INTO teams (name) VALUES ($1)", team.TeamName); err != nil {
		return nil, fmt.Errorf("insert team: %w", err)
	}

	for _, member := range team.Members {
		if _, err = tx.Exec(ctx,
			"INSERT INTO users (id, username, team_name, is_active) VALUES ($1, $2, $3, $4) RETURNING id, username, team_name, is_active",
			member.UserID, member.Username, team.TeamName, member.IsActive,
		); err != nil {
			return nil, fmt.Errorf("insert user %s: %w", member.Username, err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return team, nil
}

func (s *Storage) Team(ctx context.Context, teamName string) (*models.Team, error) {
	var team models.Team
	if err := s.db.QueryRow(ctx, "SELECT name FROM teams WHERE name = $1", teamName).Scan(&team.TeamName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}

	rows, err := s.db.Query(ctx, "SELECT id, username, is_active FROM users WHERE team_name = $1", teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.TeamMember
	for rows.Next() {
		var member models.TeamMember
		if err = rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return nil, err
		}

		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	team.Members = members

	return &team, nil
}
