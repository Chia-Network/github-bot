package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v60/github"

	"github.com/chia-network/github-bot/internal/config"
)

// FindCommunityPRs obtains PRs based on provided filters
func FindCommunityPRs(cfg *config.Config, teamMembers map[string]bool, githubClient *github.Client, owner string, repo string, minimumNumber int) ([]*github.PullRequest, error) {
	var finalPRs []*github.PullRequest
	opts := &github.PullRequestListOptions{
		State:     "open",
		Sort:      "created",
		Direction: "desc",
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	for {
		pullRequests, resp, err := githubClient.PullRequests.List(context.Background(), owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("error listing pull requests for %s/%s: %w", owner, repo, err)
		}

		for _, pullRequest := range pullRequests {
			if *pullRequest.Number < minimumNumber {
				break
			}
			if *pullRequest.Draft {
				continue
			}
			user := *pullRequest.User.Login
			if !teamMembers[user] && !cfg.SkipUsersMap[user] {
				finalPRs = append(finalPRs, pullRequest)
			}
		}

		if resp.NextPage == 0 {
			break // Exit the loop if there are no more pages
		}
		opts.Page = resp.NextPage // Set next page number
	}
	return finalPRs, nil
}
