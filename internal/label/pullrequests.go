package label

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/go-github/v60/github"

	"github.com/chia-network/github-bot/internal/config"
	github2 "github.com/chia-network/github-bot/internal/github"
)

// PullRequests applies internal or community labels to pull requests
// Internal is determined by checking if the PR author is a member of the specified internalTeam
func PullRequests(githubClient *github.Client, internalTeam string, cfg config.LabelConfig) error {
	teamMembers, _ := github2.GetTeamMemberList(githubClient, internalTeam)
	for _, fullRepo := range cfg.LabelCheckRepos {
		log.Println("checking repos")
		parts := strings.Split(fullRepo.Name, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid repository name - must contain owner and repository: %s", fullRepo.Name)
		}
		owner := parts[0]
		repo := parts[1]
		pullRequests, err := github2.FindCommunityPRs(owner, repo, teamMembers, githubClient)
		if err != nil {
			return err
		}

		for _, pullRequest := range pullRequests {
			if *pullRequest.Number < fullRepo.MinimumNumber {
				continue
			}
			if *pullRequest.Draft {
				continue
			}
			user := *pullRequest.User.Login
			if cfg.LabelSkipMap[user] {
				continue
			}
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
						log.Println("  Already labeled, skipping...")
						hasLabel = true
						break
					}
				}

				if !hasLabel {
					allLabels := []string{label}
					for _, labelP := range pullRequest.Labels {
						allLabels = append(allLabels, *labelP.Name)
					}
					_, _, err := githubClient.Issues.AddLabelsToIssue(context.TODO(), owner, repo, *pullRequest.Number, allLabels)
					if err != nil {
						return fmt.Errorf("error adding labels to pull request %d: %w", *pullRequest.Number, err)
					}
				}
			}
		}
	}

	return nil
}
