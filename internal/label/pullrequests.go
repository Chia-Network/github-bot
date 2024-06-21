package label

import (
	"context"
	"fmt"
	"strings"

	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/google/go-github/v60/github"

	"github.com/chia-network/github-bot/internal/config"
	github2 "github.com/chia-network/github-bot/internal/github"
)

// PullRequests applies internal or community labels to pull requests
func PullRequests(githubClient *github.Client, cfg *config.Config) error {
	teamMembers, err := github2.GetTeamMemberList(githubClient, cfg.InternalTeam)
	if err != nil {
		return fmt.Errorf("error getting team members: %w", err) // Properly handle and return error if team member list fetch fails
	}

	for _, fullRepo := range cfg.CheckRepos {
		slogs.Logr.Info("Checking repository", "repository", fullRepo.Name)
		parts := strings.Split(fullRepo.Name, "/")
		if len(parts) != 2 {
			slogs.Logr.Error("Invalid repository name - must contain owner and repository", "repository", fullRepo.Name)
			continue
		}
		owner, repo := parts[0], parts[1]

		pullRequests, err := github2.FindCommunityPRs(cfg, teamMembers, githubClient, owner, repo, fullRepo.MinimumNumber)
		if err != nil {
			return fmt.Errorf("error finding community PRs: %w", err) // Handle error from finding community PRs
		}

		for _, pullRequest := range pullRequests {
			user := *pullRequest.User.Login

			var label string
			if teamMembers[user] {
				label = cfg.LabelInternal
			} else {
				label = cfg.LabelExternal
			}
			if label != "" {
				slogs.Logr.Info("Labeling pull request", "PR", *pullRequest.Number, "user", user, "label", label)
				hasLabel := false
				for _, existingLabel := range pullRequest.Labels {
					if *existingLabel.Name == label {
						slogs.Logr.Info("Already labeled, skipping", "PR", *pullRequest.Number, "label", label)
						hasLabel = true
						break
					}
				}

				if !hasLabel {
					allLabels := []string{label}
					for _, labelP := range pullRequest.Labels {
						allLabels = append(allLabels, *labelP.Name)
					}
					_, _, err := githubClient.Issues.AddLabelsToIssue(context.TODO(), *pullRequest.Base.Repo.Owner.Login, *pullRequest.Base.Repo.Name, *pullRequest.Number, allLabels)
					if err != nil {
						return fmt.Errorf("error adding labels to pull request %d: %w", *pullRequest.Number, err) // Ensure error from label adding is handled
					}
				}
			}
		}
	}
	return nil
}
