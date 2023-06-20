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

usage="Updates a Gradle dependency

Usage: $(basename "$0") -d dependenciesA:versionA[,dependencyB:versionB] [FLAGS]

Flags:
  -h             the current help message"

while getopts d:h flag
do
  case "${flag}" in
    d) dependencies=${OPTARG} ;;
    h) echo "${usage}"
       exit 0 ;;
  esac
done

if [[ -z ${dependencies} ]]; then
  echo "${usage}" >&2
  exit 1
fi


. ci-tools/config-git-bot.sh

IFS=',' read -ra dependency <<< "${dependencies}"

for element in "${dependency[@]}"; do
  key=`echo ${element} | cut -d ":" -f1`
  version=`echo ${element} | cut -d ":" -f2`

  ./gradlew updateVersion --dependencyKey="${key}" --dependencyVersion="${version}"
done

if [[ -z `git diff` ]]; then
  echo "Dependencies are already in their expected version"
else
  git add .
  git commit -m "autocommit: Update dependencies"
  git push
fi
