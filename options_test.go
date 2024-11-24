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
	"github.com/opensourceways/server-common-lib/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGatherOptions(t *testing.T) {

	args := []string{
		"***",
		"--port=8511",
		"--config-file=" + findTestdata(t, "testdata"+string(os.PathSeparator)+"config.yaml"),
	}

	opt := new(robotOptions)
	_, _ = opt.gatherOptions(flag.NewFlagSet(args[0], flag.ExitOnError), args[1:]...)
	assert.Equal(t, true, opt.shutdown)
	assert.Equal(t, "webhook", opt.service.HandlePath)
	assert.Equal(t, 8511, opt.service.Port)

	args = []string{
		"***",
		"--port=8511",
		"--config-file=" + findTestdata(t, "testdata"+string(os.PathSeparator)+"config11.yaml"),
		"--handle-path=gitcode-hook",
	}

	opt = new(robotOptions)
	_, _ = opt.gatherOptions(flag.NewFlagSet(args[0], flag.ExitOnError), args[1:]...)
	assert.Equal(t, true, opt.shutdown)

	args = []string{
		"***",
		"--port=8511",
		"--config-file=" + findTestdata(t, "testdata"+string(os.PathSeparator)+"config12.yaml"),
		"--handle-path=gitcode-hook",
		"--token-path=token",
		"--del-token=false",
	}

	_, _ = opt.gatherOptions(flag.NewFlagSet(args[0], flag.ExitOnError), args[1:]...)
	assert.Equal(t, true, opt.shutdown)

	args = []string{
		"***",
		"--port=8511",
		"--config-file=" + findTestdata(t, "testdata"+string(os.PathSeparator)+"config.yaml"),
		"--handle-path=gitcode-hook",
		"--token-path=token12",
		"--del-token=false",
	}

	_, _ = opt.gatherOptions(flag.NewFlagSet(args[0], flag.ExitOnError), args[1:]...)
	assert.Equal(t, true, opt.shutdown)

	args[4] = "--token-path=" + findTestdata(t, "testdata"+string(os.PathSeparator)+"token")

	opt = new(robotOptions)
	got, token := opt.gatherOptions(flag.NewFlagSet(args[0], flag.ExitOnError), args[1:]...)
	assert.Equal(t, false, opt.shutdown)
	assert.Equal(t, "gitcode-hook", opt.service.HandlePath)
	want := &configuration{}
	_ = utils.LoadFromYaml(findTestdata(t, "testdata"+string(os.PathSeparator)+"config.yaml"), want)
	assert.Equal(t, *want, *got)
	assert.Equal(t, "gf112421415123123asdada", string(token))
}
