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
	"github.com/opensourceways/server-common-lib/config"
)

type botConfig struct {
	config.RepoFilter
	// true: issue can be closed only when its linking PR exists
	// false: issue can be directly closed
	NeedIssueHasLinkPullRequests bool `json:"need_issue_has_link_pull_requests,omitempty"`
}

func (c *botConfig) validate() error {
	if len(c.Repos) == 0 {
		return errors.New("the repositories configuration can not be empty")
	}

	return c.RepoFilter.Validate()
}

type configuration struct {
	ConfigItems []botConfig `json:"config_items,omitempty"`
	// sig information url
	SigInfoURL string `json:"sig_info_url" required:"true"`
	// use as request parameter for get sig information
	CommunityName                         string `json:"community_name" required:"true"`
	EventStateOpened                      string `json:"event_state_opened" required:"true"`
	EventStateClosed                      string `json:"event_state_closed" required:"true"`
	CommentNoPermissionOperateIssue       string `json:"comment_no_permission_operate_issue"  required:"true"`
	CommentIssueNeedsLinkPR               string `json:"comment_issue_needs_link_pr"  required:"true"`
	CommentListLinkingPullRequestsFailure string `json:"comment_list_linking_pull_requests_failure"  required:"true"`
	CommentNoPermissionOperatePR          string `json:"comment_no_permission_operate_pr"  required:"true"`
}

func (c *configuration) Validate() error {
	if c == nil {
		return nil
	}

	items := c.ConfigItems
	for i := range items {
		if err := items[i].validate(); err != nil {
			return err
		}
	}

	return nil
}

func (c *configuration) get(org, repo string) *botConfig {
	if c == nil || len(c.ConfigItems) == 0 {
		return nil
	}

	for i := range c.ConfigItems {
		ok, _ := c.ConfigItems[i].RepoFilter.CanApply(org, org+"/"+repo)
		if ok {
			return &c.ConfigItems[i]
		}
	}

	return nil
}

func (c *configuration) NeedLinkPullRequests(org, repo string) bool {
	cnf := c.get(org, repo)
	if cnf != nil {
		return cnf.NeedIssueHasLinkPullRequests
	}

	return false
}
