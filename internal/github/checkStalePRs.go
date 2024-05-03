package github

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/go-github/v60/github"

	"github.com/chia-network/github-bot/internal/config"
)

// CheckstalePRs will return a list of PRs
func CheckStalePRs(githubClient *github.Client, internalTeam string, cfg config.LabelConfig) ([]string, error) {
	var stalePRUrls []string
	cutoffDate := time.Now().AddDate(0, 0, -7) // 7 days ago
	teamMembers, err := GetTeamMemberList(githubClient, internalTeam)
	if err != nil {
		return nil, err
	}

	for _, fullRepo := range cfg.LabelCheckRepos {
		log.Println("Checking repository:", fullRepo.Name)
		parts := strings.Split(fullRepo.Name, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid repository name - must contain owner and repository: %s", fullRepo.Name)
		}
		owner, repo := parts[0], parts[1]

		communityPRs, err := FindCommunityPRs(owner, repo, teamMembers, githubClient)
		if err != nil {
			return nil, err
		}

		for _, pr := range communityPRs {
			if isStale(githubClient, pr, teamMembers, cutoffDate) {
				stalePRUrls = append(stalePRUrls, pr.GetHTMLURL()) // Collecting URLs instead of PR objects
			}
		}
	}
	return stalePRUrls, nil
}

// Checks if a PR is stale based on the last update from team members
func isStale(githubClient *github.Client, pr *github.PullRequest, teamMembers map[string]bool, cutoffDate time.Time) bool {
	// Retrieve the timeline for the PR to find the latest relevant event
	listOptions := &github.ListOptions{PerPage: 100}
	for {
		// Note: As far as the GitHub API is concerned, every pull request is an issue,
		// but not every issue is a pull request.
		events, resp, err := githubClient.Issues.ListIssueTimeline(context.TODO(), pr.Base.Repo.Owner.GetLogin(), pr.Base.Repo.GetName(), pr.GetNumber(), listOptions)
		if err != nil {
			log.Printf("Failed to get timeline for PR #%d: %v", pr.GetNumber(), err)
			break // Skip to next PR on error
		}
		for _, event := range events {
			if event.Event != nil && *event.Event == "commented" && teamMembers[*event.Actor.Login] && event.CreatedAt.After(cutoffDate) {
				return false // PR has been updated by a team member recently
			}
		}
		if resp.NextPage == 0 {
			break
		}
		listOptions.Page = resp.NextPage
	}
	return true // No recent updates by team members
}

// Take the list of PRs and send a message to a keybase channel
