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
	"os"
	"path/filepath"
	"testing"
)

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
