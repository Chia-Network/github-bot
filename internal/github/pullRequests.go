package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v60/github"
)

// FindCommunityPRs obtains non-teammember PRs
func FindCommunityPRs(owner string, repo string, teamMembers map[string]bool, githubClient *github.Client) ([]*github.PullRequest, error) {
	var finalPRs []*github.PullRequest
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
		opts.ListOptions.Page++
		pullRequests, resp, err := githubClient.PullRequests.List(context.TODO(), owner, repo, opts)
		if err != nil {
			return finalPRs, fmt.Errorf("error listing pull requests: %w", err)
		}

		for _, pullRequest := range pullRequests {
			user := *pullRequest.User.Login
			if !teamMembers[user] {
				finalPRs = append(finalPRs, pullRequest)
			}
		}

		if resp.NextPage == 0 {
			break
		}
	}
	return finalPRs, nil
}
