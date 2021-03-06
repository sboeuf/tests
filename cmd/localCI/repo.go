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

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
)

// Repo represents the repository under test
// For more information about this structure take a look to the README
type Repo struct {
	// URL is the url of the repository
	URL string

	// MasterBranch is the master branch of this repository
	MasterBranch string

	// PR is the pull request number
	PR int

	// RefreshTime is the time to wait for checking if a pull request needs to be tested
	RefreshTime string

	// Toke is the repository access token
	Token string

	// Setup contains the conmmands needed to setup the environment
	Setup []string

	// Run contains the commands to run the test
	Run []string

	// Teardown contains the commands to be executed once Run ends
	Teardown []string

	// OnSuccess contains the commands to be executed if Setup, Run and Teardown finished correctly
	OnSuccess []string

	// OnFailure contains the commands to be executed if any of Setup, Run or Teardown fail
	OnFailure []string

	// TTY specify whether a tty must be allocate to run the stages
	TTY bool

	// PostOnSuccess is the comment to be posted if the test finished correctly
	PostOnSuccess string

	// PostOnFailure is the comment to be posted if the test fails
	PostOnFailure string

	// LogDir is the logs directory
	LogDir string

	// Language is the language of the repository
	Language RepoLanguage

	// CommentTrigger is the comment that must be present to trigger the test
	CommentTrigger RepoComment

	// LogServer contains the information of the server where the logs must be placed
	LogServer LogServer

	// Whitelist is the list of users whose pull request can be tested
	Whitelist string

	// cvr control version repository
	cvr CVR

	// refresh is RefreshTime once parsed
	refresh time.Duration

	// env contains the environment variables to be used in each stage
	env []string

	// whitelistUsers is the whitelist once parsed
	whitelistUsers []string

	// logger of the repository
	logger *logrus.Entry

	// prConfig is the configuration used to create pull request objects
	prConfig pullRequestConfig
}

const (
	logDirMode    = 0755
	logFileMode   = 0664
	logServerUser = "root"
)

var defaultEnv = []string{"CI=true", "LOCALCI=true"}

var runTestsInParallel bool

var testLock sync.Mutex

func (r *Repo) setupCvr() error {
	var err error

	// validate url
	r.URL = strings.TrimSpace(r.URL)
	if len(r.URL) == 0 {
		return fmt.Errorf("missing repository url")
	}

	// set repository logger
	r.logger = ciLog.WithFields(logrus.Fields{
		"Repo": r.URL,
	})

	// get the control version repository
	r.cvr, err = newCVR(r.URL, r.Token)
	r.logger.Debugf("control version repository: %#v", r.cvr)

	return err
}

func (r *Repo) setupLogServer() error {
	if reflect.DeepEqual(r.LogServer, LogServer{}) {
		return nil
	}

	if len(r.LogServer.IP) == 0 {
		return fmt.Errorf("missing server ip")
	}

	if len(r.LogServer.User) == 0 {
		r.LogServer.User = logServerUser
	}

	if len(r.LogServer.Dir) == 0 {
		r.LogServer.Dir = defaultLogDir
	}

	return nil
}

func (r *Repo) setupLogDir() error {
	// create log directory
	if err := os.MkdirAll(r.LogDir, logDirMode); err != nil {
		return err
	}

	return nil
}

func (r *Repo) setupRefreshTime() error {
	var err error

	// validate refresh time
	r.refresh, err = time.ParseDuration(r.RefreshTime)
	if err != nil {
		return fmt.Errorf("failed to parse refresh time '%s' %s", r.RefreshTime, err)
	}

	return nil
}

func (r *Repo) setupCommentTrigger() error {
	if reflect.DeepEqual(r.CommentTrigger, RepoComment{}) {
		return nil
	}

	if len(r.CommentTrigger.Comment) == 0 {
		return fmt.Errorf("missing comment trigger")
	}

	return nil
}

func (r *Repo) setupLanguage() error {
	return r.Language.setup()
}

func (r *Repo) setupStages() error {
	if len(r.Run) == 0 {
		return fmt.Errorf("missing run commands")
	}

	return nil
}

