package usecase

import (
	"Avito/pkg/domain"
	"Avito/pkg/repo"
	"context"
)

type User struct {
	userRepo repo.UserRepository
}

func NewUser(userRepo repo.UserRepository) *User {
	return &User{
		userRepo: userRepo,
	}
}

func (u *User) SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	exists, err := u.userRepo.Exists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.NewDomainError(domain.ErrNotFound, "user not found")
	}
	user, err := u.userRepo.SetIsActive(ctx, userID, isActive)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *User) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}
