package usecase

import (
	"Avito/pkg/domain"
	"Avito/pkg/repo"
	"context"
	"math/rand"
	"time"
)

type PullRequest struct {
	prRepo            repo.PullRequestRepository
	userRepo          repo.UserRepository
	maxCountReviewers int
}

func NewPullRequest(prRepo repo.PullRequestRepository, userRepo repo.UserRepository, maxCountReviewers int) *PullRequest {
	return &PullRequest{
		prRepo:            prRepo,
		userRepo:          userRepo,
		maxCountReviewers: maxCountReviewers,
	}
}

func (p *PullRequest) CreatePullRequest(ctx context.Context, prID, prName, authorID string) (*domain.PullRequest, error) {
	exists, err := p.prRepo.Exists(ctx, prID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.NewDomainError(domain.ErrPRExists, "PR id already exists")
	}

	author, err := p.userRepo.GetByID(ctx, authorID)
	if err != nil {
		return nil, err
	}

	candidates, err := p.userRepo.GetActiveByTeamExcluding(ctx, author.TeamName, []string{authorID})
	if err != nil {
		return nil, err
	}

	reviewers := selectRandomReviewers(candidates, p.maxCountReviewers)

	pr := &domain.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            domain.PRStatusOpen,
		AssignedReviewers: reviewers,
		CreatedAt:         time.Now(),
		MergedAt:          nil,
	}

	if err := p.prRepo.Create(ctx, pr); err != nil {
		return nil, err
	}

	return pr, nil
}

func (p *PullRequest) MergePullRequest(ctx context.Context, prID string) (*domain.PullRequest, error) {
	pr, err := p.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, err
	}
	if pr.Status == domain.PRStatusMerged {
		reviewers, err := p.prRepo.GetReviewers(ctx, prID)
		if err != nil {
			return nil, err
		}
		pr.AssignedReviewers = reviewers
		return pr, nil
	}
	if err := p.prRepo.SetMerged(ctx, prID); err != nil {
		return nil, err
	}
	pr, err = p.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, err
	}
	reviewers, err := p.prRepo.GetReviewers(ctx, prID)
	if err != nil {
		return nil, err
	}
	pr.AssignedReviewers = reviewers

	return pr, nil
}

func (p *PullRequest) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, string, error) {
	pr, err := p.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, "", err
	}
	if pr.Status == domain.PRStatusMerged {
		return nil, "", domain.NewDomainError(domain.ErrPRMerged, "cannot reassign on merged PR")
	}
	reviewers, err := p.prRepo.GetReviewers(ctx, prID)
	if err != nil {
		return nil, "", err
	}
	isAssigned := false
	for _, reviewerID := range reviewers {
		if reviewerID == oldReviewerID {
			isAssigned = true
			break
		}
	}
	if !isAssigned {
		return nil, "", domain.NewDomainError(domain.ErrNotAssigned, "reviewer is not assigned to this PR")
	}
	oldReviewer, err := p.userRepo.GetByID(ctx, oldReviewerID)
	if err != nil {
		return nil, "", err
	}
	author := pr.AuthorID
	excludeIDs := append(reviewers, author)
	candidates, err := p.userRepo.GetActiveByTeamExcluding(ctx, oldReviewer.TeamName, excludeIDs)
	if err != nil {
		return nil, "", err
	}
	if len(candidates) == 0 {
		return nil, "", domain.NewDomainError(domain.ErrNoCandidate, "no active replacement candidate in team")
	}
	newReviewers := selectRandomReviewers(candidates, 1)
	if len(newReviewers) == 0 {
		return nil, "", domain.NewDomainError(domain.ErrNoCandidate, "no active replacement candidate in team")
	}
	newReviewerID := newReviewers[0]
	if err := p.prRepo.ReplaceReviewer(ctx, prID, oldReviewerID, newReviewerID); err != nil {
		return nil, "", err
	}
	pr, err = p.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, "", err
	}
	updatedReviewers, err := p.prRepo.GetReviewers(ctx, prID)
	if err != nil {
		return nil, "", err
	}
	pr.AssignedReviewers = updatedReviewers

	return pr, newReviewerID, nil
}

func (p *PullRequest) GetUserReviews(ctx context.Context, userID string) ([]*domain.PullRequest, error) {
	exists, err := p.userRepo.Exists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.NewDomainError(domain.ErrNotFound, "user not found")
	}
	prIDs, err := p.prRepo.GetPRIDsByReviewer(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(prIDs) == 0 {
		return []*domain.PullRequest{}, nil
	}

	prs, err := p.prRepo.GetByIDs(ctx, prIDs)
	if err != nil {
		return nil, err
	}
	for _, pr := range prs {
		reviewers, err := p.prRepo.GetReviewers(ctx, pr.PullRequestID)
		if err != nil {
			return nil, err
		}
		pr.AssignedReviewers = reviewers
	}

	return prs, nil
}

func selectRandomReviewers(candidates []*domain.User, maxCount int) []string {
	if len(candidates) == 0 {
		return []string{}
	}
	if len(candidates) <= maxCount {
		reviewers := make([]string, len(candidates))
		for i, user := range candidates {
			reviewers[i] = user.UserID
		}
		return reviewers
	}
	rand.Seed(time.Now().UnixNano())
	shuffled := make([]*domain.User, len(candidates))
	copy(shuffled, candidates)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	reviewers := make([]string, maxCount)
	for i := 0; i < maxCount; i++ {
		reviewers[i] = shuffled[i].UserID
	}

	return reviewers
}
