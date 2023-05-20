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

from commons.console import print_console, print_warning
from commons.constants import STAGING, URL_BUCKET_BATCH
from commons.custom_classes import DataInconsistency
from commons.file_system import read_binary_file
from commons.http_requests import post


def input_stage(file_path: str, interactive: bool) -> dict:
  """
  Will return the content for the body request.
  :param str file_path: The path to a file to get the contents to the body of the request.
  :param bool interactive: Will open a text editor to get the contents for the body request.
  """
  body: str | None = None
  if file_path is not None:
    body = read_binary_file(file_path).decode('utf-8')

  if interactive:
    if body is None:
      body = "[\n\n]"

    body = click.edit(body)

  if body is None:
    print_warning("If you wrote something, you must save it before close, otherwise the content won't be added.")
    raise DataInconsistency(message="The body of the batch can not be empty.")

  return json.loads(body)


def run(config: dict, bucket: str, file_path: str, interactive: bool, is_json: bool):
  """
  Performs a list of actions such as ADD and DELETE to a given bucket within the Staging API.
  :param dict config: The configuration containing the pdp products' url.
  :param str bucket: The name to perform the actions.
  :param str file_path: The path to a file to get the contents to the body of the request.
  :param bool interactive: Will open a text editor to get the contents for the body request.
  :param bool is_json: This is a boolean flag. It will print the results in JSON format. Default is False.
  """
  body = input_stage(file_path, interactive)
  res = post(URL_BUCKET_BATCH.format(config[STAGING], bucket=bucket), json=body)

  if res is None:
    return print_console("The batch couldn't be processed.")

  res = json.loads(res)
  if is_json:
    return print_console(res)

  res = json.dumps(res, indent=2)
  print_console(res)
