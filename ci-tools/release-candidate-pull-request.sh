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

usage="Creates a pull request for a release candidate

Usage: $(basename "$0") -b branch -v version -t type [FLAGS]

Flags:
  -d trigger-downstream   the type of trigger for release candidates
                          in downstream repositories
  -h                      the current help message
"

while getopts b:d:t:v:h flag
do
  case "${flag}" in
    b) branch=${OPTARG} ;;
    d) downstream=${OPTARG} ;;
    t) type=${OPTARG} ;;
    v) version=${OPTARG} ;;
    h) echo "${usage}"
       exit 0 ;;
  esac
done

if [[ -z ${branch} ]] || [[ -z ${type} ]] || [[ -z ${version} ]]; then
  echo "${usage}" >&2
  exit 1
fi


. ci-tools/config-branches.sh

if [[ `gh pr view ${branch} --json state --template "{{ .state }}"` != "OPEN" ]]; then
  gh label create "release:${type}" --color "#037926" --force

  gh pr create \
    --assignee "@me" \
    --base "refs/heads/${MAIN_BRANCH}" \
    --title "Release Candidate: ${version}" \
    --body "Release Candidate for version ${version} (${type})" \
    --label "release:${type}"

  if [[ -n ${downstream} ]] && [[ ${downstream} != 'none' ]]; then
    gh label create "downstream:${downstream}" --color "#EFBE7B" --force
    gh pr edit ${branch} --add-label "downstream:${downstream}"
  fi
else
  echo "An open pull request for the ${branch} branch already exists"
fi
