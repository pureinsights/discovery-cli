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

usage="Finds the version for the next hotfix from the current tag

Usage: $(basename "$0") [FLAGS]

Flags:
  -h                    the current help message

CI Variables:
  hotfix-version        the version for the next hotfix
  branch-name           the name of the branch for the hotfix
  branch-heads          the number of heads for the branch for the hotfix
"

while getopts h flag
do
  case "${flag}" in
    h) echo "${usage}"
       exit 0 ;;
  esac
done

./gradlew -Dversion.prerelease="" -Dversion.buildmeta="" incrementPatch

hotfixVersion=$( ci-tools/read-version.sh )
echo "Configuring version for Hotfix ${hotfixVersion}"
echo "hotfix-version=${hotfixVersion}" >> $GITHUB_OUTPUT

ci-tools/find-branch.sh -b "htfx/${releaseVersion}"
