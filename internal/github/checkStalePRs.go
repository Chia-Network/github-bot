package github

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/go-github/v60/github"

	"github.com/chia-network/github-bot/internal/config"
)

func CheckStalePRs(githubClient *github.Client, internalTeam string, cfg config.LabelConfig) ([]*github.PullRequest, error) {
	var stalePRs []*github.PullRequest
	cutoffDate := time.Now().AddDate(0, 0, -7) // 7 days ago

	for _, fullRepo := range cfg.LabelCheckRepos {
		log.Println("Checking repository:", fullRepo.Name)
		parts := strings.Split(fullRepo.Name, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid repository name - must contain owner and repository: %s", fullRepo.Name)
		}
		owner, repo := parts[0], parts[1]
		teamMembers, err := GetTeamMemberList(githubClient, internalTeam)
		if err != nil {
			return nil, err
		}

		communityPRs, err := FindCommunityPRs(owner, repo, teamMembers, githubClient)
		if err != nil {
			return nil, err
		}

		for _, pr := range communityPRs {
			if pr.UpdatedAt.Before(cutoffDate) {
				stalePRs = append(stalePRs, pr)
			}
		}
	}
	return stalePRs, nil
}

// Take the list of PRs and send a message to a keybase channel
