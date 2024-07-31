package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/google/go-github/v60/github" // Ensure your go-github library version matches

	"github.com/chia-network/github-bot/internal/config"
)

// PendingPR holds information about pending PRs
type PendingPR struct {
	Repo     string
	PRNumber int
	URL      string
}

// CheckForPendingCI returns a list of PR URLs that are ready for CI to run but haven't started yet.
func CheckForPendingCI(ctx context.Context, githubClient *github.Client, cfg *config.Config) ([]PendingPR, error) {
	teamMembers, _ := GetTeamMemberList(githubClient, cfg.InternalTeam)
	var pendingPRs []PendingPR

	for _, fullRepo := range cfg.CheckRepos {
		slogs.Logr.Info("Checking repository", "repository", fullRepo.Name)
		parts := strings.Split(fullRepo.Name, "/")
		if len(parts) != 2 {
			slogs.Logr.Error("Invalid repository name - must contain owner and repository", "repository", fullRepo.Name)
			continue
		}
		owner, repo := parts[0], parts[1]

		// Fetch community PRs using the FindCommunityPRs function
		communityPRs, err := FindCommunityPRs(cfg, teamMembers, githubClient, owner, repo, fullRepo.MinimumNumber)
		if err != nil {
			return nil, err
		}

		for _, pr := range communityPRs {
			slogs.Logr.Info("Checking PR", "PR", pr.GetHTMLURL())
			prctx, prcancel := context.WithTimeout(ctx, 30*time.Second) // 30 seconds timeout for each request
			defer prcancel()
			// Dynamic cutoff time based on the last commit to the PR
			slogs.Logr.Info("Fetching last commit time", "PR", pr.GetHTMLURL())
			lastCommitTime, err := getLastCommitTime(prctx, githubClient, owner, repo, pr.GetNumber())
			if err != nil {
				slogs.Logr.Error("Error retrieving last commit time", "PR", pr.GetNumber(), "repository", fullRepo.Name, "error", err)
				continue
			}
			cutoffTime := lastCommitTime.Add(2 * time.Hour) // 2 hours after the last commit

			if time.Now().Before(cutoffTime) {
				slogs.Logr.Info("Skipping PR as it's still within the 2-hour window from the last commit", "PR", pr.GetNumber(), "repository", fullRepo.Name)
				continue
			}

			slogs.Logr.Info("Checking CI status for PR", "PR", pr.GetHTMLURL())
			pendingCI, err := checkCIStatus(prctx, githubClient, owner, repo, pr.GetNumber())
			if err != nil {
				slogs.Logr.Error("Error checking CI status", "PR", pr.GetNumber(), "repository", fullRepo.Name, "error", err)
				continue
			}

			slogs.Logr.Info("Checking team member activity for PR", "PR", pr.GetHTMLURL())
			teamMemberActivity, err := checkTeamMemberActivity(prctx, githubClient, owner, repo, pr.GetNumber(), teamMembers, lastCommitTime)
			if err != nil {
				slogs.Logr.Error("Error checking team member activity", "PR", pr.GetNumber(), "repository", fullRepo.Name, "error", err)
				continue // or handle the error as needed
			}

			slogs.Logr.Info("Evaluating PR", "PR", pr.GetHTMLURL(), "Action Required for CI", pendingCI, "teamMemberActivity", teamMemberActivity)
			if pendingCI && !teamMemberActivity {
				slogs.Logr.Info("PR is ready for CI checks approval", "PR", pr.GetNumber(), "repository", fullRepo.Name, "user", pr.User.GetLogin(), "created_at", pr.CreatedAt)
				pendingPRs = append(pendingPRs, PendingPR{
					Repo:     repo,
					PRNumber: pr.GetNumber(),
					URL:      pr.GetHTMLURL(),
				})
			} else {
				slogs.Logr.Info("PR is not ready for CI approvals",
					"PR", pr.GetNumber(),
					"repository", fullRepo.Name)
			}

		}
	}
	return pendingPRs, nil
}

