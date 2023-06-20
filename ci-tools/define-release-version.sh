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

usage="Defines the final version for a release

Usage: $(basename "$0") [FLAGS]

Flags:
  -h                    the current help message

CI Variables:
  release-version       the defined release version
"

while getopts h flag
do
  case "${flag}" in
    h) echo "${usage}"
       exit 0 ;;
  esac
done


./gradlew -Dversion.prerelease="" -Dversion.buildmeta=""

releaseVersion=$(ci-tools/read-version.sh)
echo "Configuring version for Release ${releaseVersion}"
echo "release-version=${releaseVersion}" >> $GITHUB_OUTPUT
