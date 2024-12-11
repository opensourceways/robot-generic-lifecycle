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
	"reflect"
	"strings"
)

// repoConfig is a configuration struct for a organization and repository.
// It includes a RepoFilter and a boolean value indicating if an issue can be closed only when its linking PR exists.
type repoConfig struct {
	// RepoFilter is used to filter repositories.
	config.RepoFilter
	// true: issue can be closed only when its linking PR exists
	// false: issue can be directly closed
	NeedIssueHasLinkPullRequests bool `json:"need_issue_has_link_pull_requests,omitempty"`
}

// validate to check the repoConfig data's validation, returns an error if invalid
func (c *repoConfig) validate() error {
	// If the bot is not configured to monitor any repositories, return an error.
	if len(c.Repos) == 0 {
		return errors.New("the repositories configuration can not be empty")
	}

	return c.RepoFilter.Validate()
}

// configuration holds a list of repoConfig configurations.
// It also  includes sig information url, community name, event states, comment templates.
type configuration struct {
	ConfigItems []repoConfig `json:"config_items,omitempty"`
	// Sig information url.
	SigInfoURL string `json:"sig_info_url" required:"true"`
	// Community name used as a request parameter to getRepoConfig sig information.
	CommunityName string `json:"community_name" required:"true"`
	// Event state for opened issues.
	EventStateOpened string `json:"event_state_opened" required:"true"`
	// Event state for closed issues.
	EventStateClosed string `json:"event_state_closed" required:"true"`
	// Comment template for when no permission to operate on an issue.
	CommentNoPermissionOperateIssue string `json:"comment_no_permission_operate_issue"  required:"true"`
	// Comment template indicating that an issue needs a linking PR.
	CommentIssueNeedsLinkPR string `json:"comment_issue_needs_link_pr"  required:"true"`
	// Comment template for listing linking pull requests that failed.
	CommentListLinkingPullRequestsFailure string `json:"comment_list_linking_pull_requests_failure"  required:"true"`
	// Comment template for when no permission to operate on a PR.
	CommentNoPermissionOperatePR string `json:"comment_no_permission_operate_pr"  required:"true"`
}

// Validate to check the configmap data's validation, returns an error if invalid
func (c *configuration) Validate() error {
	if c == nil {
		return errors.New("configuration is nil")
	}

	// Validate each repo configuration
	items := c.ConfigItems
	for i := range items {
		if err := items[i].validate(); err != nil {
			return err
		}
	}

	return c.validateGlobalConfig()
}

func (c *configuration) validateGlobalConfig() error {
	k := reflect.TypeOf(*c)
	v := reflect.ValueOf(*c)

	var missing []string
	n := k.NumField()
	for i := 0; i < n; i++ {
		tag := k.Field(i).Tag.Get("required")
		if len(tag) > 0 {
			s, _ := v.Field(i).Interface().(string)
			if s == "" {
				missing = append(missing, k.Field(i).Tag.Get("json"))
			}
		}
	}

	if len(missing) != 0 {
		return errors.New("missing the follow config: " + strings.Join(missing, ", "))
	}

	return nil
}

// getRepoConfig retrieves a repoConfig for a given organization and repository.
// Returns the repoConfig if found, otherwise returns nil.
func (c *configuration) getRepoConfig(org, repo string) *repoConfig {
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

// NeedLinkPullRequests checks if the link to the pull request is needed for a given organization and repository.
// Returns true if the link to the pull request is needed, false otherwise.
func (c *configuration) NeedLinkPullRequests(org, repo string) bool {
	cnf := c.getRepoConfig(org, repo)
	if cnf != nil {
		return cnf.NeedIssueHasLinkPullRequests
	}

	return false
}
