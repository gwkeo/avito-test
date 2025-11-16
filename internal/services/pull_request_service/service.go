package pull_request_service

import (
	"context"

	"github.com/gwkeo/avito-test/internal/models"
	"github.com/gwkeo/avito-test/internal/storage"
)

const (
	MAX_REVIEWERS_COUNT = 2
)

type pullRequestRepository interface {
	PullRequest(ctx context.Context, pullRequestID string) (*models.PullRequest, error)
	CreatePullRequest(ctx context.Context, pullRequestID, pullRequestName, authorID string, assignedReviewers []string) (*models.PullRequest, error)
	MergePullRequest(ctx context.Context, pullRequestID string) (*models.PullRequest, error)
	ReassignPullRequest(ctx context.Context, pullRequestID, oldReviewerID, newReviewerID string) (*models.PullRequest, string, error)
}

type userService interface {
	User(ctx context.Context, userID string) (*models.User, error)
}

type teamService interface {
	Team(ctx context.Context, teamName string) (*models.Team, error)
	ActiveUsersByTeam(ctx context.Context, teamID string) ([]models.User, error)
}

type PullRequestService struct {
	pullRequestRepository
	userService
	teamService
}

func New(
	pullRequestRepository pullRequestRepository,
	userService userService,
	teamService teamService,
) *PullRequestService {
	return &PullRequestService{
		pullRequestRepository: pullRequestRepository,
		userService:           userService,
		teamService:           teamService,
	}
}

func (s *PullRequestService) usersTeamActiveMembersExceptHim(ctx context.Context, userID string) ([]string, error) {
	user, err := s.userService.User(ctx, userID)
	if err != nil {
		return nil, err
	}

	usersTeamsActiveMembers, err := s.teamService.ActiveUsersByTeam(ctx, user.TeamName)
	if err != nil {
		return nil, err
	}

	reviewers := []string{}
	reviewersCount := 0
	// TODO: рандомное или контролируемое назначение,
	// чтобы люди в конце списка активных участников
	// команды хоть раз ревьюили
	for _, v := range usersTeamsActiveMembers {
		if v.UserID != userID {
			reviewers[reviewersCount] = v.UserID
			reviewersCount++
		}
		if reviewersCount == MAX_REVIEWERS_COUNT {
			break
		}
	}

	return reviewers, nil
}

func (s *PullRequestService) Create(ctx context.Context, ID, name, authorID string) (*models.PullRequest, error) {
	_, err := s.pullRequestRepository.PullRequest(ctx, ID)
	if err == nil {
		// Ошибки нет, так как PR существует - возвращаем ошибку PR_EXISTS
		return nil, storage.ErrPRExists
	}

	reviewers, err := s.usersTeamActiveMembersExceptHim(ctx, authorID)
	if err != nil {
		return nil, err
	}

	return s.pullRequestRepository.CreatePullRequest(ctx, ID, name, authorID, reviewers)
}

func (s *PullRequestService) Merge(ctx context.Context, ID string) (*models.PullRequest, error) {
	pr, err := s.pullRequestRepository.PullRequest(ctx, ID)
	if err != nil {
		return nil, err
	}

	if pr.Status == storage.PullRequestStateMerged {
		// Merge PR - идемпотентен, значит можно
		// осуществять эту операцию несколько раз
		// и это не будет ошибкой
		return pr, nil
	}

	return s.pullRequestRepository.MergePullRequest(ctx, ID)
}

func (s *PullRequestService) Reassign(ctx context.Context, ID, oldReviewerID string) (*models.PullRequest, string, error) {
	pr, err := s.pullRequestRepository.PullRequest(ctx, ID)
	if err != nil {
		return pr, "", err
	}

	if pr.Status == storage.PullRequestStateMerged {
		return nil, "", storage.ErrPRMerged
	}

	potentialReviewers, err := s.usersTeamActiveMembersExceptHim(ctx, oldReviewerID)
	if err != nil {
		return nil, "", err
	}

	if len(potentialReviewers) == 0 {
		return nil, "", storage.ErrNoCandidate
	}

	return s.pullRequestRepository.ReassignPullRequest(ctx, ID, oldReviewerID, potentialReviewers[0])
}
