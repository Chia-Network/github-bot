package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v60/github"
)

// GetTeamMemberList obtains a list of teammembers
func GetTeamMemberList(githubClient *github.Client, internalTeam string, internalTeamIgnoredMembers []string) (map[string]bool, error) {
	teamMembers := make(map[string]bool)

	teamParts := strings.Split(internalTeam, "/")
	if len(teamParts) != 2 {
		return nil, fmt.Errorf("invalid team name - must contain org and team : %s", internalTeam)
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
			return nil, fmt.Errorf("error getting team %s: %w", internalTeam, err)
		}

		for _, member := range members {
			shouldIgnoreMember := false
			for _, ignoredMember := range internalTeamIgnoredMembers {
				if member.Login != nil {
					if strings.EqualFold(ignoredMember, *member.Login) {
						shouldIgnoreMember = true
						break
					}
				}
			}

			if member.Login != nil && !shouldIgnoreMember {
				teamMembers[*member.Login] = true
			}
		}

		if resp.NextPage == 0 {
			break
		}
	}
	return teamMembers, nil
}