func (r *Repo) setupWhitelist() error {
	// get the list of users
	r.whitelistUsers = strings.Split(r.Whitelist, ",")
	return nil
}

func (r *Repo) setupEnvars() error {
	// add environment variables
	r.env = os.Environ()
	r.env = append(r.env, defaultEnv...)
	repoSlug := fmt.Sprintf("LOCALCI_REPO_SLUG=%s", r.cvr.getRepoSlug())
	r.env = append(r.env, repoSlug)

	return nil
}

// setup the repository. This method MUST BE called before use any other
func (r *Repo) setup() error {
	var err error

	setupFuncs := []func() error{
		r.setupCvr,
		r.setupRefreshTime,
		r.setupLogDir,
		r.setupLogServer,
		r.setupCommentTrigger,
		r.setupLanguage,
		r.setupStages,
		r.setupWhitelist,
		r.setupEnvars,
	}

	for _, setupFunc := range setupFuncs {
		if err = setupFunc(); err != nil {
			return err
		}
	}

	r.prConfig = pullRequestConfig{
		cvr:            r.cvr,
		logger:         r.logger,
		commentTrigger: r.CommentTrigger,
		postOnFailure:  r.PostOnFailure,
		postOnSuccess:  r.PostOnSuccess,
		whitelist:      r.Whitelist,
	}

	r.logger.Debugf("control version repository: %#v", r.cvr)

	return nil
}

// loop to monitor the repository
func (r *Repo) loop() {
	revisionsTested := make(map[string]revision)

	r.logger.Debugf("monitoring in a loop the repository: %+v", *r)

	appendPullRequests := func(revisions *[]revision, prs []int) error {
		for _, prNumber := range prs {
			r.logger.Debugf("requesting pull request %d", prNumber)
			pr, err := newPullRequest(prNumber, r.prConfig)
			if err != nil {
				return fmt.Errorf("failed to get pull request '%d' %s", prNumber, err)
			}
			*revisions = append(*revisions, pr)
		}
		return nil
	}

	for {
		var revisionsToTest []revision

		// append master branch
		r.logger.Debugf("requesting master branch: %s", r.MasterBranch)
		branch, err := newRepoBranch(r.MasterBranch, r.cvr, r.logger)
		if err != nil {
			r.logger.Warnf("failed to get master branch %s: %s", r.MasterBranch, err)
		} else {
			revisionsToTest = append(revisionsToTest, branch)
		}

		// append pull requests
		if r.PR != 0 {
			// if PR is not 0 then we have to monitor just one PR
			if err = appendPullRequests(&revisionsToTest, []int{r.PR}); err != nil {
				r.logger.Warnf("failed to append pull request %d", r.PR, err)
			}
		} else {
			// append open pull request
			r.logger.Debugf("requesting open pull requests")
			prs, err := r.cvr.getOpenPullRequests()
			if err != nil {
				r.logger.Warnf("failed to get open pull requests: %s", err)
			} else if err = appendPullRequests(&revisionsToTest, prs); err != nil {
				r.logger.Warnf("failed to append pull requests %+v: %s", prs, err)
			}
		}

		// test only if there are at least 1 revision
		if len(revisionsToTest) > 0 {
			r.logger.Debugf("testing revisions: %#v", revisionsToTest)
			r.testRevisions(revisionsToTest, &revisionsTested)
		}

		r.logger.Debugf("going to sleep: %s", r.RefreshTime)
		time.Sleep(r.refresh)
	}
}

