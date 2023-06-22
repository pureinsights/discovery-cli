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

usage="Finds a specific branch

Usage: $(basename "$0") -b branch [FLAGS]

Flags:
  -h                  the current help message

CI Variables:
  branch-name         the name of the branch
  branch-heads        the number of heads for the branch
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


. ci-tools/config-git-bot.sh

git fetch

echo "branch-name=${branch}" >> $GITHUB_OUTPUT
echo "branch-heads=$(git ls-remote --heads origin "${branch}" | wc -l)" >> $GITHUB_OUTPUT
