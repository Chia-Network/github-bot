package github

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/go-github/v60/github"

	"github.com/chia-network/github-bot/internal/config"
)

// CheckStalePRs will return a list of PR URLs that have not been updated in the last 7 days by internal team members.
func CheckStalePRs(ctx context.Context, githubClient *github.Client, cfg *config.Config) ([]string, error) {
	var stalePRUrls []string
	cutoffDate := time.Now().Add(7 * 24 * time.Hour) // 7 days ago
	teamMembers, err := GetTeamMemberList(githubClient, cfg.InternalTeam)
	if err != nil {
		return nil, err
	}
	communityPRs, err := FindCommunityPRs(cfg, teamMembers, githubClient)
	if err != nil {
		return nil, err
	}

	for _, pr := range communityPRs {
		repoName := pr.GetBase().GetRepo().GetFullName()                      // Get the full name of the repository
		stale, err := isStale(ctx, githubClient, pr, teamMembers, cutoffDate) // Handle both returned values
		if err != nil {
			log.Printf("Error checking if PR in repo %s is stale: %v", repoName, err)
			continue // Skip this PR or handle the error appropriately
		}
		if stale {
			stalePRUrls = append(stalePRUrls, pr.GetHTMLURL()) // Append if PR is confirmed stale
		}
	}
	return stalePRUrls, nil
}

// Checks if a PR is stale based on the last update from team members
func isStale(ctx context.Context, githubClient *github.Client, pr *github.PullRequest, teamMembers map[string]bool, cutoffDate time.Time) (bool, error) {
	listOptions := &github.ListOptions{PerPage: 100}
	for {
		// Create a context for each request
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second) // 30 seconds timeout for each request
		defer cancel()
		events, resp, err := githubClient.Issues.ListIssueTimeline(ctx, pr.Base.Repo.Owner.GetLogin(), pr.Base.Repo.GetName(), pr.GetNumber(), listOptions)
		if err != nil {
			return false, fmt.Errorf("failed to get timeline for PR #%d of repo %s: %w", pr.GetNumber(), pr.Base.Repo.GetName(), err)
		}
		for _, event := range events {
			if event.Event == nil || event.Actor == nil || event.Actor.Login == nil {
				continue
			}
			if (*event.Event == "commented" || *event.Event == "reviewed") && teamMembers[*event.Actor.Login] && event.CreatedAt.After(cutoffDate) {
				return false, nil
			}
		}
		cancel() // Clean up the context at the end of the loop iteration
		if resp.NextPage == 0 {
			break
		}
		listOptions.Page = resp.NextPage
	}
	return true, nil
}
