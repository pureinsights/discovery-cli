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

usage="Bumps the version of the branch

Usage: $(basename "$0") [FLAGS]

Flags:
  -b branch      the branch to bump. If not given, the current one is used
  -c             if the new version should be committed
  -s             if the CI should be skipped after a commit
  -h             the current help message

CI Variables:
  old-version    the version before the bump
  new-version    the version after the bump
"

while getopts b:csh flag
do
  case "${flag}" in
    b) branch=${OPTARG} ;;
    c) commit=1 ;;
    s) skip=1 ;;
    h) echo "${usage}"
       exit 0 ;;
  esac
done


. ci-tools/config-branches.sh
. ci-tools/config-git-bot.sh

if [[ -z ${branch} ]]; then
  branch=$( git branch --show-current )
else
  git fetch
  git checkout -b "ci/bump-${id}" --track "origin/${branch}"
fi

oldVersion=$( ci-tools/read-version.sh )
echo "Branch ${branch} is currently in ${oldVersion} version"

if [[ ${branch} == ${DEVELOP_BRANCH} ]]; then
  ./gradlew -Dversion.buildmeta="" -Dversion.prerelease="${branch}" incrementPatch
else
  # Find the pre-release
  if [[ ${branch} == release/* ]] || [[ ${branch} == hotfix/* ]]; then
    prerelease="RC"
  elif [[ ${branch} == feature/* ]]; then
    prerelease=${branch:8}
  else
    echo "Unsupported branch for bumping build metadata: ${branch}"
    exit 1
  fi


  # Bump the build number
  buildmeta=$(grep "version.prerelease" semver.properties | cut -d "=" -f2 | cut -d "_" -f2)
  if [[ -z ${buildmeta} ]]; then
    buildmeta=1
  else
    buildmeta=$((buildmeta+1))
  fi


  # Bump the version
  ./gradlew -Dversion.prerelease="${prerelease}_${buildmeta}"
fi


newVersion=$( ci-tools/read-version.sh )

echo "old-version=${oldVersion}" >> $GITHUB_OUTPUT
echo "new-version=${newVersion}" >> $GITHUB_OUTPUT

echo "Branch ${branch} to ${newVersion}"

if [[ ${commit} -eq 1 ]]; then
  if [[ ${skip} -eq 1 ]]; then
    ci-tools/commit-version-bump.sh -s -o ${oldVersion} -n ${newVersion}
  else
    ci-tools/commit-version-bump.sh -o ${oldVersion} -n ${newVersion}
  fi
fi
