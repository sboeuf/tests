#!/bin/bash
#
# Copyright (c) 2017 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

function get_repo_slug() {
	[ -n "$SEMAPHORE_REPO_SLUG" ] && { echo "$SEMAPHORE_REPO_SLUG"; return; }
	[ -n "$TRAVIS_REPO_SLUG" ] && { echo "$TRAVIS_REPO_SLUG"; return; }
}

if [ "$TRAVIS" != true ]
then
	exit 0
fi

if [ "$(get_repo_slug)" != "clearcontainers/tests" ]
then
	exit 0
fi

echo "Set up tests repo"

# Check the commits in the branch
checkcommits_dir="cmd/checkcommits"
(cd "${checkcommits_dir}" && make)
checkcommits \
	--need-fixes \
	--need-sign-offs \
	--body-length 72 \
	--subject-length 75 \
	--verbose