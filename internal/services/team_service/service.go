package team_service

import (
	"context"

	"github.com/gwkeo/avito-test/internal/models"
)

type teamRepository interface {
	AddTeam(ctx context.Context, team *models.Team) (*models.Team, error)
	Team(ctx context.Context, teamName string) (*models.Team, error)
}

type userService interface {
	User(ctx context.Context, userID string) (*models.User, error)
}

type TeamService struct {
	teamRepository
	userService
}

func New(teamRepository teamRepository, userService userService) *TeamService {
	return &TeamService{
		teamRepository: teamRepository,
		userService:    userService,
	}
}

func (s *TeamService) Add(ctx context.Context, team *models.Team) (*models.Team, error) {
	// _, err := s.Team(ctx, team.TeamName)
	// if err != nil && !errors.Is(err, storage.ErrNotFound) {
	// 	return nil, err
	// }

	// for _, member := range team.Members {
	// 	_, err := s.userService.User(ctx, member.UserID)
	// 	if err != nil && !errors.Is(err, storage.ErrNotFound) {
	// 		return nil, err
	// 	}
	// }

	return s.teamRepository.AddTeam(ctx, team)
}

func (s *TeamService) Team(ctx context.Context, teamName string) (*models.Team, error) {
	// TODO: Redis implementation
	return s.teamRepository.Team(ctx, teamName)
}

func (s *TeamService) ActiveUsersByTeam(ctx context.Context, teamID string) ([]models.User, error) {
	team, err := s.Team(ctx, teamID)
	if err != nil {
		return nil, err
	}

	res := []models.User{}
	for _, member := range team.Members {
		if member.IsActive {
			res = append(res, models.User{
				UserID:   member.UserID,
				Username: member.Username,
				TeamName: team.TeamName,
				IsActive: member.IsActive,
			})
		}
	}

	return res, nil
}
