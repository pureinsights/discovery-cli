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

usage="Commits an updated dependency

Usage: $(basename "$0") -i id -k key -v version [FLAGS]

Flags:
  -b true|false    if a branch should be created (default: false)
  -h               the current help message

CI Variables:
  branch-ref       the resulting commit or branch reference
"

while getopts i:b:k:v:h flag
do
  case "${flag}" in
    b) branch=${OPTARG} ;;
    i) id=${OPTARG} ;;
    k) key=${OPTARG} ;;
    v) version=${OPTARG} ;;
    h) echo "${usage}"
       exit 0 ;;
  esac
done

if [[ -z ${id} ]] || [[ -z ${key} ]] || [[ -z ${version} ]]; then
  echo "${usage}" >&2
  exit 1
fi


# Configure Git with the CI account
. ci-tools/config-git-bot.sh

updateMessage="Update '${key}' dependency to ${version}"
commitMessage="autocommit: ${updateMessage}"
echo "update-message=${updateMessage}" >> $GITHUB_OUTPUT

git add .

if [[ ${branch} == 'true' ]]; then
  branchName="bot/update-upstream-dependencies-${id}"

  git checkout -b "${branchName}"
  git commit -m "${commitMessage}"
  git push --set-upstream origin "${branchName}"
else
  git commit -m "${commitMessage}"
  git push
fi
