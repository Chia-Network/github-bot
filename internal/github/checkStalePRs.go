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
func CheckStalePRs(githubClient *github.Client, internalTeam string, cfg config.CheckStalePending) ([]string, error) {
	var stalePRUrls []string
	cutoffDate := time.Now().Add(7 * 24 * time.Hour) // 7 days ago
	teamMembers, err := GetTeamMemberList(githubClient, internalTeam)
	if err != nil {
		return nil, err
	}
	communityPRs, err := FindCommunityPRs(cfg.CheckStalePending, teamMembers, githubClient)
	if err != nil {
		return nil, err
	}

	for _, pr := range communityPRs {
		stale, err := isStale(githubClient, pr, teamMembers, cutoffDate) // Handle both returned values
		if err != nil {
			log.Printf("Error checking if PR is stale: %v", err) // Log or handle the error
			continue                                             // Skip this PR or handle the error appropriately
		}
		if stale {
			stalePRUrls = append(stalePRUrls, pr.GetHTMLURL()) // Append if PR is confirmed stale
		}
	}
	return stalePRUrls, nil
}

// Checks if a PR is stale based on the last update from team members
func isStale(githubClient *github.Client, pr *github.PullRequest, teamMembers map[string]bool, cutoffDate time.Time) (bool, error) {
	// Set up a context with a timeout to control all operations within this function
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel() // Ensure resources are cleaned up correctly after the function exits

	listOptions := &github.ListOptions{PerPage: 100}
	for {
		events, resp, err := githubClient.Issues.ListIssueTimeline(ctx, pr.Base.Repo.Owner.GetLogin(), pr.Base.Repo.GetName(), pr.GetNumber(), listOptions)
		if err != nil {
			return false, fmt.Errorf("failed to get timeline for PR #%d: %w", pr.GetNumber(), err)
		}
		for _, event := range events {
			if event.Event == nil || event.Actor == nil || event.Actor.Login == nil {
				continue
			}
			if (*event.Event == "commented" || *event.Event == "reviewed") && teamMembers[*event.Actor.Login] && event.CreatedAt.After(cutoffDate) {
				return false, nil
			}
		}
		if resp.NextPage == 0 {
			break
		}
		listOptions.Page = resp.NextPage
	}
	return true, nil
}

// Take the list of PRs and send a message to a keybase channel
