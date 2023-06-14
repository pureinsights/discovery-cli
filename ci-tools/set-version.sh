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

usage="Sets the version of the branch

Usage: $(basename "$0") -v version [FLAGS]

Flags:
  -b branch      the branch to bump. If not given, the current one is used
  -c             if the new version should be committed
  -s             if the CI should be skipped after a commit
  -h             the current help message

CI Variables:
  old-version    the version before the bump
  new-version    the version after the bump
"

while getopts v:b:csh flag
do
  case "${flag}" in
    b) branch=${OPTARG} ;;
    v) version=${OPTARG} ;;
    c) commit=1 ;;
    s) skip=1 ;;
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

currentBranch=$( git branch --show-current )
if [[ -z ${branch} ]]; then
  branch="${currentBranch}"
else
  git fetch
  git checkout -b "ci/${branch}" --track "origin/${branch}"
fi

oldVersion=$( ci-tools/read-version.sh )
echo "Branch ${branch} is currently in ${oldVersion} version"

majorVersion=$(echo ${version} | cut -d. -f1)
minorVersion=$(echo ${version} | cut -d. -f2)
patchVersion=$(echo ${version} | cut -d. -f3)

./gradlew -Dversion.major=${majorVersion} \
  -Dversion.minor=${minorVersion} \
  -Dversion.patch=${patchVersion} \
  -Dversion.prerelease="${branch}"

newVersion=$( ci-tools/read-version.sh )

echo "old-version=${oldVersion}" >> $GITHUB_OUTPUT
echo "new-version=${newVersion}" >> $GITHUB_OUTPUT

echo "Branch ${branch} to ${newVersion}"

if [[ ${commit} -eq 1 ]]; then
  if [[ ${skip} -eq 1 ]]; then
    ci-tools/commit-version-bump.sh -s -o ${oldVersion} -n ${newVersion} -b ${branch}
  else
    ci-tools/commit-version-bump.sh -o ${oldVersion} -n ${newVersion} -b ${branch}
  fi
fi

if [[ "${currentBranch}" != "${branch}" ]]; then
  git switch -
fi
