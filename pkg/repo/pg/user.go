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

type User struct {
	psql sq.StatementBuilderType
	pool *pgxpool.Pool
}

func NewUser(pool *pgxpool.Pool) *User {
	return &User{
		psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		pool: pool,
	}
}

func (u *User) Create(ctx context.Context, user *domain.User) error {
	q := u.psql.Insert("users").
		Columns("user_id", "username", "team_name", "is_active").
		Values(user.UserID, user.Username, user.TeamName, user.IsActive)
	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	_, err = u.pool.Exec(ctx, sql, args...)

	if err != nil {
		return fmt.Errorf("error creating player: %w", err)
	}
	return nil
}

func (u *User) Update(ctx context.Context, userID string, patch *domain.UserUpdate) (*domain.User, error) {
	q := u.psql.Update("users").
		Suffix("RETURNING user_id, username, team_name, is_active")
	if patch.Username != nil {
		q = q.Set("username", *patch.Username)
	}
	if patch.TeamName != nil {
		q = q.Set("team_name", *patch.TeamName)
	}
	if patch.IsActive != nil {
		q = q.Set("is_active", *patch.IsActive)
	}
	q = q.Where(sq.Eq{"user_id": userID})
	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building query: %w", err)
	}

	var user domain.User
	err = u.pool.QueryRow(ctx, sql, args...).Scan(
		&user.UserID, &user.Username, &user.TeamName, &user.IsActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &domain.DomainError{Code: domain.ErrNotFound, Message: "user not found"}
		}
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	return &user, nil

}

func (u *User) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	q := u.psql.Select("user_id", "username", "team_name", "is_active").
		From("users").
		Where(sq.Eq{"user_id": userID})

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building query: %w", err)
	}

	var user domain.User
	err = u.pool.QueryRow(ctx, sql, args...).Scan(
		&user.UserID, &user.Username, &user.TeamName, &user.IsActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &domain.DomainError{Code: domain.ErrNotFound, Message: "user not found"}
		}
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return &user, nil
}

func (u *User) GetByTeamName(ctx context.Context, teamName string) ([]*domain.User, error) {
	q := u.psql.Select("user_id", "username", "team_name", "is_active").
		From("users").
		Where(sq.Eq{"team_name": teamName}).
		OrderBy("user_id")

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building query: %w", err)
	}

	rows, err := u.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, fmt.Errorf("error scanning user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

func (u *User) SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	q := u.psql.Update("users").
		Set("is_active", isActive).
		Where(sq.Eq{"user_id": userID}).
		Suffix("RETURNING user_id, username, team_name, is_active")
	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building query: %w", err)
	}

	var user domain.User
	err = u.pool.QueryRow(ctx, sql, args...).Scan(
		&user.UserID, &user.Username, &user.TeamName, &user.IsActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &domain.DomainError{Code: domain.ErrNotFound, Message: "user not found"}
		}
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	return &user, nil
}

func (u *User) GetActiveByTeamExcluding(ctx context.Context, teamName string, excludeUserIDs []string) ([]*domain.User, error) {
	q := u.psql.Select("user_id", "username", "team_name", "is_active").
		From("users").
		Where(sq.Eq{"team_name": teamName, "is_active": true})

	if len(excludeUserIDs) > 0 {
		q = q.Where(sq.NotEq{"user_id": excludeUserIDs})
	}

	q = q.OrderBy("user_id")

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building query: %w", err)
	}

	rows, err := u.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, fmt.Errorf("error scanning user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

func (u *User) Exists(ctx context.Context, userID string) (bool, error) {
	q := u.psql.Select("1").
		From("users").
		Where(sq.Eq{"user_id": userID}).
		Prefix("SELECT EXISTS (").
		Suffix(")")

	sql, args, err := q.ToSql()
	if err != nil {
		return false, fmt.Errorf("error building query: %w", err)
	}

	var exists bool
	err = u.pool.QueryRow(ctx, sql, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking user existence: %w", err)
	}
	return exists, nil
}
