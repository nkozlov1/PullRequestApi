package usecase

import (
	"Avito/pkg/config"
	"Avito/pkg/repo/pg"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Cases struct {
	User        *User
	Team        *Team
	PullRequest *PullRequest
}

func Setup(cfg *config.Config, pool *pgxpool.Pool) *Cases {
	userRepo := pg.NewUser(pool)
	teamRepo := pg.NewTeam(pool)
	pullRequestRepo := pg.NewPullRequest(pool)

	userCase := NewUser(userRepo)
	teamCase := NewTeam(teamRepo, userRepo)
	pullRequestCase := NewPullRequest(pullRequestRepo, userRepo, cfg.MaxCountReviewers)

	return &Cases{
		User:        userCase,
		Team:        teamCase,
		PullRequest: pullRequestCase,
	}
}
