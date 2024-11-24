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
	"flag"
	"github.com/opensourceways/robot-framework-lib/client"
	"github.com/opensourceways/robot-framework-lib/config"
	"github.com/opensourceways/server-common-lib/secret"
	"github.com/sirupsen/logrus"
	"os"
)

type robotOptions struct {
	service   config.FrameworkOptions
	delToken  bool
	shutdown  bool
	tokenPath string
}

func (o *robotOptions) loadToken(fs *flag.FlagSet) func() []byte {
	fs.StringVar(
		&o.tokenPath, "token-path", "",
		"Path to the file containing the token secret.",
	)
	fs.BoolVar(
		&o.delToken, "del-token", true,
		"whether to delete token secret file.",
	)

	return func() []byte {
		token, err := secret.LoadSingleSecret(o.tokenPath)
		if err != nil {
			logrus.Errorf("load token, err:%s", err.Error())
			o.shutdown = true
		}
		if o.delToken {
			if err = os.Remove(o.tokenPath); err != nil {
				logrus.Errorf("remove token, err:%s", err.Error())
				o.shutdown = true
			}
		}
		return token
	}
}

func (o *robotOptions) gatherOptions(fs *flag.FlagSet, args ...string) (*configuration, []byte) {

	o.service.AddFlagsComposite(fs)
	tokenFunc := o.loadToken(fs)

	_ = fs.Parse(args)

	if err := o.service.ValidateComposite(); err != nil {
		logrus.Errorf("invalid service options, err:%s", err.Error())
		o.shutdown = true
		return nil, nil
	}
	configmap, err := config.NewConfigmapAgent(&configuration{}, o.service.ConfigFile)
	if err != nil {
		logrus.Errorf("load config, err:%s", err.Error())
		return nil, nil
	}

	cnf := configmap.GetConfigmap().(*configuration)
	client.SetSigInfoBaseURL(cnf.SigInfoURL)
	client.SetCommunityName(cnf.CommunityName)
	commentNoPermissionOperateIssue = cnf.CommentNoPermissionOperateIssue
	commentIssueNeedsLinkPR = cnf.CommentIssueNeedsLinkPR
	commentListLinkingPullRequestsFailure = cnf.CommentListLinkingPullRequestsFailure
	commentNoPermissionOperatePR = cnf.CommentNoPermissionOperatePR

	return cnf, tokenFunc()
}
