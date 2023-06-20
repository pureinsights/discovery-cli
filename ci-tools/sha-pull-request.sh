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

usage="Creates a pull request to merge the branch with the given SHA into the target branch
Usage: $(basename "$0") -S SHA -r 'owner/repository' -s source -t target -i id [FLAGS]
Flags:
  -m        the message for the pull request
  -h        the current help message"

while getopts S:s:r:m:t:i:h flag
do
  case "${flag}" in
    S) sha=${OPTARG} ;;
    r) repository=${OPTARG} ;;
    s) source=${OPTARG} ;;
    m) commitMessage=${OPTARG} ;;
    t) target=${OPTARG} ;;
    i) id=${OPTARG} ;;
    h) echo "${usage}"
       exit 0 ;;
  esac
done

if [[ -z ${sha} ]] || [[ -z ${repository} ]] || [[ -z ${source} ]] || [[ -z ${target} ]] || [[ -z ${id} ]]; then
  echo "${usage}" >&2
  exit 1
fi

branchName="bot/sha-${id}"

if [[ -z ${commitMessage} ]]; then
  commitMessage="Merge ${source} into ${target}"
fi

. ci-tools/config-git-bot.sh

# Restore SHA
gh api \
  --method POST \
  -H "Accept: application/vnd.github+json" \
  "/repos/${repository}/git/refs" \
  -f ref="refs/heads/${branchName}" \
  -f sha="${sha}"

git reset --hard HEAD
git fetch
git pull

gradlePluginsVersion=
coreVersion=
stagingVersion=
ingestionVersion=

git checkout -b "ci/target-${id}" --track "origin/${target}"
cp -v semver.properties "semver.properties.${id}"

# Find version of internal artifacts
for f in $(find . -name "gradle.properties"); do
  if [[ -z ${gradlePluginsVersion} ]]; then
    gradlePluginsVersion=$(grep 'gradlePluginsVersion=' ${f} | cut -d= -f2)
  fi

  if [[ -z ${coreVersion} ]]; then
    coreVersion=$(grep 'coreVersion=' ${f} | cut -d= -f2)
  fi

  if [[ -z ${stagingVersion} ]]; then
    stagingVersion=$(grep 'stagingVersion=' ${f} | cut -d= -f2)
  fi

  if [[ -z ${ingestionVersion} ]]; then
    ingestionVersion=$(grep 'ingestionVersion=' ${f} | cut -d= -f2)
  fi
done

git checkout "${branchName}"

git fetch
git pull

mv -f -v "semver.properties.${id}" semver.properties

# Restore version of internal artifacts
for f in $(find . -name "gradle.properties"); do
  if [[ -n ${gradlePluginsVersion} ]]; then
    sed -i "s/gradlePluginsVersion=.*/gradlePluginsVersion=${gradlePluginsVersion}/g" ${f}
  fi

  if [[ -n ${coreVersion} ]]; then
    sed -i "s/coreVersion=.*/coreVersion=${coreVersion}/g" ${f}
  fi

  if [[ -n ${stagingVersion} ]]; then
    sed -i "s/stagingVersion=.*/stagingVersion=${stagingVersion}/g" ${f}
  fi

  if [[ -n ${ingestionVersion} ]]; then
    sed -i "s/ingestionVersion=.*/ingestionVersion=${ingestionVersion}/g" ${f}
  fi

  git add ${f}
done

# Commit changes and open the PR
git add .
git rm -r --cached ./ci-tools
git commit -m "${commitMessage}"
git push --set-upstream origin "${branchName}"

gh pr create \
  --assignee "@me" \
  --base "refs/heads/${target}" \
  --title "${commitMessage}" \
  --body "Automated PR to merge changes from ${source}"