// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"github.com/opensourceways/robot-framework-lib/client"
	"github.com/opensourceways/robot-framework-lib/utils"
	"regexp"
	"strings"
)

const (
	// placeholderCommenter is a placeholder string for the commenter's name
	placeholderCommenter = "__commenter__"
	// placeholderAction is a placeholder string for the action
	placeholderAction = "__action__"
)

var (
	// regexpReopenComment is a compiled regular expression for reopening comments
	regexpReopenComment = regexp.MustCompile(`^/reopen$`)
	// regexpCloseComment is a compiled regular expression for closing comments
	regexpCloseComment = regexp.MustCompile(`^/close$`)
)

// handleReopenEvent only handles the reopening of an issue event.
// Handle completed, set the interrupt flag to interrupt the subsequent operations.
func (bot *robot) handleReopenEvent(evt *client.GenericEvent, org, repo, number string) (interrupt bool) {
	comment, state, commentKind := utils.GetString(evt.Comment), utils.GetString(evt.State), utils.GetString(evt.CommentKind)
	commenter, author := utils.GetString(evt.Commenter), utils.GetString(evt.Author)
	// If the comment is on an issue and the comment matches the reopen comment and the state is closed
	if commentKind == client.CommentOnIssue &&
		regexpReopenComment.MatchString(strings.TrimSpace(comment)) && state == bot.cnf.EventStateClosed {
		interrupt = true
		// Check if the commenter has the permission to operate
		if !bot.checkCommenterPermission(org, repo, number, author, commenter, commentKind, "reopen") {
			return
		}

		bot.cli.UpdateIssue(org, repo, number, bot.cnf.EventStateOpened)
	}
	return
}

// handleCloseEvent  handles the closing of an issue or pull request event
func (bot *robot) handleCloseEvent(evt *client.GenericEvent, configmap *repoConfig, org, repo, number string) {
	comment, state, commentKind := utils.GetString(evt.Comment), utils.GetString(evt.State), utils.GetString(evt.CommentKind)
	commenter, author := utils.GetString(evt.Commenter), utils.GetString(evt.Author)
	// If the comment matches the close comment and the state is opened
	if regexpCloseComment.MatchString(strings.TrimSpace(comment)) && state == bot.cnf.EventStateOpened {
		// Check if the commenter has the permission to operate
		if !bot.checkCommenterPermission(org, repo, number, author, commenter, commentKind, "close") {
			return
		}

		// If the comment kind is an pull request, update the pull request state to closed and return
		if commentKind != client.CommentOnIssue {
			bot.cli.UpdatePR(org, repo, number, bot.cnf.EventStateClosed)
			return
		}

		// Check if the issue needs linking to a pull request, and update the issue state to closed
		bot.checkIssueNeedLinkingPR(configmap, org, repo, number, commenter)
	}
}

// handleCloseEvent  handles the closing of an issue
func (bot *robot) checkIssueNeedLinkingPR(configmap *repoConfig, org, repo, number, commenter string) {
	if configmap.NeedIssueHasLinkPullRequests {
		// issue can be closed only when its linking PR exists
		num, success := bot.cli.GetIssueLinkedPRNumber(org, repo, number)
		// If the request is failed that means not be sure to close issue,
		// create a comment indicating do closing again and return
		if !success {
			bot.cli.CreateIssueComment(org, repo, number,
				strings.ReplaceAll(bot.cnf.CommentListLinkingPullRequestsFailure, placeholderCommenter, commenter))
			return
		}

		// If the linked pull request number is zero,
		// create a comment indicating that the issue needs a linked pull request and return
		if num == 0 {
			bot.cli.CreateIssueComment(org, repo, number,
				strings.ReplaceAll(bot.cnf.CommentIssueNeedsLinkPR, placeholderCommenter, commenter))
			return
		}
	}

	bot.cli.UpdateIssue(org, repo, number, bot.cnf.EventStateClosed)
}

func (bot *robot) checkCommenterPermission(org, repo, number, author, commenter, commentKind, action string) (pass bool) {
	if author == commenter {
		return true
	}
	pass, success := bot.cli.CheckPermission(org, repo, commenter)
	if success && !pass {
		bot.handleNoPermissionOperateIssueOrPR(org, repo, number, commenter, commentKind, action)
	}
	return pass && success
}

func (bot *robot) handleNoPermissionOperateIssueOrPR(org, repo, number, commenter, commentKind, action string) {
	if commentKind == client.CommentOnIssue {
		bot.cli.CreateIssueComment(org, repo, number,
			strings.ReplaceAll(strings.ReplaceAll(bot.cnf.CommentNoPermissionOperateIssue,
				placeholderCommenter, commenter), placeholderAction, action))
	} else {
		bot.cli.CreatePRComment(org, repo, number,
			strings.ReplaceAll(strings.ReplaceAll(bot.cnf.CommentNoPermissionOperatePR,
				placeholderCommenter, commenter), placeholderAction, action))
	}
}
