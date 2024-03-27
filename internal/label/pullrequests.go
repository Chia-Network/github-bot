package label

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/go-github/v60/github"
)

// PullRequests applies internal or community labels to pull requests
// Internal is determined by checking if the PR author is a member of the specified internalTeam
func PullRequests(githubClient *github.Client, internalTeam string, repos []string) error {
	teamMembers := map[string]bool{}

	teamParts := strings.Split(internalTeam, "/")
	if len(teamParts) != 2 {
		return fmt.Errorf("invalid team name - must contain org and team : %s", internalTeam)
	}

	teamOpts := &github.TeamListTeamMembersOptions{
		Role: "all",
		ListOptions: github.ListOptions{
			Page:    0,
			PerPage: 100,
		},
	}

	for {
		teamOpts.ListOptions.Page++
		members, resp, err := githubClient.Teams.ListTeamMembersBySlug(context.TODO(), teamParts[0], teamParts[1], teamOpts)
		if err != nil {
			return fmt.Errorf("error getting team %s: %w", internalTeam, err)
		}

		for _, member := range members {
			teamMembers[*member.Login] = true
		}

		if resp.NextPage == 0 {
			break
		}
	}

	for _, repo := range repos {
		log.Println("checking repos")
		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid repository name - must contain owner and repository: %s", repo)
		}
		opts := &github.PullRequestListOptions{
			State:     "open",
			Sort:      "created",
			Direction: "desc",
			ListOptions: github.ListOptions{
				Page:    0,
				PerPage: 100,
			},
		}
		for {
			opts.ListOptions.Page++
			pullRequests, resp, err := githubClient.PullRequests.List(context.TODO(), parts[0], parts[1], opts)
			if err != nil {
				return fmt.Errorf("error listing pull requests: %w", err)
			}

			for _, pullRequest := range pullRequests {
				if *pullRequest.Draft {
					continue
				}
				user := *pullRequest.User.Login
				var label string
				if teamMembers[user] {
					label = "internal"
				} else {
					label = "community"
				}

				log.Printf("Pull Request %d by %s will be labelled %s\n", *pullRequest.Number, user, label)
			}

			if resp.NextPage == 0 {
				break
			}
		}
	}

	return nil
}
