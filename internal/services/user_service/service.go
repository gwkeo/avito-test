package user_service

import (
	"context"

	"github.com/gwkeo/avito-test/internal/models"
)

type UserRepository interface {
	SetIsUserActive(ctx context.Context, userID string, isActive bool) (*models.User, error)
	UsersReviews(ctx context.Context, userID string) ([]models.PullRequestShort, error)
	User(ctx context.Context, userID string) (*models.User, error)
}

type UserService struct {
	UserRepository
}

// ActiveUsersByTeam implements pull_request_service.userService.
func (s *UserService) ActiveUsersByTeam(ctx context.Context, teamID string) ([]models.User, error) {
	panic("unimplemented")
}

func New(userRepository UserRepository) *UserService {
	return &UserService{
		UserRepository: userRepository,
	}
}

func (s *UserService) UpdateFlag(ctx context.Context, userId string, flag bool) (*models.User, error) {
	return s.SetIsUserActive(ctx, userId, flag)
}

func (s *UserService) UsersPRList(ctx context.Context, userId string) ([]models.PullRequestShort, error) {
	return s.UsersReviews(ctx, userId)
}

func (s *UserService) User(ctx context.Context, userID string) (*models.User, error) {
	return s.UserRepository.User(ctx, userID)
}
