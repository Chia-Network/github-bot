package label

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/v60/github"

	"github.com/chia-network/github-bot/internal/config"
	github2 "github.com/chia-network/github-bot/internal/github"
)

// PullRequests applies internal or community labels to pull requests
func PullRequests(githubClient *github.Client, internalTeam string, cfg *config.Config) error {
	teamMembers, err := github2.GetTeamMemberList(githubClient, internalTeam)
	if err != nil {
		return fmt.Errorf("error getting team members: %w", err) // Properly handle and return error if team member list fetch fails
	}

	pullRequests, err := github2.FindCommunityPRs(cfg, teamMembers, githubClient)
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
			log.Printf("Pull Request %d by %s will be labeled %s\n", *pullRequest.Number, user, label)
			hasLabel := false
			for _, existingLabel := range pullRequest.Labels {
				if *existingLabel.Name == label {
					log.Println("Already labeled, skipping...")
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

	return nil
}
