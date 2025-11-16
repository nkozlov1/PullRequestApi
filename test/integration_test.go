package test

import (
	"Avito/pkg/domain"
	"Avito/pkg/repo/pg"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	testPool *pgxpool.Pool
	ctx      = context.Background()
)

func TestMain(m *testing.M) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://root:root@localhost:5432/root?sslmode=disable"
	}

	var err error
	testPool, err = pgxpool.New(ctx, databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer testPool.Close()

	code := m.Run()
	os.Exit(code)
}

func cleanupDB(t *testing.T) {
	t.Helper()
	queries := []string{
		"DELETE FROM pr_reviewers",
		"DELETE FROM pull_requests",
		"DELETE FROM users",
		"DELETE FROM teams",
	}

	for _, query := range queries {
		_, err := testPool.Exec(ctx, query)
		if err != nil {
			t.Fatalf("Failed to cleanup database: %v", err)
		}
	}
}

func TestTeamRepository(t *testing.T) {
	cleanupDB(t)
	teamRepo := pg.NewTeam(testPool)

	t.Run("Create and Get Team", func(t *testing.T) {
		team := &domain.Team{
			TeamName: "backend-team",
		}

		err := teamRepo.Create(ctx, team)
		if err != nil {
			t.Fatalf("Failed to create team: %v", err)
		}

		retrievedTeam, err := teamRepo.GetByName(ctx, "backend-team")
		if err != nil {
			t.Fatalf("Failed to get team: %v", err)
		}

		if retrievedTeam.TeamName != team.TeamName {
			t.Errorf("Expected team name %s, got %s", team.TeamName, retrievedTeam.TeamName)
		}
	})

	t.Run("Check Team Exists", func(t *testing.T) {
		exists, err := teamRepo.Exists(ctx, "backend-team")
		if err != nil {
			t.Fatalf("Failed to check team existence: %v", err)
		}

		if !exists {
			t.Error("Expected team to exist")
		}

		exists, err = teamRepo.Exists(ctx, "non-existing-team")
		if err != nil {
			t.Fatalf("Failed to check team existence: %v", err)
		}

		if exists {
			t.Error("Expected team to not exist")
		}
	})

	t.Run("Get Non-Existing Team", func(t *testing.T) {
		_, err := teamRepo.GetByName(ctx, "non-existing-team")
		if err == nil {
			t.Error("Expected error when getting non-existing team")
		}
	})
}

