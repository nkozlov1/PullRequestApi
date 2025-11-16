package repo

import (
	"Avito/pkg/domain"
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, userID string, patch *domain.UserUpdate) (*domain.User, error)
	GetByID(ctx context.Context, userID string) (*domain.User, error)
	GetByTeamName(ctx context.Context, teamName string) ([]*domain.User, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
	GetActiveByTeamExcluding(ctx context.Context, teamName string, excludeUserIDs []string) ([]*domain.User, error)
	Exists(ctx context.Context, userID string) (bool, error)
}

type TeamRepository interface {
	Create(ctx context.Context, team *domain.Team) error
	GetByName(ctx context.Context, teamName string) (*domain.Team, error)
	Exists(ctx context.Context, teamName string) (bool, error)
}

type PullRequestRepository interface {
	Create(ctx context.Context, pr *domain.PullRequest) error
	GetByID(ctx context.Context, prID string) (*domain.PullRequest, error)
	GetByIDs(ctx context.Context, prIDs []string) ([]*domain.PullRequest, error)
	AddReviewer(ctx context.Context, prID string, userID string) error
	RemoveReviewer(ctx context.Context, prID string, userID string) error
	ReplaceReviewer(ctx context.Context, prID string, oldUserID string, newUserID string) error
	SetMerged(ctx context.Context, prID string) error
	GetReviewers(ctx context.Context, prID string) ([]string, error)
	GetPRIDsByReviewer(ctx context.Context, userID string) ([]string, error)
	Exists(ctx context.Context, prID string) (bool, error)
}
