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

usage="Waits until the URL is available

Usage: $(basename "$0") -u URL [FLAGS]

Flags:
  -s         the status code to expect. Default: 200
  -d         the delay in seconds between retries. Default: 5
  -t         the timeout in seconds. Default: 60
  -r         the number of retries. Default: 10
  -h         the current help message
"

while getopts s:u:t:d:r:h flag
do
  case "${flag}" in
    s) status=${OPTARG} ;;
    t) timeout=${OPTARG} ;;
    u) url=${OPTARG} ;;
    d) delay=${OPTARG} ;;
    r) retry=${OPTARG} ;;
    h) echo "${usage}"
       exit 0 ;;
  esac
done

if [[ -z ${status} ]]; then
  status=200
fi

if [[ -z ${retry} ]]; then
  retry=10
fi

if [[ -z ${delay} ]]; then
  delay=5
fi

if [[ -z ${timeout} ]]; then
  timeout=60
fi

counter=0
until [[ "$(curl -s -w '%{http_code}' -o /dev/null -f ${url})" -eq ${status} ]]; do
  if [[ ${counter} -ge ${retry} ]]; then
    echo "Maximum number of retries"
    exit 1
  fi

  if [[ $((counter * delay)) -ge ${timeout} ]]; then
    echo "Timeout"
    exit 1
  fi

  counter=$((counter+1))

  printf '.'
  sleep ${delay}
done
