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
	"github.com/opensourceways/go-gitcode/openapi"
	"github.com/opensourceways/robot-framework-lib/client"
	"github.com/opensourceways/robot-framework-lib/config"
	"github.com/opensourceways/robot-framework-lib/framework"
	"github.com/sirupsen/logrus"
	"regexp"
)

type iClient interface {
	CreatePRComment(owner, repo, number, comment string) (result any, success bool, err error)
	CreateIssueComment(owner, repo, number, comment string) (result any, success bool, err error)
	GetRepoAllMember(org, repo string) (result []any, success bool, err error)
}

type robot struct {
	cli iClient
	cnf *configuration
}

func newRobot(c *configuration, token []byte) *robot {
	lgr := logrus.NewEntry(logrus.StandardLogger())
	return &robot{cli: client.NewClient(token, lgr), cnf: c}
}

func (bot *robot) NewConfig() config.Configmap {
	return &configuration{}
}

func (bot *robot) RegisterEventHandler(p framework.HandlerRegister) {
	p.RegisterIssueCommentHandler(bot.handleIssueCommentEvent)
	p.RegisterPullRequestCommentHandler(bot.handlePullRequestCommentEvent)
}

func checkEventInvalid(evt *client.GenericEvent) bool {
	if evt.Action != nil && *evt.Action != "open" {
		return true
	}

	if evt.State != nil && *evt.State != "opened" {
		return true
	}

	if evt.Org != nil && *evt.Org == "" {
		return true
	}

	if evt.Repo != nil && *evt.Repo == "" {
		return true
	}

	if evt.Author != nil && *evt.Author == "" {
		return true
	}

	return false
}

func (bot *robot) getConfig(cnf config.Configmap, org, repo string) (*botConfig, error) {
	c := cnf.(*configuration)
	if bc := c.get(org, repo); bc != nil {
		return bc, nil
	}

	return nil, errors.New("no config for this repo: " + org + "," + repo)
}

const (
	eventStateOpened  = "opened"
	eventStateClosed  = "closed"
	emptyErrorMessage = ""
)

var (
	regexpReopenComment = regexp.MustCompile(`(?mi)^/reopen\s*$`)
	regexpCloseComment  = regexp.MustCompile(`(?mi)^/close\s*$`)
)

func checkEvent(evt *client.GenericEvent) (bool, string) {
	checking(evt.Comment, "", false, "author is empty")
	if regexpReopenComment.MatchString(*evt.Comment) {
		checking(evt.State, eventStateClosed, false, "do reopen the event, but it's state is not "+eventStateClosed)
	}
	if regexpCloseComment.MatchString(*evt.Comment) {
		checking(evt.State, eventStateOpened, false, "event state is not "+eventStateOpened)
	}
	checking(evt.State, eventStateOpened, false, "event state is not "+eventStateOpened)
	checking(evt.Org, eventStateOpened, true, "org is empty")
	checking(evt.Repo, eventStateOpened, false, "repo is empty")
	checking(evt.Author, eventStateOpened, false, "author is empty")
	return true, ""
}

func checking(key *string, value string, equal bool, message string) string {
	if key == nil {
		return emptyErrorMessage
	}
	if equal && *key == value {
		return message
	}
	if *key != value {
		return message
	}
	return emptyErrorMessage
}

func (bot *robot) handleIssueCommentEvent(evt *client.GenericEvent, cnf config.Configmap, lgr *logrus.Entry) error {
	if checkEventInvalid(evt) {
		return errors.New("issue status is invalid")
	}

	cfg, err := bot.getConfig(cnf, *evt.Org, *evt.Repo)
	if err != nil {
		return err
	}

	comment, err := bot.genWelcomeMessage(*evt.Org, *evt.Repo, *evt.Author, cfg, lgr)
	if err != nil {
		return err
	}

	_, success, err := bot.cli.CreateIssueComment(*evt.Org, *evt.Repo, *evt.Number, comment)
	lgr.Infof("Issue[%s/%s/%s] create comment result: %v", *evt.Org, *evt.Repo, *evt.Number, success)
	return nil
}

func (bot *robot) handlePullRequestCommentEvent(evt *client.GenericEvent, cnf config.Configmap, lgr *logrus.Entry) error {
	if checkEventInvalid(evt) {
		return errors.New("PR status is invalid")
	}

	cfg, err := bot.getConfig(cnf, *evt.Org, *evt.Repo)
	if err != nil {
		return err
	}

	comment, err := bot.genWelcomeMessage(*evt.Org, *evt.Repo, *evt.Author, cfg, lgr)
	if err != nil {
		return err
	}

	_, success, err := bot.cli.CreatePRComment(*evt.Org, *evt.Repo, *evt.Number, comment)
	lgr.Infof("PR[%s/%s/%s] create comment result: %v", *evt.Org, *evt.Repo, *evt.Number, success)
	return err
}

func (bot *robot) genWelcomeMessage(org, repo, author string, cfg *botConfig, lgr *logrus.Entry) (string, error) {

	arr, success, err := bot.cli.GetRepoAllMember(org, repo)
	lgr.Infof("PR[%s/%s] get all member result: %v", org, repo, success)
	if !success || err != nil {
		return fmt.Sprintf(cfg.WelcomeMessage, author, cfg.CommunityName, cfg.CommunityName, cfg.CommunityName, cfg.CommandLink), err
	}

	maintainers := make([]string, 0, len(arr))
	// TODO 后面改为反射获取统一结构
	for i := range arr {
		u := arr[i].(openapi.User)
		if u.UserName != nil && u.Permissions != nil && u.Permissions.Admin != nil && *u.Permissions.Admin {
			maintainers = append(maintainers, *u.UserName)
		}
	}

	// TODO
	author = fmt.Sprintf("[@%s](https://gitcode.com/%s)", author, author)
	maintainersList := ""
	for i := 0; i < len(maintainers); i++ {
		maintainersList = maintainersList + fmt.Sprintf(", [@%s](https://gitcode.com/%s)", maintainers[i], maintainers[i])
	}
	if len(maintainers) > 2 {
		maintainersList = maintainersList[2:]
	}
	return fmt.Sprintf(cfg.WelcomeMessage, author, cfg.CommunityName, cfg.CommunityName,
		maintainersList, cfg.CommandLink), nil
}
