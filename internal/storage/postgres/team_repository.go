package postgres

import (
	"context"
	"errors"

	"github.com/gwkeo/avito-test/internal/models"
	"github.com/gwkeo/avito-test/internal/storage"
	"github.com/jackc/pgx/v5"
)

func (s *Storage) AddTeam(ctx context.Context, team *models.Team) (*models.Team, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	for _, member := range team.Members {
		_, err := tx.Query(ctx, "INSERT INTO users VALUES (?, ?, ?, ?)", member.UserID, member.Username, team.TeamName, member.IsActive)
		if err != nil {
			return nil, err
		}
	}

	_, err = tx.Query(ctx, "INSERT INTO teams VALUES name = ?", team.TeamName)
	if err != nil {
		return nil, err
	}

	tx.Commit(ctx)

	return team, nil
}

func (s *Storage) Team(ctx context.Context, teamName string) (*models.Team, error) {
	var team models.Team
	if err := s.db.QueryRow(ctx, "SELECT name FROM teams WHERE name = ?", teamName).Scan(&team.TeamName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}

	rows, err := s.db.Query(ctx, "SELECT (id, username, is_active) FROM users WHERE team_name = ?", teamName)
	if err != nil {
		return nil, err
	}

	var members []models.TeamMember
	for rows.Next() {
		var member models.TeamMember
		if err = rows.Scan(&member.Username, &member.Username, &member.IsActive); err != nil {
			return nil, err
		}

		members = append(members, member)
	}

	team.Members = members

	return &team, nil
}
