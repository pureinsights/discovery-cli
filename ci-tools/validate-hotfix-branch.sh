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

usage="Validates the hotfix branch

Usage: $(basename "$0") [FLAGS]

Flags:
  -t        the type of reference branch to create the hotfix
  -h        the current help message"

while getopts t:h flag
do
  case "${flag}" in
    t) type=${OPTARG} ;;
    h) echo "${usage}"
       exit 0 ;;
  esac
done

if [[ -n "${type}" ]]; then
  if [[ "${type}" != 'tag' ]]; then
    echo "A hotfix can only be created from a tag"
    exit 1
  fi
else
  branch=$( git branch --show-current )
  if [[ ${branch} != hotfix/* ]]; then
    echo "The branch is not a hotfix"
    exit 1
  fi
fi
