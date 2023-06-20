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

usage="Creates a pull request to merge the source branch into the target branch

Usage: $(basename "$0") -s source -t target -i id [FLAGS]

Flags:
  -S        SHA with the commit ID
  -f        create the pull request without processing the branches
  -h        the current help message"

while getopts S:s:t:i:fh flag
do
  case "${flag}" in
    s) source=${OPTARG} ;;
    S) sha=${OPTARG} ;;
    t) target=${OPTARG} ;;
    i) id=${OPTARG} ;;
    f) force=1 ;;
    h) echo "${usage}"
       exit 0 ;;
  esac
done

if [[ -z ${source} ]] || [[ -z ${target} ]] || [[ -z ${id} ]]; then
  echo "${usage}" >&2
  exit 1
fi

commitMessage="Merge ${source} into ${target}"

git fetch
git pull

if [[ "${force}" -ne 1 ]]; then
  . ci-tools/config-git-bot.sh

  echo "Backing up semver.properties"
  git checkout -b "ci/target-${id}" --track "origin/${target}"
  cp -v semver.properties "semver.properties.${id}"

  echo "Using target semver.properties"
  if [[ -z "${sha}" ]]; then
    git checkout -b "ci/source-${id}" --track "origin/${source}"
  else
    git checkout "${sha}"
  fi
  mv -f -v "semver.properties.${id}" semver.properties
  git add semver.properties

  # Create branch and pull request
  branchName="bot/merge-${id}"

  git checkout -b "${branchName}"
  git commit -m "${commitMessage}"
  git push --set-upstream origin "${branchName}"
fi

gh pr create \
  --assignee "@me" \
  --base "refs/heads/${target}" \
  --title "${commitMessage}" \
  --body "Automated PR to merge changes from ${source}"
