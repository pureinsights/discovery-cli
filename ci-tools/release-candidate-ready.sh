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

usage="Marks the pull request of the branch as ready to close

Usage: $(basename "$0") -b branch [FLAGS]

Flags:
  -h            the current help message
"

while getopts b:h flag
do
  case "${flag}" in
    b) branch=${OPTARG} ;;
    h) echo "${usage}"
       exit 0 ;;
  esac
done

if [[ -z ${branch} ]]; then
  echo "${usage}" >&2
  exit 1
fi


. ci-tools/config-branches.sh

if [[ `gh pr view ${branch} --json state --template "{{ .state }}"` == "OPEN" ]]; then
  gh label create "automerge" --color "#D0F4D4" --force
  gh pr edit ${branch} --add-label "automerge"
fi