func TestUserRepository(t *testing.T) {
	cleanupDB(t)
	teamRepo := pg.NewTeam(testPool)
	userRepo := pg.NewUser(testPool)
	team := &domain.Team{TeamName: "frontend-team"}
	err := teamRepo.Create(ctx, team)
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	t.Run("Create and Get User", func(t *testing.T) {
		user := &domain.User{
			UserID:   "user-1",
			Username: "alice",
			TeamName: "frontend-team",
			IsActive: true,
		}

		err := userRepo.Create(ctx, user)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		retrievedUser, err := userRepo.GetByID(ctx, "user-1")
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}
		if retrievedUser.UserID != user.UserID {
			t.Errorf("Expected user ID %s, got %s", user.UserID, retrievedUser.UserID)
		}
		if retrievedUser.Username != user.Username {
			t.Errorf("Expected username %s, got %s", user.Username, retrievedUser.Username)
		}
		if retrievedUser.TeamName != user.TeamName {
			t.Errorf("Expected team name %s, got %s", user.TeamName, retrievedUser.TeamName)
		}
		if retrievedUser.IsActive != user.IsActive {
			t.Errorf("Expected is_active %v, got %v", user.IsActive, retrievedUser.IsActive)
		}
	})

	t.Run("Update User Status", func(t *testing.T) {
		updatedUser, err := userRepo.SetIsActive(ctx, "user-1", false)
		if err != nil {
			t.Fatalf("Failed to update user status: %v", err)
		}

		if updatedUser.IsActive {
			t.Error("Expected user to be inactive")
		}
		retrievedUser, err := userRepo.GetByID(ctx, "user-1")
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		if retrievedUser.IsActive {
			t.Error("Expected user to be inactive after update")
		}
	})

	t.Run("Get Users By Team", func(t *testing.T) {
		user2 := &domain.User{
			UserID:   "user-2",
			Username: "bob",
			TeamName: "frontend-team",
			IsActive: true,
		}
		user3 := &domain.User{
			UserID:   "user-3",
			Username: "charlie",
			TeamName: "frontend-team",
			IsActive: false,
		}

		err := userRepo.Create(ctx, user2)
		if err != nil {
			t.Fatalf("Failed to create user2: %v", err)
		}
		err = userRepo.Create(ctx, user3)
		if err != nil {
			t.Fatalf("Failed to create user3: %v", err)
		}
		users, err := userRepo.GetByTeamName(ctx, "frontend-team")
		if err != nil {
			t.Fatalf("Failed to get users by team: %v", err)
		}

		if len(users) != 3 {
			t.Errorf("Expected 3 users, got %d", len(users))
		}
	})

	t.Run("Get Active Users Excluding", func(t *testing.T) {
		activeUsers, err := userRepo.GetActiveByTeamExcluding(ctx, "frontend-team", []string{"user-1"})
		if err != nil {
			t.Fatalf("Failed to get active users: %v", err)
		}
		if len(activeUsers) != 1 {
			t.Errorf("Expected 1 active user, got %d", len(activeUsers))
		}

		if len(activeUsers) > 0 && activeUsers[0].UserID != "user-2" {
			t.Errorf("Expected user-2, got %s", activeUsers[0].UserID)
		}
	})

	t.Run("Check User Exists", func(t *testing.T) {
		exists, err := userRepo.Exists(ctx, "user-1")
		if err != nil {
			t.Fatalf("Failed to check user existence: %v", err)
		}

		if !exists {
			t.Error("Expected user to exist")
		}

		exists, err = userRepo.Exists(ctx, "non-existing-user")
		if err != nil {
			t.Fatalf("Failed to check user existence: %v", err)
		}

		if exists {
			t.Error("Expected user to not exist")
		}
	})

	t.Run("Update User", func(t *testing.T) {
		user := &domain.User{
			UserID:   "user-update",
			Username: "oldname",
			TeamName: "frontend-team",
			IsActive: true,
		}
		err := userRepo.Create(ctx, user)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		newUsername := "newname"
		patch := &domain.UserUpdate{
			Username: &newUsername,
		}
		updatedUser, err := userRepo.Update(ctx, "user-update", patch)
		if err != nil {
			t.Fatalf("Failed to update user: %v", err)
		}

		if updatedUser.Username != newUsername {
			t.Errorf("Expected username %s, got %s", newUsername, updatedUser.Username)
		}
		retrievedUser, err := userRepo.GetByID(ctx, "user-update")
		if err != nil {
			t.Fatalf("Failed to get updated user: %v", err)
		}

		if retrievedUser.Username != newUsername {
			t.Errorf("Expected username %s after retrieval, got %s", newUsername, retrievedUser.Username)
		}
	})

	t.Run("Update User Multiple Fields", func(t *testing.T) {
		team2 := &domain.Team{TeamName: "backend-team"}
		err := teamRepo.Create(ctx, team2)
		if err != nil {
			t.Fatalf("Failed to create team: %v", err)
		}
		user := &domain.User{
			UserID:   "user-multi-update",
			Username: "alice",
			TeamName: "frontend-team",
			IsActive: true,
		}
		err = userRepo.Create(ctx, user)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		newUsername := "bob"
		newTeam := "backend-team"
		newActive := false
		patch := &domain.UserUpdate{
			Username: &newUsername,
			TeamName: &newTeam,
			IsActive: &newActive,
		}
		updatedUser, err := userRepo.Update(ctx, "user-multi-update", patch)
		if err != nil {
			t.Fatalf("Failed to update user: %v", err)
		}

		if updatedUser.Username != newUsername {
			t.Errorf("Expected username %s, got %s", newUsername, updatedUser.Username)
		}
		if updatedUser.TeamName != newTeam {
			t.Errorf("Expected team %s, got %s", newTeam, updatedUser.TeamName)
		}
		if updatedUser.IsActive != newActive {
			t.Errorf("Expected is_active %v, got %v", newActive, updatedUser.IsActive)
		}
	})

	t.Run("Update Non-Existing User", func(t *testing.T) {
		newUsername := "test"
		patch := &domain.UserUpdate{
			Username: &newUsername,
		}
		_, err := userRepo.Update(ctx, "non-existing-user-id", patch)
		if err == nil {
			t.Error("Expected error when updating non-existing user")
		}
	})
}

