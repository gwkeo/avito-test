package team_service

import (
	"context"

	"github.com/gwkeo/avito-test/internal/models"
)

type TeamRepository interface {
	AddTeam(ctx context.Context, team models.Team) (models.Team, error)
	Team(ctx context.Context, teamName string) (models.Team, error)
}

type TeamService struct {
	TeamRepository
}

func New(teamRepository TeamRepository) *TeamService {
	return &TeamService{
		TeamRepository: teamRepository,
	}
}

func (r *TeamService) Add(ctx context.Context, team models.Team) (models.Team, error) {
	// TODO: queue / Kafka implementation
	return r.TeamRepository.AddTeam(ctx, team)
}

func (r *TeamService) Team(ctx context.Context, teamName string) (models.Team, error) {
	// TODO: Redis implementation
	return r.TeamRepository.Team(ctx, teamName)
}
