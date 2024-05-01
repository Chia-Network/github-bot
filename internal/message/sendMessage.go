package sendMessage

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/go-github/v60/github"

	"github.com/chia-network/github-bot/internal/config"
)

// PullRequests applies internal or community labels to pull requests
// Internal is determined by checking if the PR author is a member of the specified internalTeam
func PullRequests(githubClient *github.Client, internalTeam string, cfg config.LabelConfig) error {
	teamMembers := map[string]bool{}

	teamParts := strings.Split(internalTeam, "/")
	if len(teamParts) != 2 {
		return fmt.Errorf("invalid team name - must contain org and team : %s", internalTeam)
	}

	teamOpts := &github.TeamListTeamMembersOptions{
		Role: "all",
		ListOptions: github.ListOptions{
			Page:    0,
			PerPage: 100,
		},
	}

	for {
		teamOpts.ListOptions.Page++
		members, resp, err := githubClient.Teams.ListTeamMembersBySlug(context.TODO(), teamParts[0], teamParts[1], teamOpts)
		if err != nil {
			return fmt.Errorf("error getting team %s: %w", internalTeam, err)
		}

		for _, member := range members {
			teamMembers[*member.Login] = true
		}

		if resp.NextPage == 0 {
			break
		}
	}

	for _, fullRepo := range cfg.LabelCheckRepos {
		log.Println("checking repos")
		parts := strings.Split(fullRepo.Name, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid repository name - must contain owner and repository: %s", fullRepo.Name)
		}
		opts := &github.PullRequestListOptions{
			State:     "open",
			Sort:      "created",
			Direction: "desc",
			ListOptions: github.ListOptions{
				Page:    0,
				PerPage: 100,
			},
		}
		for {
			lowestNumber := 0
			opts.ListOptions.Page++
			owner := parts[0]
			repo := parts[1]
			pullRequests, resp, err := githubClient.PullRequests.List(context.TODO(), owner, repo, opts)
			if err != nil {
				return fmt.Errorf("error listing pull requests: %w", err)
			}

			for _, pullRequest := range pullRequests {
				lowestNumber = *pullRequest.Number
				if *pullRequest.Number < fullRepo.MinimumNumber {
					// Break, not continue, since our order ensures PR numbers are getting smaller
					break
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

			if resp.NextPage == 0 || lowestNumber <= fullRepo.MinimumNumber {
				break
			}
		}
	}

	return nil
}
