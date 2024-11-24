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
	"github.com/opensourceways/server-common-lib/config"
	"os"
	"path/filepath"
	"testing"
)

func TestRepoConfigValidate(t *testing.T) {
	// 测试空的Repos字段
	rc := &repoConfig{}
	err := rc.validate()
	if err == nil {
		t.Error("Expected an error for empty Repos field, but got nil")
	}

	// 测试非空的Repos字段
	rc.Repos = []string{"repo1", "repo2"}
	err = rc.validate()
	if err != nil {
		t.Errorf("Expected no error for non-empty Repos field, but got %v", err)
	}
}

func TestConfigurationValidate(t *testing.T) {
	// 测试空的配置
	c := &configuration{}
	err := c.Validate()
	if err == nil {
		t.Error("Expected an error for nil configuration, but got nil")
	}

	// 测试非空的配置
	c.ConfigItems = []repoConfig{
		{
			RepoFilter: config.RepoFilter{
				Repos: []string{"repo1", "repo2"},
			},
		},
	}
	err = c.Validate()
	if err != nil {
		t.Errorf("Expected no error for non-nil configuration, but got %v", err)
	}
}

func TestConfigurationGet(t *testing.T) {
	// 测试空的配置
	c := &configuration{}
	rc := c.get("org1", "repo1")
	if rc != nil {
		t.Error("Expected nil repoConfig for empty configuration, but got non-nil")
	}

	// 测试非空的配置，但不匹配的组织和仓库
	c.ConfigItems = []repoConfig{
		{
			RepoFilter: config.RepoFilter{
				Repos: []string{"repo1", "repo2"},
			},
		},
	}
	rc = c.get("org2", "repo3")
	if rc != nil {
		t.Error("Expected nil repoConfig for non-matching organization and repository, but got non-nil")
	}

	// 测试非空的配置，匹配的组织和仓库
	rc = c.get("org1", "repo1")
	if rc == nil {
		t.Error("Expected non-nil repoConfig for matching organization and repository, but got nil")
	}
}

func TestConfigurationNeedLinkPullRequests(t *testing.T) {
	// 测试空的配置
	c := &configuration{}
	needLink := c.NeedLinkPullRequests("org1", "repo1")
	if needLink {
		t.Error("Expected false for needing link to pull request for empty configuration, but got true")
	}

	// 测试非空的配置，但不匹配的组织和仓库
	c.ConfigItems = []repoConfig{
		{
			RepoFilter: config.RepoFilter{
				Repos: []string{"repo1", "repo2"},
			},
		},
	}
	needLink = c.NeedLinkPullRequests("org2", "repo3")
	if needLink {
		t.Error("Expected false for needing link to pull request for non-matching organization and repository, but got true")
	}

	// 测试非空的配置，匹配的组织和仓库，但NeedIssueHasLinkPullRequests为false
	c.ConfigItems[0].NeedIssueHasLinkPullRequests = false
	needLink = c.NeedLinkPullRequests("org1", "repo1")
	if needLink {
		t.Error("Expected false for needing link to pull request when NeedIssueHasLinkPullRequests is false, but got true")
	}

	// 测试非空的配置，匹配的组织和仓库，NeedIssueHasLinkPullRequests为true
	c.ConfigItems[0].NeedIssueHasLinkPullRequests = true
	needLink = c.NeedLinkPullRequests("org1", "repo1")
	if !needLink {
		t.Error("Expected true for needing link to pull request when NeedIssueHasLinkPullRequests is true, but got false")
	}
}

//
//func TestValidate(t *testing.T) {
//
//	type args struct {
//		cnf  *configuration
//		path string
//	}
//
//	testCases := []struct {
//		no  string
//		in  args
//		out []error
//	}{
//		{
//			"case0",
//			args{
//				&configuration{},
//				"",
//			},
//			[]error{nil, nil},
//		},
//		{
//			"case1",
//			args{
//				&configuration{
//					ConfigItems: make([]botConfig, 0),
//				},
//				"",
//			},
//			[]error{nil, nil},
//		},
//		{
//			no: "case2",
//			in: args{
//				&configuration{
//					ConfigItems: []botConfig{
//						{
//							config.RepoFilter{
//								[]string{},
//								[]string{},
//							},
//							"123132",
//							"fasdadads",
//							"",
//						},
//					},
//				},
//				"",
//			},
//			out: []error{nil, errors.New("the repositories configuration can not be empty")},
//		},
//		{
//			"case3",
//			args{
//				&configuration{},
//				"config1.yaml",
//			},
//			[]error{nil, errors.New("the repositories configuration can not be empty")},
//		},
//		{
//			"case4",
//			args{
//				&configuration{},
//				"config2.yaml",
//			},
//			[]error{nil, errors.New("the community_name configuration can not be empty")},
//		},
//		{
//			"case5",
//			args{
//				&configuration{},
//				"config3.yaml",
//			},
//			[]error{nil, errors.New("the command_link configuration can not be empty")},
//		},
//		{
//			"case6",
//			args{
//				nil,
//				"",
//			},
//			[]error{nil, nil},
//		},
//		{
//			"case7",
//			args{
//				&configuration{},
//				"config.yaml",
//			},
//			[]error{nil, nil},
//		},
//	}
//	for i := range testCases {
//		t.Run(testCases[i].no, func(t *testing.T) {
//			if testCases[i].in.path != "" {
//				err := utils.LoadFromYaml(findTestdata(t, "testdata"+string(os.PathSeparator)+testCases[i].in.path), testCases[i].in.cnf)
//				assert.Equal(t, testCases[i].out[0], err)
//			}
//
//			err1 := testCases[i].in.cnf.Validate()
//			assert.Equal(t, testCases[i].out[1], err1)
//		})
//	}
//
//}
//
//func TestGetConfig(t *testing.T) {
//	cnf := &configuration{}
//
//	got := cnf.get("owner1", "")
//	assert.Equal(t, (*botConfig)(nil), got)
//	_ = utils.LoadFromYaml(findTestdata(t, "testdata"+string(os.PathSeparator)+"config.yaml"), cnf)
//
//	got = cnf.get("owner1", "")
//	assert.Equal(t, "openUBMC1", got.CommunityName)
//	assert.Equal(t, "fafsadaf", got.CommandLink)
//
//	got = cnf.get("owner2", "repo1")
//	assert.Equal(t, "openUBMC1", got.CommunityName)
//	assert.Equal(t, "fafsadaf", got.CommandLink)
//
//	got = cnf.get("owner3", "repo1")
//	assert.Equal(t, "openUBMC2", got.CommunityName)
//	assert.Equal(t, "fafsadaf13", got.CommandLink)
//
//	got = cnf.get("owner4", "repo2")
//	assert.Equal(t, "openUBMC2", got.CommunityName)
//	assert.Equal(t, "fafsadaf13", got.CommandLink)
//
//	got = cnf.get("owner5", "repo2")
//	assert.Equal(t, (*botConfig)(nil), got)
//}

func findTestdata(t *testing.T, path string) string {

	i := 0
retry:
	absPath, err := filepath.Abs(path)
	if err != nil {
		t.Error(path + " not found")
		return ""
	}
	if _, err = os.Stat(absPath); !os.IsNotExist(err) {
		return absPath
	} else {
		i++
		path = ".." + string(os.PathSeparator) + path
		if i <= 3 {
			goto retry
		}
	}

	t.Log(path + " not found")
	return ""
}
