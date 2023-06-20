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

usage="Commits the version of the current branch

Usage: $(basename "$0") -o oldVersion -n newVersion [FLAGS]

Flags:
  -b branch      the origin branch for the push. If not given, the current one is used
  -s             skip the CI workflow that could be triggered by the commit
  -h             the current help message

CI Variables:
  branch-ref     the resulting commit or branch reference
"

while getopts b:o:n:hs flag
do
  case "${flag}" in
    b) branch=${OPTARG} ;;
    o) oldVersion=${OPTARG} ;;
    n) newVersion=${OPTARG} ;;
    s) skip=1 ;;
    h) echo "${usage}"
       exit 0 ;;
  esac
done

if [[ -z ${oldVersion} ]] || [[ -z ${newVersion} ]]; then
  echo "${usage}" >&2
  exit 1
fi


# Configure Git with the CI account
. ci-tools/config-git-bot.sh

commitMessage="autocommit: bump version from ${oldVersion} to ${newVersion}"
if [[ ${skip} -eq 1 ]]; then
  commitMessage="${commitMessage} [skip ci]"
fi

git add semver.properties
commitCommand=`git commit -m "${commitMessage}"`
echo "${commitCommand}"

if [[ -z "${branch}" ]]; then
  git push
else
  git push origin HEAD:${branch}
fi

commitSHA=$(git rev-parse `echo "${commitCommand}" | head -n 1 | cut -d "[" -f2 | cut -d "]" -f1 | awk '{print $2}'`)
echo "branch-ref=${commitSHA}" >> $GITHUB_OUTPUT

