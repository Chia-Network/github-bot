package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/google/go-github/v60/github"

	"github.com/chia-network/github-bot/internal/config"
)

// StalePR holds information about pending PRs
type StalePR struct {
	Repo     string
	PRNumber int
	URL      string
}

// CheckStalePRs will return a list of PR URLs that have not been updated in the last 7 days by internal team members.
func CheckStalePRs(ctx context.Context, githubClient *github.Client, cfg *config.Config) ([]StalePR, error) {
	var stalePRs []StalePR
	cutoffDate := time.Now().Add(-7 * 24 * time.Hour) // 7 days ago
	teamMembers, err := GetTeamMemberList(githubClient, cfg.InternalTeam)
	if err != nil {
		return nil, err
	}

	for _, fullRepo := range cfg.CheckRepos {
		slogs.Logger.Info("Checking repository", "repository", fullRepo.Name)
		parts := strings.Split(fullRepo.Name, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid repository name - must contain owner and repository: %s", fullRepo.Name)
		}
		owner, repo := parts[0], parts[1]

		communityPRs, err := FindCommunityPRs(cfg, teamMembers, githubClient, owner, repo, fullRepo.MinimumNumber)
		if err != nil {
			return nil, err
		}

		for _, pr := range communityPRs {
			repoName := pr.GetBase().GetRepo().GetFullName() // Get the full name of the repository
			slogs.Logger.Info("Checking PR", "PR", pr.GetHTMLURL())
			stale, err := isStale(ctx, githubClient, pr, teamMembers, cutoffDate) // Handle both returned values
			if err != nil {
				slogs.Logger.Error("Error checking if PR is stale", "repository", repoName, "error", err)
				continue // Skip this PR or handle the error appropriately
			}
			if stale {
				slogs.Logger.Info("PR is has no team member within the last seven days", "PR", pr.GetNumber(), "repository", fullRepo.Name, "user", pr.User.GetLogin(), "created_at", pr.CreatedAt)
				stalePRs = append(stalePRs, StalePR{
					Repo:     repo,
					PRNumber: pr.GetNumber(),
					URL:      pr.GetHTMLURL(),
				})
			} else {
				slogs.Logger.Info("PR is not stale",
					"PR", pr.GetNumber(),
					"repository", fullRepo.Name)
			}
		}
	}

	return stalePRs, nil
}

// Checks if a PR is stale based on the last update from team members
func isStale(ctx context.Context, githubClient *github.Client, pr *github.PullRequest, teamMembers map[string]bool, cutoffDate time.Time) (bool, error) {
	listOptions := &github.ListOptions{PerPage: 100}
	for {
		// Create a context for each request
		staleCtx, staleCancel := context.WithTimeout(ctx, 30*time.Second) // 30 seconds timeout for each request
		defer staleCancel()
		events, resp, err := githubClient.Issues.ListIssueTimeline(staleCtx, pr.Base.Repo.Owner.GetLogin(), pr.Base.Repo.GetName(), pr.GetNumber(), listOptions)
		if err != nil {
			slogs.Logger.Error("Failed to get timeline for PR", "PR", pr.GetNumber(), "repository", pr.Base.Repo.GetName(), "error", err)
			return false, err
		}
		for _, event := range events {
			if event.Event == nil || event.Actor == nil || event.Actor.Login == nil {
				continue
			}
			if (*event.Event == "commented" || *event.Event == "reviewed") && teamMembers[*event.Actor.Login] && event.CreatedAt.After(cutoffDate) {
				return false, nil
			}
		}
		staleCancel() // Clean up the context at the end of the loop iteration
		if resp.NextPage == 0 {
			break
		}
		listOptions.Page = resp.NextPage
	}
	return true, nil
}