func (r *Repo) testRevisions(revisions []revision, revisionsTested *map[string]revision) {
	// remove revisions that are not in the list of open pull request and already tested
	for k, v := range *revisionsTested {
		found := false
		// iterate over open pull requests and master branch
		for _, r := range revisions {
			if r.id() == k {
				found = true
				break
			}
		}

		if !found && !v.isBeingTested() {
			delete((*revisionsTested), k)
		}
	}

	for _, revision := range revisions {
		tested, ok := (*revisionsTested)[revision.id()]
		if ok {
			// checking if the old version of the PR is being tested
			if tested.isBeingTested() {
				r.logger.Debugf("revision is being tested: %#v", tested)
				continue
			}
			if revision.equal(tested) {
				r.logger.Debugf("revision was already tested: %#v", revision)
				continue
			}
		}

		// check if the revision can be tested
		if err := revision.canBeTested(); err != nil {
			r.logger.Debugf("revision %s cannot be tested: %s", revision.id(), err)
			continue
		}

		// setup revision
		langEnv, err := r.setupRevision(revision)
		if err != nil {
			r.logger.Errorf("failed to setup revision %#v: %s", revision, err)
			continue
		}

		// cleanup revision
		defer func() {
			err = langEnv.cleanup()
			if err != nil {
				r.logger.Error(err)
			}
		}()

		// test revision
		if runTestsInParallel {
			go func() {
				if err := r.testRevision(revision, langEnv); err != nil {
					r.logger.Errorf("failed to test revision %#v %s", revision, err)
				}
			}()
		} else {
			testLock.Lock()
			if err := r.testRevision(revision, langEnv); err != nil {
				r.logger.Errorf("failed to test revision %#v %s", revision, err)
			}
			testLock.Unlock()
		}

		// copy the PR that was tested
		(*revisionsTested)[revision.id()] = revision
	}
}

// test the pull request specified in the configuration file
// if pr does not exist an error is returned
func (r *Repo) test() error {
	if r.PR == 0 {
		return fmt.Errorf("Missing pull request number in configuration file")
	}

	rev, err := newPullRequest(r.PR, r.prConfig)
	if err != nil {
		return fmt.Errorf("failed to get pull request %d %s", r.PR, err)
	}

	// run tests in parallel does not make sense when
	// we are just testing one pull request
	runTestsInParallel = false

	// setup revision
	langEnv, err := r.setupRevision(rev)
	if err != nil {
		return fmt.Errorf("failed to setup revision %#v: %s", rev, err)
	}

	// cleanup revision
	defer func() {
		err = langEnv.cleanup()
		if err != nil {
			r.logger.Error(err)
		}
	}()

	// test revision
	return r.testRevision(rev, langEnv)
}

// setupRevision generates a language environment, downloads the revision
// and creates the logs directory
func (r *Repo) setupRevision(rev revision) (languageConfig, error) {
	// generate a new environment to run the stages
	langEnv, err := r.Language.generateEnvironment(r.cvr.getProjectSlug())
	if err != nil {
		return languageConfig{}, err
	}

	// download the revision
	if err = rev.download(langEnv.workingDir); err != nil {
		return languageConfig{}, err
	}

	// cleanup and set the log directory of the pull request
	logDir := filepath.Join(r.LogDir, rev.logDirName())
	_ = os.RemoveAll(logDir)
	if err = os.MkdirAll(logDir, logDirMode); err != nil {
		return languageConfig{}, err
	}

	return langEnv, nil
}

// testRevision tests a specific revision
// returns an error if the test fail
func (r *Repo) testRevision(rev revision, langEnv languageConfig) error {
	config := stageConfig{
		logger:     r.logger,
		workingDir: langEnv.workingDir,
		tty:        r.TTY,
		logDir:     filepath.Join(r.LogDir, rev.logDirName()),
	}

	// set environment variables
	config.env = r.env

	// appends language environment variables
	if len(langEnv.env) > 0 {
		config.env = append(config.env, langEnv.env...)
	}

	// copy logs to server if we have an IP address
	if len(r.LogServer.IP) != 0 {
		defer func() {
			if err := r.LogServer.copy(config.logDir); err != nil {
				r.logger.Errorf("failed to copy log dir %s to server %+v", config.logDir, r.LogServer)
			}
		}()
	}

	r.logger.Debugf("stage config: %+v", config)

	stages := map[string]stage{
		"setup":     stage{name: "setup", commands: r.Setup},
		"run":       stage{name: "run", commands: r.Run},
		"teardown":  stage{name: "teardown", commands: r.Teardown},
		"onSuccess": stage{name: "onSuccess", commands: r.OnSuccess},
		"onFailure": stage{name: "onFailure", commands: r.OnFailure},
	}

	// run test
	return rev.test(config, stages)
}
