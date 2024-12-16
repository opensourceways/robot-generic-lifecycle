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
	"encoding/json"
	"github.com/opensourceways/robot-framework-lib/client"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"
)

type mockClient struct {
	mock.Mock
	successfulCreatePRComment        bool
	successfulUpdatePR               bool
	successfulUpdateIssue            bool
	successfulCreateIssueComment     bool
	successfulGetIssueLinkedPRNumber bool
	successfulCheckPermission        bool
	permission                       bool
	method                           string
	issueLinkingPRNum                int
}

func (m *mockClient) CreatePRComment(org, repo, number, comment string) bool {
	m.method = "CreatePRComment"
	return m.successfulCreatePRComment
}

func (m *mockClient) CreateIssueComment(org, repo, number, comment string) bool {
	m.method = "CreateIssueComment"
	return m.successfulCreateIssueComment
}

func (m *mockClient) UpdateIssue(org, repo, number, state string) bool {
	m.method = "UpdateIssue"
	return m.successfulUpdateIssue
}

func (m *mockClient) UpdatePR(org, repo, number, state string) bool {
	m.method = "UpdatePR"
	return m.successfulUpdatePR
}

func (m *mockClient) GetIssueLinkedPRNumber(org, repo, number string) (int, bool) {
	m.method = "GetIssueLinkedPRNumber"
	return m.issueLinkingPRNum, m.successfulGetIssueLinkedPRNumber
}

func (m *mockClient) CheckPermission(org, repo, username string) (bool, bool) {
	m.method = "CheckPermission"
	return m.permission, m.successfulCheckPermission
}

const (
	org       = "org1"
	repo      = "repo1"
	number    = "1"
	commenter = "commenter1"
	comment   = "/close"
	comment1  = "/reopen"
)

func TestHandleReopenEvent(t *testing.T) {

	mc := new(mockClient)
	bot := &robot{cli: mc, cnf: &configuration{
		EventStateClosed: "closed",
	}}

	cli, ok := bot.cli.(*mockClient)
	assert.Equal(t, true, ok)

	event := new(client.GenericEvent)
	data, _ := os.ReadFile(findTestdata(t, "note_event.json"))
	err := json.Unmarshal(data, event)
	assert.Equal(t, nil, err)

	case1 := "the comment is not matching anyone comment command"
	cli.method = case1
	bot.handleReopenEvent(event, org, repo, number)
	execMethod1 := cli.method
	assert.Equal(t, case1, execMethod1)

	*event.Comment = comment1
	case2 := "CheckPermission"
	*event.CommentKind = client.CommentOnIssue
	*event.State = bot.cnf.EventStateClosed
	cli.method = ""
	bot.handleReopenEvent(event, org, repo, number)
	execMethod2 := cli.method
	assert.Equal(t, case2, execMethod2)

	case3 := "UpdateIssue"
	cli.method = ""
	author := commenter
	event.Author = &author
	*event.Commenter = commenter
	bot.handleReopenEvent(event, org, repo, number)
	execMethod3 := cli.method
	assert.Equal(t, case3, execMethod3)

}

func TestHandleCloseEvent(t *testing.T) {

	mc := new(mockClient)
	bot := &robot{cli: mc, cnf: &configuration{
		CommentNoPermissionOperateIssue: " [@__commenter__](***/__commenter__)  you ",
		EventStateOpened:                "opened",
	}}

	cli, ok := bot.cli.(*mockClient)
	assert.Equal(t, true, ok)

	event := new(client.GenericEvent)
	data, _ := os.ReadFile(findTestdata(t, "note_event.json"))
	err := json.Unmarshal(data, event)
	assert.Equal(t, nil, err)

	repoCnf := &repoConfig{}

	case1 := "the comment is not matching anyone comment command"
	cli.method = case1
	bot.handleCloseEvent(event, repoCnf, org, repo, number)
	execMethod1 := cli.method
	assert.Equal(t, case1, execMethod1)

	*event.Comment = comment
	case2 := "CheckPermission"
	cli.method = ""
	bot.handleCloseEvent(event, repoCnf, org, repo, number)
	execMethod2 := cli.method
	assert.Equal(t, case2, execMethod2)

	case3 := "UpdatePR"
	cli.method = ""
	author := commenter
	event.Author = &author
	*event.Commenter = commenter
	bot.handleCloseEvent(event, repoCnf, org, repo, number)
	execMethod3 := cli.method
	assert.Equal(t, case3, execMethod3)

	case4 := "UpdateIssue"
	cli.method = ""
	*event.CommentKind = client.CommentOnIssue
	bot.handleCloseEvent(event, repoCnf, org, repo, number)
	execMethod4 := cli.method
	assert.Equal(t, case4, execMethod4)

	case5 := "CreateIssueComment"
	cli.method = ""
	repoCnf.NeedIssueHasLinkPullRequests = true
	bot.handleCloseEvent(event, repoCnf, org, repo, number)
	execMethod5 := cli.method
	assert.Equal(t, case5, execMethod5)

	case6 := "CreateIssueComment"
	cli.method = ""
	cli.successfulGetIssueLinkedPRNumber = true
	bot.handleCloseEvent(event, repoCnf, org, repo, number)
	execMethod6 := cli.method
	assert.Equal(t, case6, execMethod6)
}

func TestCheckCommenterPermission(t *testing.T) {

	mc := new(mockClient)
	bot := &robot{cli: mc, cnf: &configuration{
		CommentNoPermissionOperateIssue: " [@__commenter__](***/__commenter__)  you ",
	}}

	cli, ok := bot.cli.(*mockClient)
	assert.Equal(t, true, ok)
	action := "open1"

	cli.method = ""
	pass := bot.checkCommenterPermission(org, repo, number, commenter, commenter, client.CommentOnIssue, action)
	assert.Equal(t, true, pass)
	execMethod1 := cli.method
	assert.Equal(t, "", execMethod1)

	author := commenter + "ff"
	case2 := "CheckPermission"
	cli.method = case2
	pass1 := bot.checkCommenterPermission(org, repo, number, author, commenter, client.CommentOnIssue, action)
	assert.Equal(t, false, pass1)
	execMethod2 := cli.method
	assert.Equal(t, case2, execMethod2)

	case3 := "CreateIssueComment"
	cli.method = ""
	cli.successfulCheckPermission = true
	cli.permission = false
	bot.checkCommenterPermission(org, repo, number, author, commenter, client.CommentOnIssue, action)
	execMethod3 := cli.method
	assert.Equal(t, case3, execMethod3)

	case4 := "CreatePRComment"
	cli.method = ""
	bot.checkCommenterPermission(org, repo, number, author, commenter, client.CommentOnPR, action)
	execMethod4 := cli.method
	assert.Equal(t, case4, execMethod4)
}
