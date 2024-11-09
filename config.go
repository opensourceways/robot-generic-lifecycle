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
	"k8s.io/apimachinery/pkg/util/sets"
)

type botConfig struct {
	config.RepoFilter
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
	if c == nil {
		return false
	}
	orgRepo := org + "/" + repo
	items := c.ConfigItems
	for _, item := range items {
		repoFilter := item.RepoFilter
		v := sets.NewString(repoFilter.Repos...)
		needLinkPullRequests := item.NeedIssueHasLinkPullRequests
		if v.Has(orgRepo) {
			return needLinkPullRequests
		}
		if !v.Has(org) {
			return false
		}
		if len(repoFilter.ExcludedRepos) > 0 && sets.NewString(repoFilter.ExcludedRepos...).Has(orgRepo) {
			return false
		}
		return needLinkPullRequests
	}
	return false
}
