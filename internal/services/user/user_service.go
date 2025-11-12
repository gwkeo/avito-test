package user

import (
	"context"

	"github.com/gwkeo/avito-test/internal/models"
)

type creator interface {
	CreateUsers(ctx context.Context, users []models.User)
}

type updater interface {
	UpdateIsActiveFlag(ctx context.Context, userId string)
}

type UserService struct {
	creator
	updater
}
