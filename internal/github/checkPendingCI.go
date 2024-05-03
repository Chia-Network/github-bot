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

// CheckForPendingCI will return a list of PR URLs that are ready for CI to run but haven't started yet.
func CheckForPendingCI(githubClient *github.Client, internalTeam string, cfg config.LabelConfig) ([]string, error) {
	teamMembers, _ := GetTeamMemberList(githubClient, internalTeam)
	var pendingPRs []string
	for _, fullRepo := range cfg.LabelCheckRepos {
		log.Println("Checking repository:", fullRepo.Name)
		parts := strings.Split(fullRepo.Name, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid repository name - must contain owner and repository: %s", fullRepo.Name)
		}
		owner := parts[0]
		repo := parts[1]

		// Fetch community PRs using the FindCommunityPRs function
		communityPRs, err := FindCommunityPRs(owner, repo, teamMembers, githubClient)
		if err != nil {
			return nil, err
		}

		// Now check if they've been updated in the last 2 hours and are awaiting CI actions
		cutoffTime := time.Now().Add(-2 * time.Hour) // 2 hours ago

		for _, pr := range communityPRs {
			if pr.CreatedAt.After(cutoffTime) {
				// Check if any CI runs have occurred using the new function
				hasCIRuns, err := checkCIStatus(githubClient, owner, repo, pr.GetNumber())
				if err != nil {
					log.Printf("Error checking CI status for PR #%d: %v", pr.GetNumber(), err)
					continue // Proceed to the next PR in case of error
				}

				// Only consider the PR pending if no CI runs have occurred yet
				if !hasCIRuns {
					log.Printf("PR #%d by %s is ready for CI since %v but no CI actions have started yet", pr.GetNumber(), pr.User.GetLogin(), pr.CreatedAt)
					pendingPRs = append(pendingPRs, pr.GetHTMLURL())
				}
			}
		}
	}
	return pendingPRs, nil
}

func checkCIStatus(client *github.Client, owner, repo string, prNumber int) (bool, error) {
	checks, _, err := client.Checks.ListCheckRunsForRef(context.Background(), owner, repo, strconv.Itoa(prNumber), &github.ListCheckRunsOptions{})
	if err != nil {
		return false, err
	}

	hasCIRuns := checks.GetTotal() > 0
	return hasCIRuns, nil
}
