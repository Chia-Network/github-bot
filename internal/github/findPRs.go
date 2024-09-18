package github

import (
	"context"
	"fmt"

	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/google/go-github/v60/github"

	"github.com/chia-network/github-bot/internal/config"
)

// FindCommunityPRs obtains PRs based on provided filters for community members
func FindCommunityPRs(cfg *config.Config, teamMembers map[string]bool, githubClient *github.Client, owner string, repo string, minimumNumber int) ([]*github.PullRequest, error) {
	return findPRs(cfg, teamMembers, githubClient, owner, repo, minimumNumber, true)
}

// FindAllPRs obtains all PRs for the repository
func FindAllPRs(cfg *config.Config, teamMembers map[string]bool, githubClient *github.Client, owner string, repo string, minimumNumber int) ([]*github.PullRequest, error) {
	return findPRs(cfg, teamMembers, githubClient, owner, repo, minimumNumber, false)
}

// findPRs handles fetching and filtering PRs based on community or all contributors
func findPRs(cfg *config.Config, teamMembers map[string]bool, githubClient *github.Client, owner string, repo string, minimumNumber int, filterCommunity bool) ([]*github.PullRequest, error) {
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
			// If filtering community PRs, skip PRs by internal team members and users in SkipUsersMap
			if filterCommunity && (teamMembers[user] || cfg.SkipUsersMap[user]) {
				slogs.Logr.Info("Pull request does not meet criteria, skipping", "PR", pullRequest.GetHTMLURL(), "user", user)
				continue
			}

			slogs.Logr.Info("Pull request meets criteria, adding to final list", "PR", pullRequest.GetHTMLURL(), "user", user)
			finalPRs = append(finalPRs, pullRequest)
		}

		if resp.NextPage == 0 {
			break // Exit the loop if there are no more pages
		}
		opts.Page = resp.NextPage // Set next page number
	}

	return finalPRs, nil
}
