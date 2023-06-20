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

usage="Finds snapshot dependencies

Usage: $(basename "$0")  [FLAGS]

Flags:
  -h                  the current help message

CI Variables:
  branch-name         the name of the branch
  branch-heads        the number of heads for the branch
"

while getopts h flag
do
  case "${flag}" in
    h) echo "${usage}"
       exit 0 ;;
  esac
done


snapshots=false
for f in $(find . -name "gradle.properties"); do
  if [[ $(grep -E -i '((\-SNAPSHOT|\-develop|\-RC(_[0-9]+)?)[[:space:]]?)$' ${f} | wc -l) -gt 0 ]]; then
    echo "${f} has SNAPSHOT/RC/develop dependencies"
    snapshots=true
  fi
done

if [[ ${snapshots} == true ]]; then
  echo "The repository can't have SNAPSHOT/RC/develop dependencies"
  exit 1
fi
