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

usage="Finds the release before the given version

Usage: $(basename "$0") [FLAGS]

Flags:
  -v            the reference version. If not given, the current one will be used
  -h            the current help message
"

while getopts v:h flag
do
  case "${flag}" in
    v) version=${OPTARG} ;;
    h) echo "${usage}"
       exit 0 ;;
  esac
done

if [[ -z ${version} ]]; then
  version=$( ci-tools/read-version.sh )
  exit 1
fi

releases="$(gh release list | cut -f1)"
releases="$(printf "${releases}\n${version}" | sort -V -r)"

versionPosition=$(echo "${releases}" | grep -n "${version}" | cut -f1 -d:)

echo "$releases" | head -n $(($versionPosition+1)) | tail -1
