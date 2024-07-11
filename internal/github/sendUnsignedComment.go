package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/google/go-github/v60/github"
)

// CheckAndComment checks for the specific comment from a specific account and posts a comment if it doesn't exist.
func CheckAndComment(ctx context.Context, client *github.Client, owner, repo string, prNumber int) error {
	commentAuthor := "ChiaAutomation"
	commentBody := "Your commits are not signed and our branch protection rules require signed commits. For more information on how to create signed commits, please visit this page: https://github.com/Chia-Network/chia-blockchain/blob/main/CONTRIBUTING.md#creating-signed-commits"
	comments, _, err := client.Issues.ListComments(ctx, owner, repo, prNumber, nil)
	if err != nil {
		return fmt.Errorf("error fetching comments: %v", err)
	}

	// Check if the comment already exists
	for _, comment := range comments {
		if comment.GetUser().GetLogin() == commentAuthor && strings.EqualFold(comment.GetBody(), commentBody) {
			slogs.Logr.Info("Comment already exists for", "repo", repo, "PR", prNumber)
			return nil
		}
	}

	// Post the comment if it doesn't exist
	comment := &github.IssueComment{
		Body: github.String(commentBody),
	}
	slogs.Logr.Info("Comment", "comment", comment)
	//_, _, err = client.Issues.CreateComment(ctx, owner, repo, prNumber, comment)
	//if err != nil {
	//	return fmt.Errorf("error creating comment: %v", err)
	//}

	return nil
}
