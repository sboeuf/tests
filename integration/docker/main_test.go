// Copyright (c) 2017 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package docker

import (
	"testing"

	. "github.com/clearcontainers/tests"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	shouldFail    = true
	shouldNotFail = false
)

func randomDockerName() string {
	return RandID(30)
}

func runDockerCommand(expectedExitCode int, args ...string) string {
	cmd := NewCommand(Docker, args...)
	Expect(cmd).ToNot(BeNil())
	stdout, _, exitCode := cmd.Run()
	Expect(exitCode).To(Equal(expectedExitCode))
	return stdout
}

func TestIntegration(t *testing.T) {
	// before start we have to download the docker images
	images := []string{
		Image,
		AlpineImage,
	}

	for _, i := range images {
		_, _, exitCode := DockerPull(i)
		if exitCode != 0 {
			t.Fatalf("failed to pull docker image: %s\n", i)
		}
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}
