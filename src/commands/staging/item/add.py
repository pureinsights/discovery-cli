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
from tabulate import tabulate

from commons.console import print_console, print_error, print_warning
from commons.constants import STAGING, URL_GENERIC_ITEM
from commons.custom_classes import DataInconsistency
from commons.file_system import read_binary_file
from commons.http_requests import put


def input_stage(path: str | None, interactive: bool):
  """
  Reads the content to store in the item. Could be provided by a file, in an interactive way or both.
  :param str path: The path to read the item's content.\
  :param bool interactive: If true will launch your default text editor to allow you to modify the contents of the item.
  """
  # Messages
  can_not_be_empty = "The content of the item can not be empty."
  save_before_close = "If you wrote something, you must save it before close, otherwise the content won't be added."

  if path is None:
    if not interactive:
      raise DataInconsistency(message=can_not_be_empty)
    content_str: str | None = click.edit('{\n\n}')
    if content_str is None:
      print_warning(save_before_close)
      raise DataInconsistency(message=can_not_be_empty)
    return json.loads(content_str)

  content_str = read_binary_file(path).decode('utf-8')
  if interactive:
    content_str: str | None = click.edit(content_str)
    if content_str is None:
      print_warning(save_before_close)
      raise DataInconsistency(message=can_not_be_empty)

  return json.loads(content_str)


def run(config: dict, bucket: str, item_id: str, content_path: str | None, interactive: bool, parent_id: str | None,
        is_json: bool, verbose: bool):
  """
  Add or updates an item on the staging API.
  :param dict config: The configuration containing the pdp products' url.
  :param str bucket: The name of the bucket to store the item. If the bucket doesn't exist it will be created.
  :param str item_id: The id of the content that will be stored.
  :param str content_path: Is the path where the content of the item is stored. can be None.
  :param bool interactive: If true will launch your default text editor to allow you to modify the contents of the item.
  :param str parent_id: This allows to store the item as child of other item.
  :param bool is_json: This is a boolean flag. It will print the results in JSON format.
  :param bool verbose: Will show more information.
  """
  content = input_stage(content_path, interactive)

  res = put(URL_GENERIC_ITEM.format(config[STAGING], bucket=bucket, content_id=item_id), params={'parentId': parent_id},
            json=content)

  if res is None:
    print_error(f"The item with id {item_id} couldn't be added to {bucket}.", True)

  if is_json:
    return print_console(json.loads(res))

  print_console(click.style("The item was added successfully.", fg='green'))
  if verbose:
    res = json.loads(res)
    headers = res.keys()
    values = [headers, [res[header] for header in headers]]
    print_console(tabulate(values, headers='firstrow'))
