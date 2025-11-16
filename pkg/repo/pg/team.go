package pg

import (
	"Avito/pkg/domain"
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Team struct {
	psql sq.StatementBuilderType
	pool *pgxpool.Pool
}

func NewTeam(pool *pgxpool.Pool) *Team {
	return &Team{
		psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		pool: pool,
	}
}
func (t *Team) Create(ctx context.Context, team *domain.Team) error {
	q := t.psql.Insert("teams").
		Columns("team_name").
		Values(team.TeamName)

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	_, err = t.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("error creating team: %w", err)
	}

	return nil
}

func (t *Team) GetByName(ctx context.Context, teamName string) (*domain.Team, error) {
	q := t.psql.Select("team_name").
		From("teams").
		Where(sq.Eq{"team_name": teamName})

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building query: %w", err)
	}

	var team domain.Team
	err = t.pool.QueryRow(ctx, sql, args...).Scan(&team.TeamName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &domain.DomainError{Code: domain.ErrNotFound, Message: "team not found"}
		}
		return nil, fmt.Errorf("error getting team: %w", err)
	}
	return &team, nil
}

func (t *Team) Exists(ctx context.Context, teamName string) (bool, error) {
	q := t.psql.Select("1").
		From("teams").
		Where(sq.Eq{"team_name": teamName}).
		Prefix("SELECT EXISTS (").
		Suffix(")")

	sql, args, err := q.ToSql()
	if err != nil {
		return false, fmt.Errorf("error building query: %w", err)
	}

	var exists bool
	err = t.pool.QueryRow(ctx, sql, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking team existence: %w", err)
	}

	return exists, nil
}
