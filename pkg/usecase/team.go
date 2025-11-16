package usecase

import (
	"Avito/pkg/domain"
	"Avito/pkg/repo"
	"context"
	"fmt"
)

type Team struct {
	teamRepo repo.TeamRepository
	userRepo repo.UserRepository
}

func NewTeam(teamRepo repo.TeamRepository, userRepo repo.UserRepository) *Team {
	return &Team{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

func (t *Team) CreateTeam(ctx context.Context, team *domain.Team) (*domain.Team, error) {
	exists, err := t.teamRepo.Exists(ctx, team.TeamName)
	if err != nil {
		return nil, err
	}
	var createdMembers []*domain.TeamMember
	if !exists {
		if err := t.teamRepo.Create(ctx, team); err != nil {
			return nil, err
		}
	} else {
		usersOld, err := t.userRepo.GetByTeamName(ctx, team.TeamName)
		if err != nil {
			return nil, err
		}
		for _, user := range usersOld {
			createdMembers = append(createdMembers, &domain.TeamMember{
				UserID:   user.UserID,
				Username: user.Username,
				IsActive: user.IsActive,
			})
		}
	}
	for _, m := range team.Members {
		user := &domain.User{}
		existingUser, err := t.userRepo.Exists(ctx, m.UserID)
		if err != nil {
			return nil, err
		}
		if existingUser {
			updatedUser := &domain.UserUpdate{
				Username: &m.Username,
				TeamName: &team.TeamName,
				IsActive: &m.IsActive,
			}
			newUser, err := t.userRepo.Update(ctx, m.UserID, updatedUser)
			if err != nil {
				return nil, fmt.Errorf("failed to update user: %w", err)
			}
			user.UserID = newUser.UserID
			user.Username = newUser.Username
			user.TeamName = newUser.TeamName
			user.IsActive = newUser.IsActive
		} else {
			user.UserID = m.UserID
			user.Username = m.Username
			user.TeamName = team.TeamName
			user.IsActive = m.IsActive
			if err := t.userRepo.Create(ctx, user); err != nil {
				return nil, fmt.Errorf("failed to add user: %e", err)
			}
		}
		teamMembers := &domain.TeamMember{
			UserID:   user.UserID,
			Username: user.Username,
			IsActive: user.IsActive,
		}
		createdMembers = append(createdMembers, teamMembers)
	}

	team.Members = createdMembers
	return team, nil
}

func (t *Team) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	team, err := t.teamRepo.GetByName(ctx, teamName)
	if err != nil {
		return nil, err
	}
	users, err := t.userRepo.GetByTeamName(ctx, teamName)
	if err != nil {
		return nil, err
	}
	var members []*domain.TeamMember
	for _, user := range users {
		teamMembers := &domain.TeamMember{
			UserID:   user.UserID,
			Username: user.Username,
			IsActive: user.IsActive,
		}
		members = append(members, teamMembers)
	}

	team.Members = members
	return team, nil
}
