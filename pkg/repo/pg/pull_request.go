package pg

import (
	"Avito/pkg/domain"
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PullRequest struct {
	psql sq.StatementBuilderType
	pool *pgxpool.Pool
}

func NewPullRequest(pool *pgxpool.Pool) *PullRequest {
	return &PullRequest{
		psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		pool: pool,
	}
}

func (p *PullRequest) Create(ctx context.Context, pr *domain.PullRequest) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := p.psql.Insert("pull_requests").
		Columns("pull_request_id", "pull_request_name", "author_id", "status", "created_at").
		Values(pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status, pr.CreatedAt)

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}
	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("error creating pull request: %w", err)
	}

	if len(pr.AssignedReviewers) > 0 {
		reviewerQ := p.psql.Insert("pr_reviewers").
			Columns("pull_request_id", "user_id")

		for _, reviewerID := range pr.AssignedReviewers {
			reviewerQ = reviewerQ.Values(pr.PullRequestID, reviewerID)
		}
		reviewerSql, reviewerArgs, err := reviewerQ.ToSql()
		if err != nil {
			return fmt.Errorf("error building reviewers query: %w", err)
		}
		_, err = tx.Exec(ctx, reviewerSql, reviewerArgs...)
		if err != nil {
			return fmt.Errorf("error adding reviewers: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (p *PullRequest) GetByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	q := p.psql.Select("pull_request_id", "pull_request_name", "author_id", "status", "created_at", "merged_at").
		From("pull_requests").
		Where(sq.Eq{"pull_request_id": prID})

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building query: %w", err)
	}
	var pr domain.PullRequest
	err = p.pool.QueryRow(ctx, sql, args...).Scan(
		&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &domain.DomainError{Code: domain.ErrNotFound, Message: "pull request not found"}
		}
		return nil, fmt.Errorf("error getting pull request: %w", err)
	}
	return &pr, nil
}

func (p *PullRequest) GetByIDs(ctx context.Context, prIDs []string) ([]*domain.PullRequest, error) {
	if len(prIDs) == 0 {
		return []*domain.PullRequest{}, nil
	}

	q := p.psql.Select("pull_request_id", "pull_request_name", "author_id", "status", "created_at", "merged_at").
		From("pull_requests").
		Where(sq.Eq{"pull_request_id": prIDs}).
		OrderBy("created_at DESC")

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building query: %w", err)
	}

	rows, err := p.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying pull requests: %w", err)
	}
	defer rows.Close()

	var prs []*domain.PullRequest
	for rows.Next() {
		var pr domain.PullRequest
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil {
			return nil, fmt.Errorf("error scanning pull request: %w", err)
		}
		prs = append(prs, &pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pull requests: %w", err)
	}

	return prs, nil
}

func (p *PullRequest) AddReviewer(ctx context.Context, prID string, userID string) error {
	q := p.psql.Insert("pr_reviewers").
		Columns("pull_request_id", "user_id").
		Values(prID, userID).
		Suffix("ON CONFLICT (pull_request_id, user_id) DO NOTHING")

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	_, err = p.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("error adding reviewer: %w", err)
	}

	return nil
}

func (p *PullRequest) RemoveReviewer(ctx context.Context, prID string, userID string) error {
	q := p.psql.Delete("pr_reviewers").
		Where(sq.Eq{"pull_request_id": prID, "user_id": userID})

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	result, err := p.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("error removing reviewer: %w", err)
	}

	if result.RowsAffected() == 0 {
		return &domain.DomainError{Code: domain.ErrNotAssigned, Message: "reviewer is not assigned to this PR"}
	}

	return nil
}

func (p *PullRequest) ReplaceReviewer(ctx context.Context, prID string, oldUserID string, newUserID string) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	deleteQ := p.psql.Delete("pr_reviewers").
		Where(sq.Eq{"pull_request_id": prID, "user_id": oldUserID})

	deleteSql, deleteArgs, err := deleteQ.ToSql()
	if err != nil {
		return fmt.Errorf("error building delete query: %w", err)
	}

	result, err := tx.Exec(ctx, deleteSql, deleteArgs...)
	if err != nil {
		return fmt.Errorf("error removing old reviewer: %w", err)
	}

	if result.RowsAffected() == 0 {
		return &domain.DomainError{Code: domain.ErrNotAssigned, Message: "reviewer is not assigned to this PR"}
	}
	insertQ := p.psql.Insert("pr_reviewers").
		Columns("pull_request_id", "user_id").
		Values(prID, newUserID).
		Suffix("ON CONFLICT (pull_request_id, user_id) DO NOTHING")

	insertSql, insertArgs, err := insertQ.ToSql()
	if err != nil {
		return fmt.Errorf("error building insert query: %w", err)
	}

	_, err = tx.Exec(ctx, insertSql, insertArgs...)
	if err != nil {
		return fmt.Errorf("error adding new reviewer: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (p *PullRequest) SetMerged(ctx context.Context, prID string) error {
	q := p.psql.Update("pull_requests").
		Set("status", domain.PRStatusMerged).
		Set("merged_at", time.Now()).
		Where(sq.Eq{"pull_request_id": prID})

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	result, err := p.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("error setting merged status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return &domain.DomainError{Code: domain.ErrNotFound, Message: "pull request not found"}
	}

	return nil
}

func (p *PullRequest) GetReviewers(ctx context.Context, prID string) ([]string, error) {
	q := p.psql.Select("user_id").
		From("pr_reviewers").
		Where(sq.Eq{"pull_request_id": prID}).
		OrderBy("user_id")

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building query: %w", err)
	}

	rows, err := p.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying reviewers: %w", err)
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, fmt.Errorf("error scanning reviewer: %w", err)
		}
		reviewers = append(reviewers, reviewerID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reviewers: %w", err)
	}

	return reviewers, nil
}

func (p *PullRequest) GetPRIDsByReviewer(ctx context.Context, userID string) ([]string, error) {
	q := p.psql.Select("pull_request_id").
		From("pr_reviewers").
		Where(sq.Eq{"user_id": userID}).
		OrderBy("pull_request_id")

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building query: %w", err)
	}

	rows, err := p.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying pr ids: %w", err)
	}
	defer rows.Close()

	var prIDs []string
	for rows.Next() {
		var prID string
		if err := rows.Scan(&prID); err != nil {
			return nil, fmt.Errorf("error scanning pr id: %w", err)
		}
		prIDs = append(prIDs, prID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pr ids: %w", err)
	}

	return prIDs, nil
}

func (p *PullRequest) Exists(ctx context.Context, prID string) (bool, error) {
	q := p.psql.Select("1").
		From("pull_requests").
		Where(sq.Eq{"pull_request_id": prID}).
		Prefix("SELECT EXISTS (").
		Suffix(")")

	sql, args, err := q.ToSql()
	if err != nil {
		return false, fmt.Errorf("error building query: %w", err)
	}

	var exists bool
	err = p.pool.QueryRow(ctx, sql, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking PR existence: %w", err)
	}

	return exists, nil
}
