package github

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/go-github/v60/github"

	"github.com/chia-network/github-bot/internal/config"
)

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
			if pr.CreatedAt.After(cutoffTime) && !pr.GetMergeable() {
				log.Printf("PR #%d by %s is pending CI actions since %v", pr.GetNumber(), pr.User.GetLogin(), pr.CreatedAt)
				pendingPRs = append(pendingPRs, pr.GetHTMLURL())
			}
		}
	}
	return pendingPRs, nil
}
