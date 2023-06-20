#  Copyright (c) 2023 Pureinsights Technology Ltd. All rights reserved.
#
#  Permission to use, copy, modify or distribute this software and its
#  documentation for any purpose is subject to a licensing agreement with
#  Pureinsights Technology Ltd.
#
#  All information contained within this file is the property of
#  Pureinsights Technology Ltd. The distribution or reproduction of this
#  file or any information contained within is strictly forbidden unless
#  prior written permission has been granted by Pureinsights Technology Ltd.
import json

import click

from commons.console import print_console
from commons.constants import STAGING, URL_BUCKET_STATUS
from commons.http_requests import get


def print_status(bucket: str, status: dict, is_json: bool):
  """
  Prints the content status for a bucket.
  :param str bucket: The name of the bucket to show the status.
  :param dict status: The status of the bucket to show.
  :param bool is_json: This is a boolean flag. It will print the results in JSON format.
  """
  if is_json:
    return print_console(status)

  print_console(f"Status for the bucket {click.style(bucket, fg='cyan')}.")
  return print_console(json.dumps(status, indent=2), suffix='\n')


def run(config: dict, bucket: str, query_params: dict, is_json: bool):
  """
  Retrieves the information of all the items of the given bucket.
  :param dict config: The configuration containing the pdp products' url.
  :param str bucket: The name of the bucket to get the item.
  :param dict query_params: A dict containing the query params for the endpoint.
  :param bool is_json: This is a boolean flag. It will print the results in JSON format.
  """
  content = get(URL_BUCKET_STATUS.format(config[STAGING], bucket=bucket), params=query_params)
  if content is None:
    return print_console(f"Couldn't get the status for the bucket {click.style(bucket, fg='cyan')}.")

  content = json.loads(content)
  if bucket != '' or is_json:
    return print_status(bucket, content, is_json)

  statuses = content.get('content', [])
  for status in statuses:
    print_status(status.get('bucket', ''), status, is_json)
