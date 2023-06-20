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

usage="Validates the branch for the release candidate

Usage: $(basename "$0")

Flags:
  -h             the current help message"

while getopts h flag
do
  case "${flag}" in
    h) echo "${usage}"
       exit 0 ;;
  esac
done

. ci-tools/config-branches.sh

branch=$( git branch --show-current )
if [[ ${branch} != ${DEVELOP_BRANCH} ]]; then
  echo "A release candidate can only be created from the ${DEVELOP_BRANCH} branch"
  exit 1
fi