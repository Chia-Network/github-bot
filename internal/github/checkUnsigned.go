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

// UnsignedPRs holds information about pending PRs
type UnsignedPRs struct {
	Owner    string
	Repo     string
	PRNumber int
	URL      string
}

// CheckUnsignedCommits will return a list of PR URLs that have not been updated in the last 7 days by internal team members.
func CheckUnsignedCommits(ctx context.Context, githubClient *github.Client, cfg *config.Config) ([]UnsignedPRs, error) {
	var unsignedPRs []UnsignedPRs
	teamMembers, err := GetTeamMemberList(githubClient, cfg.InternalTeam)
	if err != nil {
		return nil, err
	}

	for _, fullRepo := range cfg.CheckRepos {
		slogs.Logr.Info("Checking repository", "repository", fullRepo.Name)
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
			slogs.Logr.Info("Checking if PR has unsigned commits", "PR", pr.GetHTMLURL())
			unsigned, err := hasUnsignedCommits(ctx, githubClient, pr, teamMembers) // Handle both returned values
			if err != nil {
				slogs.Logr.Error("Error checking if PR has unsigned commits", "repository", repoName, "error", err)
				continue // Skip this PR or handle the error appropriately
			}
			if unsigned {
				slogs.Logr.Info("PR has unsigned commits", "PR", pr.GetNumber(), "repository", fullRepo.Name, "user", pr.User.GetLogin(), "created_at", pr.CreatedAt)
				unsignedPRs = append(unsignedPRs, UnsignedPRs{
					Owner:    owner,
					Repo:     repo,
					PRNumber: pr.GetNumber(),
					URL:      pr.GetHTMLURL(),
				})

			} else {
				slogs.Logr.Info("No commits are unsigned",
					"PR", pr.GetNumber(),
					"repository", fullRepo.Name)
			}
		}
	}

	return unsignedPRs, nil
}

// Checks if a PR has unsigned commits based on the last update from team members
func hasUnsignedCommits(ctx context.Context, githubClient *github.Client, pr *github.PullRequest, teamMembers map[string]bool) (bool, error) {
	listOptions := &github.ListOptions{PerPage: 100}
	for {
		// Create a context for each request
		unsignedCtx, unsignedCancel := context.WithTimeout(ctx, 30*time.Second) // 30 seconds timeout for each request
		defer unsignedCancel()
		commits, resp, err := githubClient.PullRequests.ListCommits(unsignedCtx, pr.Base.Repo.Owner.GetLogin(), pr.Base.Repo.GetName(), pr.GetNumber(), listOptions)
		if err != nil {
			slogs.Logr.Error("Failed to get commits for PR", "PR", pr.GetNumber(), "repository", pr.Base.Repo.GetName(), "error", err)
			return false, err
		}
		// Check each commit to see if it is signed
		for _, commit := range commits {
			if commit == nil || commit.Commit == nil {
				continue
			}
			verification := commit.Commit.Verification
			if verification == nil || !*verification.Verified {
				return true, nil
			}
		}
		unsignedCancel() // Clean up the context at the end of the loop iteration
		if resp.NextPage == 0 {
			break
		}
		listOptions.Page = resp.NextPage
	}
	return true, nil
}

// CheckAndComment checks for the specific comment from a specific account and posts a comment if it doesn't exist.
func CheckAndComment(ctx context.Context, client *github.Client, owner, repo string, prNumber int) error {
	commentAuthor := "ChiaAutomation"
	commentBody := "Your commits are not signed and our branch protection rules require signed commits. For more information on how to create signed commits, please visit this page: https://docs.github.com/en/authentication/managing-commit-signature-verification/about-commit-signature-verification. Please use the button towards the bottom of the page to close this pull request and open a new one with signed commits."
	comments, _, err := client.Issues.ListComments(ctx, owner, repo, prNumber, nil)
	if err != nil {
		return fmt.Errorf("error fetching comments: %v", err)
	}

	// Check if the comment already exists
	for _, comment := range comments {
		if comment.GetUser().GetLogin() == commentAuthor && strings.EqualFold(comment.GetBody(), commentBody) {
			slogs.Logr.Info("Unsigned commit comment already exists", "repo", repo, "PR", prNumber)
			return nil
		}
	}

	// Post the comment if it doesn't exist
	comment := &github.IssueComment{
		Body: github.String(commentBody),
	}
	//
	_, _, err = client.Issues.CreateComment(ctx, owner, repo, prNumber, comment)
	if err != nil {
		return fmt.Errorf("error creating comment: %v", err)
	}

	return nil
}
