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

usage="Creates the branch for a hotfix

Usage: $(basename "$0") -b branch -v version [FLAGS]

Flags:
  -h             the current help message"

while getopts b:v:h flag
do
  case "${flag}" in
    b) branch=${OPTARG} ;;
    v) version=${OPTARG} ;;
    h) echo "${usage}"
       exit 0 ;;
  esac
done

if [[ -z ${branch} ]] || [[ -z ${version} ]]; then
  echo "${usage}" >&2
  exit 1
fi


# Configure Git with the CI account
. ci-tools/config-git-bot.sh

# Find the branch
git fetch

./gradlew -Dversion.semver="${version}"

git checkout -b "${branch}"
git add semver.properties
git commit -m "Configure version for Hotfix ${version}"
git push --set-upstream origin "${branch}"
