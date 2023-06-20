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

usage="Checks if a release already exists

Usage: $(basename "$0") -v version [FLAGS]

Flags:
  -h             the current help message"

while getopts v:h flag
do
  case "${flag}" in
    v) version=${OPTARG} ;;
    h) echo "${usage}"
       exit 0 ;;
  esac
done

if [[ -z ${version} ]]; then
  echo "${usage}" >&2
  exit 1
fi


if [[ `gh release view "${version}" --json tagName --template "{{ .tagName }}"` == "${version}" ]]; then
  echo "The release for tag "${version}" already exists"
  exit 1
fi