func TestPullRequestRepository(t *testing.T) {
	cleanupDB(t)
	teamRepo := pg.NewTeam(testPool)
	userRepo := pg.NewUser(testPool)
	prRepo := pg.NewPullRequest(testPool)
	team := &domain.Team{TeamName: "dev-team"}
	err := teamRepo.Create(ctx, team)
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	author := &domain.User{
		UserID:   "author-1",
		Username: "author",
		TeamName: "dev-team",
		IsActive: true,
	}
	reviewer := &domain.User{
		UserID:   "reviewer-1",
		Username: "reviewer",
		TeamName: "dev-team",
		IsActive: true,
	}

	err = userRepo.Create(ctx, author)
	if err != nil {
		t.Fatalf("Failed to create author: %v", err)
	}
	err = userRepo.Create(ctx, reviewer)
	if err != nil {
		t.Fatalf("Failed to create reviewer: %v", err)
	}

	t.Run("Create and Get Pull Request", func(t *testing.T) {
		pr := &domain.PullRequest{
			PullRequestID:     "pr-1",
			PullRequestName:   "Add new feature",
			AuthorID:          "author-1",
			Status:            domain.PRStatusOpen,
			AssignedReviewers: []string{"reviewer-1"},
			CreatedAt:         time.Now(),
		}

		err := prRepo.Create(ctx, pr)
		if err != nil {
			t.Fatalf("Failed to create pull request: %v", err)
		}
		retrievedPR, err := prRepo.GetByID(ctx, "pr-1")
		if err != nil {
			t.Fatalf("Failed to get pull request: %v", err)
		}

		if retrievedPR.PullRequestID != pr.PullRequestID {
			t.Errorf("Expected PR ID %s, got %s", pr.PullRequestID, retrievedPR.PullRequestID)
		}
		if retrievedPR.PullRequestName != pr.PullRequestName {
			t.Errorf("Expected PR name %s, got %s", pr.PullRequestName, retrievedPR.PullRequestName)
		}
		if retrievedPR.Status != pr.Status {
			t.Errorf("Expected status %s, got %s", pr.Status, retrievedPR.Status)
		}
		reviewers, err := prRepo.GetReviewers(ctx, "pr-1")
		if err != nil {
			t.Fatalf("Failed to get reviewers: %v", err)
		}
		if len(reviewers) != 1 {
			t.Errorf("Expected 1 reviewer, got %d", len(reviewers))
		}
		if len(reviewers) > 0 && reviewers[0] != "reviewer-1" {
			t.Errorf("Expected reviewer-1, got %s", reviewers[0])
		}
	})

	t.Run("Merge Pull Request", func(t *testing.T) {
		err := prRepo.SetMerged(ctx, "pr-1")
		if err != nil {
			t.Fatalf("Failed to merge pull request: %v", err)
		}
		mergedPR, err := prRepo.GetByID(ctx, "pr-1")
		if err != nil {
			t.Fatalf("Failed to get merged PR: %v", err)
		}

		if mergedPR.Status != domain.PRStatusMerged {
			t.Errorf("Expected status %s, got %s", domain.PRStatusMerged, mergedPR.Status)
		}
		if mergedPR.MergedAt == nil {
			t.Error("Expected merged_at to be set")
		}
	})

	t.Run("Check PR Exists", func(t *testing.T) {
		exists, err := prRepo.Exists(ctx, "pr-1")
		if err != nil {
			t.Fatalf("Failed to check PR existence: %v", err)
		}
		if !exists {
			t.Error("Expected PR to exist")
		}
		exists, err = prRepo.Exists(ctx, "non-existing-pr")
		if err != nil {
			t.Fatalf("Failed to check PR existence: %v", err)
		}

		if exists {
			t.Error("Expected PR to not exist")
		}
	})

	t.Run("Add and Remove Reviewers", func(t *testing.T) {
		pr := &domain.PullRequest{
			PullRequestID:     "pr-2",
			PullRequestName:   "Test reviewers",
			AuthorID:          "author-1",
			Status:            domain.PRStatusOpen,
			AssignedReviewers: []string{},
			CreatedAt:         time.Now(),
		}

		err := prRepo.Create(ctx, pr)
		if err != nil {
			t.Fatalf("Failed to create PR: %v", err)
		}
		err = prRepo.AddReviewer(ctx, "pr-2", "reviewer-1")
		if err != nil {
			t.Fatalf("Failed to add reviewer: %v", err)
		}
		reviewers, err := prRepo.GetReviewers(ctx, "pr-2")
		if err != nil {
			t.Fatalf("Failed to get reviewers: %v", err)
		}
		if len(reviewers) != 1 {
			t.Errorf("Expected 1 reviewer, got %d", len(reviewers))
		}
		err = prRepo.RemoveReviewer(ctx, "pr-2", "reviewer-1")
		if err != nil {
			t.Fatalf("Failed to remove reviewer: %v", err)
		}
		reviewers, err = prRepo.GetReviewers(ctx, "pr-2")
		if err != nil {
			t.Fatalf("Failed to get reviewers after removal: %v", err)
		}
		if len(reviewers) != 0 {
			t.Errorf("Expected 0 reviewers, got %d", len(reviewers))
		}
	})

	t.Run("Replace Reviewer", func(t *testing.T) {
		reviewer2 := &domain.User{
			UserID:   "reviewer-2",
			Username: "reviewer2",
			TeamName: "dev-team",
			IsActive: true,
		}
		err := userRepo.Create(ctx, reviewer2)
		if err != nil {
			t.Fatalf("Failed to create reviewer2: %v", err)
		}
		pr := &domain.PullRequest{
			PullRequestID:     "pr-3",
			PullRequestName:   "Replace reviewer test",
			AuthorID:          "author-1",
			Status:            domain.PRStatusOpen,
			AssignedReviewers: []string{"reviewer-1"},
			CreatedAt:         time.Now(),
		}
		err = prRepo.Create(ctx, pr)
		if err != nil {
			t.Fatalf("Failed to create PR: %v", err)
		}
		err = prRepo.ReplaceReviewer(ctx, "pr-3", "reviewer-1", "reviewer-2")
		if err != nil {
			t.Fatalf("Failed to replace reviewer: %v", err)
		}
		reviewers, err := prRepo.GetReviewers(ctx, "pr-3")
		if err != nil {
			t.Fatalf("Failed to get reviewers: %v", err)
		}
		if len(reviewers) != 1 {
			t.Errorf("Expected 1 reviewer, got %d", len(reviewers))
		}
		if len(reviewers) > 0 && reviewers[0] != "reviewer-2" {
			t.Errorf("Expected reviewer-2, got %s", reviewers[0])
		}
	})

	t.Run("Get PRs by Reviewer", func(t *testing.T) {
		prIDs, err := prRepo.GetPRIDsByReviewer(ctx, "reviewer-2")
		if err != nil {
			t.Fatalf("Failed to get PR IDs by reviewer: %v", err)
		}

		if len(prIDs) != 1 {
			t.Errorf("Expected 1 PR for reviewer-2, got %d", len(prIDs))
		}
		if len(prIDs) > 0 && prIDs[0] != "pr-3" {
			t.Errorf("Expected pr-3, got %s", prIDs[0])
		}
	})

	t.Run("Get PRs by IDs", func(t *testing.T) {
		pr4 := &domain.PullRequest{
			PullRequestID:     "pr-4",
			PullRequestName:   "Feature A",
			AuthorID:          "author-1",
			Status:            domain.PRStatusOpen,
			AssignedReviewers: []string{"reviewer-1"},
			CreatedAt:         time.Now(),
		}
		pr5 := &domain.PullRequest{
			PullRequestID:     "pr-5",
			PullRequestName:   "Feature B",
			AuthorID:          "author-1",
			Status:            domain.PRStatusOpen,
			AssignedReviewers: []string{"reviewer-1"},
			CreatedAt:         time.Now().Add(-1 * time.Hour),
		}

		err := prRepo.Create(ctx, pr4)
		if err != nil {
			t.Fatalf("Failed to create pr-4: %v", err)
		}
		err = prRepo.Create(ctx, pr5)
		if err != nil {
			t.Fatalf("Failed to create pr-5: %v", err)
		}
		prs, err := prRepo.GetByIDs(ctx, []string{"pr-4", "pr-5"})
		if err != nil {
			t.Fatalf("Failed to get PRs by IDs: %v", err)
		}

		if len(prs) != 2 {
			t.Errorf("Expected 2 PRs, got %d", len(prs))
		}
		if len(prs) >= 2 {
			if prs[0].PullRequestID != "pr-4" {
				t.Errorf("Expected first PR to be pr-4 (newer), got %s", prs[0].PullRequestID)
			}
			if prs[1].PullRequestID != "pr-5" {
				t.Errorf("Expected second PR to be pr-5 (older), got %s", prs[1].PullRequestID)
			}
		}
	})

	t.Run("Get PRs by IDs - Empty List", func(t *testing.T) {
		prs, err := prRepo.GetByIDs(ctx, []string{})
		if err != nil {
			t.Fatalf("Failed to get PRs with empty list: %v", err)
		}

		if len(prs) != 0 {
			t.Errorf("Expected 0 PRs for empty list, got %d", len(prs))
		}
	})

	t.Run("Get PRs by IDs - Non-Existing IDs", func(t *testing.T) {
		prs, err := prRepo.GetByIDs(ctx, []string{"non-existing-pr-1", "non-existing-pr-2"})
		if err != nil {
			t.Fatalf("Failed to get PRs with non-existing IDs: %v", err)
		}

		if len(prs) != 0 {
			t.Errorf("Expected 0 PRs for non-existing IDs, got %d", len(prs))
		}
	})
}

