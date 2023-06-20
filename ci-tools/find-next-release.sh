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

usage="Finds the version for the next release of the current repository

Usage: $(basename "$0") -t type [FLAGS]

Flags:
  -h                    the current help message

CI Variables:
  release-version       the version for the next release
  branch-name           the name of the branch for the release
  branch-heads          the number of heads for the branch for the release
"

while getopts t:h flag
do
  case "${flag}" in
    t) type=${OPTARG} ;;
    h) echo "${usage}"
       exit 0 ;;
  esac
done

if [[ -z ${type} ]]; then
  echo "${usage}" >&2
  exit 1
fi

. ci-tools/config-branches.sh
. ci-tools/config-git-bot.sh

git fetch

echo "Finding current version on default branch"
git checkout -b "ci/master" --track "origin/${MAIN_BRANCH}"
cp -f semver.properties semver.properties.bkp

if [[ ${type} == 'minor' ]]; then
  ./gradlew -Dversion.prerelease="" -Dversion.buildmeta="" incrementMinor
elif [[ ${type} == 'major' ]]; then
  ./gradlew -Dversion.prerelease="" -Dversion.buildmeta="" incrementMajor
else
  echo "Invalid release type: ${type}"
  exit 1
fi

releaseVersion=$( ci-tools/read-version.sh )
echo "Configuring version for Release Candidate ${releaseVersion}"
echo "release-version=${releaseVersion}" >> $GITHUB_OUTPUT

ci-tools/find-branch.sh -b "rls/${releaseVersion}"

mv -f semver.properties.bkp semver.properties
