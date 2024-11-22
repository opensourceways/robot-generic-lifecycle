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
	"errors"
	"fmt"
	"github.com/opensourceways/robot-framework-lib/client"
	"github.com/opensourceways/robot-framework-lib/config"
	"github.com/opensourceways/robot-framework-lib/framework"
	"github.com/opensourceways/robot-framework-lib/utils"
	"github.com/sirupsen/logrus"
	"regexp"
)

type iClient interface {
	CreatePRComment(org, repo, number, comment string) (success bool)
	CreateIssueComment(org, repo, number, comment string) (success bool)
	CheckPermission(org, repo, username, permission string) (pass, success bool)
	UpdateIssue(org, repo, number, state string) (success bool)
	UpdatePR(org, repo, number, state string) (success bool)
	GetIssueLinkedPRNumber(org, repo, number string) (num int, success bool)
}

type robot struct {
	cli       iClient
	cnf       *configuration
	log       *logrus.Entry
	interrupt bool
}

func newRobot(c *configuration, token []byte) *robot {
	m := make(logrus.Fields, 2)
	m["component"] = component
	logger := framework.NewLogger(m)
	return &robot{cli: client.NewClient(token, logger), cnf: c, log: logger}
}

func (bot *robot) NewConfig() config.Configmap {
	return &configuration{}
}

func (bot *robot) RegisterEventHandler(p framework.HandlerRegister) {
	p.RegisterIssueCommentHandler(bot.handleCommentEvent)
	p.RegisterPullRequestCommentHandler(bot.handleCommentEvent)
}

func (bot *robot) GetLogger() *logrus.Entry {
	return bot.log
}

func (bot *robot) getConfig(cnf config.Configmap, org, repo string) (*botConfig, error) {
	c := cnf.(*configuration)
	if bc := c.get(org, repo); bc != nil {
		return bc, nil
	}

	return nil, errors.New("no config for this repo: " + org + "," + repo)
}

const (
	eventStateOpened                      = "opened"
	eventStateClosed                      = "closed"
	commentNoPermissionOperateIssue       = `***@%s*** you can't %s an issue unless you are the author of it or a collaborator.`
	commentIssueNeedsLinkPR               = `***@%s*** you can't close an issue unless the issue has link pull requests.`
	commentListLinkingPullRequestsFailure = `***@%s*** fail to check link pull requests of the issue, please retry.`
	commentNoPermissionOperatePR          = `***@%s*** you can't %s a pull request unless you are the author of it or a collaborator.`
)

var (
	regexpReopenComment = regexp.MustCompile(`(?mi)^/reopen\s*$`)
	regexpCloseComment  = regexp.MustCompile(`(?mi)^/close\s*$`)
)

func (bot *robot) handleCommentEvent(evt *client.GenericEvent, cnf config.Configmap) {
	org, repo, number := utils.GetString(evt.Org), utils.GetString(evt.Repo), utils.GetString(evt.Number)
	configmap, err := bot.getConfig(cnf, org, repo)
	if err != nil {
		bot.log.WithError(err).Error()
		return
	}

	bot.handleReopenEvent(evt, org, repo, number)
	if bot.interrupt {
		return
	}

	bot.handleCloseEvent(evt, configmap, org, repo, number)
}

func (bot *robot) handleReopenEvent(evt *client.GenericEvent, org, repo, number string) {
	comment, state := utils.GetString(evt.Comment), utils.GetString(evt.State)
	commenter, author := utils.GetString(evt.Commenter), utils.GetString(evt.Author)
	if utils.GetBool(evt.CommentOfPR) == "false" && regexpReopenComment.MatchString(comment) && state == eventStateClosed {
		bot.interrupt = true
		if !bot.checkCommenterPermission(org, repo, author, commenter, func() {
			bot.cli.CreateIssueComment(org, repo, number, fmt.Sprintf(commentNoPermissionOperateIssue, commenter, "reopen"))
		}) {
			return
		}

		bot.cli.UpdateIssue(org, repo, number, eventStateOpened)
	}
}

func (bot *robot) handleCloseEvent(evt *client.GenericEvent, configmap *botConfig, org, repo, number string) {
	comment, state := utils.GetString(evt.Comment), utils.GetString(evt.State)
	commenter, author := utils.GetString(evt.Commenter), utils.GetString(evt.Author)
	if regexpCloseComment.MatchString(comment) && state == eventStateOpened {
		if !bot.checkCommenterPermission(org, repo, author, commenter, func() {
			if utils.GetBool(evt.CommentOfPR) == "true" {
				bot.cli.CreatePRComment(org, repo, number, fmt.Sprintf(commentNoPermissionOperatePR, commenter, "close"))
			} else {
				bot.cli.CreateIssueComment(org, repo, number, fmt.Sprintf(commentNoPermissionOperateIssue, commenter, "close"))
			}
		}) {
			return
		}

		if utils.GetBool(evt.CommentOfPR) == "true" {
			bot.cli.UpdatePR(org, repo, number, eventStateClosed)
			return
		}

		bot.checkIssueNeedLinkingPR(configmap, org, repo, number, commenter)
	}
}

func (bot *robot) checkIssueNeedLinkingPR(configmap *botConfig, org, repo, number, commenter string) {
	if configmap.NeedIssueHasLinkPullRequests {
		// issue can be closed only when its linking PR exists
		num, success := bot.cli.GetIssueLinkedPRNumber(org, repo, number)
		bot.log.Infof("list the issue[%s/%s,%s] linking PR number is successful: %v, number: %d", org, repo, number, success, num)
		if !success {
			bot.cli.CreateIssueComment(org, repo, number, fmt.Sprintf(commentListLinkingPullRequestsFailure, commenter))
			return
		}

		if num == 0 {
			bot.cli.CreateIssueComment(org, repo, number, fmt.Sprintf(commentIssueNeedsLinkPR, commenter))
			return
		}
	}

	bot.cli.UpdateIssue(org, repo, number, eventStateClosed)
}

func (bot *robot) checkCommenterPermission(org, repo, author, commenter string, fn func()) (pass bool) {
	if author == commenter {
		return true
	}
	pass, success := bot.cli.CheckPermission(org, repo, commenter, "")
	bot.log.Infof("request success: %t, the %s has permission to the repo[%s/%s]: %t", success, commenter, org, repo, pass)

	if success && !pass {
		fn()
	}
	return pass && success
}
