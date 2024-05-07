package github

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v60/github" // Ensure your go-github library version matches

	"github.com/chia-network/github-bot/internal/config"
)

// CheckForPendingCI returns a list of PR URLs that are ready for CI to run but haven't started yet.
func CheckForPendingCI(githubClient *github.Client, internalTeam string, cfg config.CheckStalePending) ([]string, error) {
	teamMembers, _ := GetTeamMemberList(githubClient, internalTeam)
	var pendingPRs []string

	for _, fullRepo := range cfg.CheckStalePending {
		log.Println("Checking repository:", fullRepo.Name)
		parts := strings.Split(fullRepo.Name, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid repository name - must contain owner and repository: %s", fullRepo.Name)
		}
		owner, repo := parts[0], parts[1]

		// Fetch community PRs using the FindCommunityPRs function
		communityPRs, err := FindCommunityPRs(cfg.CheckStalePending, teamMembers, githubClient)
		if err != nil {
			return nil, err
		}

		for _, pr := range communityPRs {
			// Dynamic cutoff time based on the last commit to the PR
			lastCommitTime, err := getLastCommitTime(githubClient, owner, repo, pr.GetNumber())
			if err != nil {
				log.Printf("Error retrieving last commit time for PR #%d: %v", pr.GetNumber(), err)
				continue
			}
			cutoffTime := lastCommitTime.Add(2 * time.Hour) // 2 hours after the last commit

			if time.Now().Before(cutoffTime) {
				log.Printf("Skipping PR #%d as it's still within the 2-hour window from the last commit.", pr.GetNumber())
				continue
			}

			hasCIRuns, err := checkCIStatus(githubClient, owner, repo, pr.GetNumber())
			if err != nil {
				log.Printf("Error checking CI status for PR #%d: %v", pr.GetNumber(), err)
				continue
			}

			needsReApproval, err := checkForDismissedReviews(githubClient, owner, repo, pr.GetNumber())
			if err != nil {
				log.Printf("Error checking review status for PR #%d: %v", pr.GetNumber(), err)
				continue
			}

			if !hasCIRuns || needsReApproval {
				log.Printf("PR #%d by %s is ready for CI since %v but no CI actions have started yet, or it requires re-approval.", pr.GetNumber(), pr.User.GetLogin(), pr.CreatedAt)
				pendingPRs = append(pendingPRs, pr.GetHTMLURL())
			}
		}
	}
	return pendingPRs, nil
}

func getLastCommitTime(client *github.Client, owner, repo string, prNumber int) (time.Time, error) {
	commits, _, err := client.PullRequests.ListCommits(context.Background(), owner, repo, prNumber, nil)
	if err != nil {
		return time.Time{}, err // Properly handle API errors
	}
	if len(commits) == 0 {
		return time.Time{}, fmt.Errorf("no commits found for PR #%d", prNumber) // Handle case where no commits are found
	}
	lastCommit := commits[len(commits)-1]
	commitDate := lastCommit.GetCommit().GetAuthor().GetDate() // commitDate is of type Timestamp

	// Since GetDate() returns a Timestamp (not *Timestamp), use the address to call GetTime()
	commitTime := commitDate.GetTime() // Correctly accessing GetTime(), which returns *time.Time

	if commitTime == nil {
		return time.Time{}, fmt.Errorf("commit time is nil for PR #%d", prNumber)
	}
	return *commitTime, nil // Safely dereference *time.Time to get time.Time
}

func checkCIStatus(client *github.Client, owner, repo string, prNumber int) (bool, error) {
	checks, _, err := client.Checks.ListCheckRunsForRef(context.Background(), owner, repo, strconv.Itoa(prNumber), &github.ListCheckRunsOptions{})
	if err != nil {
		return false, err
	}
	return checks.GetTotal() > 0, nil
}

func checkForDismissedReviews(client *github.Client, owner, repo string, prNumber int) (bool, error) {
	reviews, _, err := client.PullRequests.ListReviews(context.Background(), owner, repo, prNumber, nil)
	if err != nil {
		return false, err
	}
	for _, review := range reviews {
		if review.GetState() == "DISMISSED" {
			return true, nil
		}
	}
	return false, err
}