func TestComplexScenario(t *testing.T) {
	cleanupDB(t)
	teamRepo := pg.NewTeam(testPool)
	userRepo := pg.NewUser(testPool)
	prRepo := pg.NewPullRequest(testPool)
	t.Run("Full Workflow", func(t *testing.T) {
		team := &domain.Team{TeamName: "fullstack-team"}
		err := teamRepo.Create(ctx, team)
		if err != nil {
			t.Fatalf("Failed to create team: %v", err)
		}
		users := []*domain.User{
			{UserID: "fs-1", Username: "john", TeamName: "fullstack-team", IsActive: true},
			{UserID: "fs-2", Username: "jane", TeamName: "fullstack-team", IsActive: true},
			{UserID: "fs-3", Username: "jack", TeamName: "fullstack-team", IsActive: false},
		}

		for _, user := range users {
			err := userRepo.Create(ctx, user)
			if err != nil {
				t.Fatalf("Failed to create user %s: %v", user.UserID, err)
			}
		}
		pr := &domain.PullRequest{
			PullRequestID:     "pr-complex-1",
			PullRequestName:   "Refactor database layer",
			AuthorID:          "fs-1",
			Status:            domain.PRStatusOpen,
			AssignedReviewers: []string{"fs-2"},
			CreatedAt:         time.Now(),
		}

		err = prRepo.Create(ctx, pr)
		if err != nil {
			t.Fatalf("Failed to create PR: %v", err)
		}
		teamExists, err := teamRepo.Exists(ctx, "fullstack-team")
		if err != nil || !teamExists {
			t.Error("Team should exist")
		}

		teamUsers, err := userRepo.GetByTeamName(ctx, "fullstack-team")
		if err != nil {
			t.Fatalf("Failed to get team users: %v", err)
		}
		if len(teamUsers) != 3 {
			t.Errorf("Expected 3 users in team, got %d", len(teamUsers))
		}

		retrievedPR, err := prRepo.GetByID(ctx, "pr-complex-1")
		if err != nil {
			t.Fatalf("Failed to get PR: %v", err)
		}
		if retrievedPR.Status != domain.PRStatusOpen {
			t.Errorf("Expected PR status OPEN, got %s", retrievedPR.Status)
		}
		_, err = userRepo.SetIsActive(ctx, "fs-2", false)
		if err != nil {
			t.Fatalf("Failed to deactivate user: %v", err)
		}
		activeUsers, err := userRepo.GetActiveByTeamExcluding(ctx, "fullstack-team", []string{})
		if err != nil {
			t.Fatalf("Failed to get active users: %v", err)
		}
		if len(activeUsers) != 1 {
			t.Errorf("Expected 1 active user, got %d", len(activeUsers))
		}
		err = prRepo.SetMerged(ctx, "pr-complex-1")
		if err != nil {
			t.Fatalf("Failed to merge PR: %v", err)
		}

		finalPR, err := prRepo.GetByID(ctx, "pr-complex-1")
		if err != nil {
			t.Fatalf("Failed to get final PR: %v", err)
		}
		if finalPR.Status != domain.PRStatusMerged {
			t.Errorf("Expected PR status MERGED, got %s", finalPR.Status)
		}
	})
}