func getLastCommitTime(ctx context.Context, client *github.Client, owner, repo string, prNumber int) (time.Time, error) {
	commits, _, err := client.PullRequests.ListCommits(ctx, owner, repo, prNumber, nil)
	if err != nil {
		return time.Time{}, err // Properly handle API errors
	}
	if len(commits) == 0 {
		return time.Time{}, fmt.Errorf("no commits found for PR #%d of repo %s", prNumber, repo) // Handle case where no commits are found
	}
	// Requesting a list of commits will return the json list in descending order
	lastCommit := commits[len(commits)-1]
	commitDate := lastCommit.GetCommit().GetAuthor().GetDate() // commitDate is of type Timestamp

	// Since GetDate() returns a Timestamp (not *Timestamp), use the address to call GetTime()
	commitTime := commitDate.GetTime() // Correctly accessing GetTime(), which returns *time.Time
	if commitTime == nil {
		return time.Time{}, fmt.Errorf("commit time is nil for PR #%d of repo %s", prNumber, repo)
	}
	slogs.Logr.Info("The last commit time", "time", commitTime.Format(time.RFC3339), "PR", prNumber, "repository", repo)

	return *commitTime, nil // Safely dereference *time.Time to get time.Time
}

func checkCIStatus(ctx context.Context, client *github.Client, owner, repo string, prNumber int) (bool, error) {
	pr, _, err := client.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		return false, fmt.Errorf("failed to fetch pull request #%d: %w", prNumber, err)
	}

	headSHA := pr.GetHead().GetSHA()

	opts := &github.ListWorkflowRunsOptions{
		Status:  "action_required",
		HeadSHA: headSHA,
	}
	workflowRuns, _, err := client.Actions.ListRepositoryWorkflowRuns(ctx, owner, repo, opts)
	if err != nil {
		return false, fmt.Errorf("failed to fetch workflow runs for repository %s/%s: %w", owner, repo, err)
	}

	// Check for any workflows waiting for approval
	for _, workflows := range workflowRuns.WorkflowRuns {
		// This will check to see if there are any workflows that need approval and also ensure that its the same commit SHA as the PR we care about
		if workflows.GetConclusion() == "action_required" && workflows.GetHeadSHA() == headSHA {
			slogs.Logr.Info("Workflow awaiting approval for", "PR", prNumber, "repository", repo)
			return true, nil // A workflow is awaiting approval
		}
	}

	return false, nil // No workflows are awaiting approval
}

func checkTeamMemberActivity(ctx context.Context, client *github.Client, owner, repo string, prNumber int, teamMembers map[string]bool, lastCommitTime time.Time) (bool, error) {
	comments, _, err := client.Issues.ListComments(ctx, owner, repo, prNumber, nil)
	if err != nil {
		return false, fmt.Errorf("failed to fetch comments: %w", err)
	}

	for _, comment := range comments {
		slogs.Logr.Info("Checking comment from user", "user", comment.User.GetLogin(), "created_at", comment.CreatedAt.Format(time.RFC3339), "PR", prNumber, "repository", repo)
		if _, ok := teamMembers[comment.User.GetLogin()]; ok && comment.CreatedAt.After(lastCommitTime) {
			slogs.Logr.Info("Found team member comment after last commit time", "time", comment.CreatedAt.Format(time.RFC3339), "PR", prNumber, "repository", repo)
			// Check if the comment is after the last commit
			return true, nil // Active and relevant participation
		}
	}

	reviews, _, err := client.PullRequests.ListReviews(context.Background(), owner, repo, prNumber, nil)
	if err != nil {
		return false, fmt.Errorf("failed to fetch reviews: %w for PR #%d of repo %s", err, prNumber, repo)
	}

	for _, review := range reviews {
		if _, ok := teamMembers[review.User.GetLogin()]; ok && review.SubmittedAt.After(lastCommitTime) {
			switch review.GetState() {
			case "DISMISSED", "CHANGES_REQUESTED", "COMMENTED":
				// Check if the review is after the last commit and is in one of the specified states
				return true, nil
			}
		}
	}

	return false, nil // No recent relevant activity from team members
}
