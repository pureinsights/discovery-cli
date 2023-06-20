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

usage="Finds the commit ID where the version was bumped

Usage: $(basename "$0") -v version [FLAGS]

Flags:
  -b          the branch with the pre-release
  -p          if checking for the patch version
  -h          the current help message"

while getopts b:v:ph flag
do
  case "${flag}" in
    b) branch=${OPTARG} ;;
    v) version=${OPTARG} ;;
    p) patch=1 ;;
    h) echo "${usage}"
       exit 0 ;;
  esac
done

if [[ -z ${version} ]]; then
  echo "${usage}" >&2
  exit 1
fi

if [[ "${patch}" -eq 1 ]]; then
  versionPattern="$(echo ${version} | cut -d. -f1).$(echo ${version} | cut -d. -f2).$(echo ${version} | cut -d. -f3)-"
else
  versionPattern="$(echo ${version} | cut -d. -f1).$(echo ${version} | cut -d. -f2).([0-9]+)-"
fi

if [[ -n "${branch}" ]]; then
  versionPattern+="${branch}"
else
  versionPattern+="(.*)"
fi

git log --extended-regexp --grep="^autocommit: bump version.*to ${versionPattern}( \[skip ci\])?$" --author=ci@pureinsights.com --reverse --pretty=format:"%H" --all | head -n 1
