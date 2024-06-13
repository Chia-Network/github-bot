package github

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/go-github/v60/github" // Ensure your go-github library version matches

	"github.com/chia-network/github-bot/internal/config"
)

// CheckForPendingCI returns a list of PR URLs that are ready for CI to run but haven't started yet.
func CheckForPendingCI(ctx context.Context, githubClient *github.Client, cfg *config.Config) ([]string, error) {
	teamMembers, _ := GetTeamMemberList(githubClient, cfg.InternalTeam)
	var pendingPRs []string

	for _, fullRepo := range cfg.CheckRepos {
		log.Println("Checking repository:", fullRepo.Name)
		parts := strings.Split(fullRepo.Name, "/")
		if len(parts) != 2 {
			log.Printf("invalid repository name - must contain owner and repository: %s", fullRepo.Name)
			continue
		}
		owner, repo := parts[0], parts[1]

		// Fetch community PRs using the FindCommunityPRs function
		communityPRs, err := FindCommunityPRs(cfg, teamMembers, githubClient, owner, repo, fullRepo.MinimumNumber)
		if err != nil {
			return nil, err
		}

		for _, pr := range communityPRs {
			prctx, prcancel := context.WithTimeout(ctx, 30*time.Second) // 30 seconds timeout for each request
			defer prcancel()
			// Dynamic cutoff time based on the last commit to the PR
			lastCommitTime, err := getLastCommitTime(prctx, githubClient, owner, repo, pr.GetNumber())
			if err != nil {
				log.Printf("Error retrieving last commit time for PR #%d in %s/%s: %v", pr.GetNumber(), owner, repo, err)
				continue
			}
			cutoffTime := lastCommitTime.Add(2 * time.Hour) // 2 hours after the last commit

			if time.Now().Before(cutoffTime) {
				log.Printf("Skipping PR #%d from %s/%s repo as it's still within the 2-hour window from the last commit.", pr.GetNumber(), owner, repo)
				continue
			}

			hasCIRuns, err := checkCIStatus(prctx, githubClient, owner, repo, pr.GetNumber())
			if err != nil {
				log.Printf("Error checking CI status for PR #%d in %s/%s: %v", pr.GetNumber(), owner, repo, err)
				continue
			}

			teamMemberActivity, err := checkTeamMemberActivity(prctx, githubClient, owner, repo, pr.GetNumber(), teamMembers, lastCommitTime)
			if err != nil {
				log.Printf("Error checking team member activity for PR #%d in %s/%s: %v", pr.GetNumber(), owner, repo, err)
				continue // or handle the error as needed
			}
			if !hasCIRuns && !teamMemberActivity {
				log.Printf("PR #%d in %s/%s by %s is ready for CI since %v but no CI actions have started yet, or it requires re-approval.", pr.GetNumber(), owner, repo, pr.User.GetLogin(), pr.CreatedAt)
				pendingPRs = append(pendingPRs, pr.GetHTMLURL())
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
	log.Printf("The last commit time is %s for PR #%d of repo %s", commitTime.Format(time.RFC3339), prNumber, repo)

	return *commitTime, nil // Safely dereference *time.Time to get time.Time
}

func checkCIStatus(ctx context.Context, client *github.Client, owner, repo string, prNumber int) (bool, error) {
	pr, _, err := client.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		return false, fmt.Errorf("failed to fetch pull request #%d: %w", prNumber, err)
	}
	// Obtaining CI runs for the most recent commit
	commitSHA := pr.GetHead().GetSHA()

	checks, _, err := client.Checks.ListCheckRunsForRef(ctx, owner, repo, commitSHA, &github.ListCheckRunsOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to fetch check runs for ref %s: %w", commitSHA, err)
	}
	// Potentially add logic later for checking for passing CI
	/*
		for _, checkRun := range checks.CheckRuns {
		if checkRun.GetStatus() != "completed" || checkRun.GetConclusion() != "success" {
			return false, nil // There are check runs that are not completed or not successful
		}
	*/
	return checks.GetTotal() > 0, nil
}

func checkTeamMemberActivity(ctx context.Context, client *github.Client, owner, repo string, prNumber int, teamMembers map[string]bool, lastCommitTime time.Time) (bool, error) {
	comments, _, err := client.Issues.ListComments(ctx, owner, repo, prNumber, nil)
	if err != nil {
		return false, fmt.Errorf("failed to fetch comments: %w", err)
	}

	for _, comment := range comments {
		log.Printf("Checking comment by %s at %s for PR #%d of repo %s", comment.User.GetLogin(), comment.CreatedAt.Format(time.RFC3339), prNumber, repo)
		if _, ok := teamMembers[comment.User.GetLogin()]; ok && comment.CreatedAt.After(lastCommitTime) {
			log.Printf("Found team member comment after last commit time: %s for PR #%d of repo %s", comment.CreatedAt.Format(time.RFC3339), prNumber, repo)
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
