#!/bin/bash

#
# Copyright (c) 2022 Pureinsights Technology Ltd. All rights reserved.
#
# Permission to use, copy, modify or distribute this software and its
# documentation for any purpose is subject to a licensing agreement with
# Pureinsights Technology Ltd.
#
# All information contained within this file is the property of
# Pureinsights Technology Ltd. The distribution or reproduction of this
# file or any information contained within is strictly forbidden unless
# prior written permission has been granted by Pureinsights Technology Ltd.
#

usage="Commits the final version for a release

Usage: $(basename "$0") -v version [FLAGS]

Flags:
  -h                 the current help message

CI Variables:
  branch-ref         the commit or branch reference for the release
"

while getopts v:h flag
do
  case "${flag}" in
    v) version=${OPTARG} ;;
    h) echo "${usage}"
       exit 0 ;;
  esac
done

if [[ -z ${version} ]]; then
  echo "${usage}" >&2
  exit 1
fi


. ci-tools/config-branches.sh
. ci-tools/config-git-bot.sh

if [[ $(ci-tools/read-version.sh) != "${version}" ]]; then
  ./gradlew -Dversion.semver="${version}"

  git add semver.properties
  commitMessage=`git commit -m "Configure version for Release ${version}"`
  echo "${commitMessage}"
  git push
  commitSHA=$(git rev-parse `echo "${commitMessage}" | head -n 1 | cut -d "[" -f2 | cut -d "]" -f1 | awk '{print $2}'`)
  echo "branch-ref=${commitSHA}" >> $GITHUB_OUTPUT
else
  echo "branch-ref=refs/heads/${MAIN_BRANCH}" >> $GITHUB_OUTPUT
fi
