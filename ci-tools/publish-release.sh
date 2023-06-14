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

usage="Creates a tag and publishes the release for the current Git repository

Usage: $(basename "$0") -v version [FLAGS]

Flags:
  -n notes       the release notes
  -h             the current help message"

while getopts v:n:h flag
do
  case "${flag}" in
    v) version=${OPTARG} ;;
    n) notes=${OPTARG} ;;
    h) echo "${usage}"
       exit 0 ;;
  esac
done

if [[ -z ${version} ]]; then
  echo "${usage}" >&2
  exit 1
fi


# Identify the ticket numbers
tickets=()
for ticket in "$(echo "${notes}" | grep -Po "[A-Z]+-[0-9]+")"; do
  tickets+=("${ticket}")
done


# Remove duplicates
declare -A unique
for k in ${tickets[@]} ; do unique[$k]=1 ; done


# Link each ticket to Jira
for ticket in ${!unique[@]}; do
  link="[${ticket}](https://pureinsights.atlassian.net/browse/${ticket})"
  notes=$( echo "${notes}" | sed -e "s#${ticket}\([[:space:]]\|:\|/\|-\)#${link}#g" )
done


# Clean up empty lines and empty headers
notes=$( echo "${notes}" | tac )
notes=$( echo "${notes}" | sed -Ez '$ s/\n+$//' )
while [[ "${notes}" == \#* ]]; do
  notes=$( echo "${notes}" | tail -n +2 )
done

notes=$( echo "${notes}" | tac )

# Configure Git with the CI account
. ci-tools/config-git-bot.sh


# Tag and release
git tag --force \
  --annotate "${version}" \
  --message "Release ${version}"

git push origin "${version}"

gh release create "${version}" \
  --title "${version}" \
  --notes "${notes}"
