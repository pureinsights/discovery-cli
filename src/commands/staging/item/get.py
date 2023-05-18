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

from commons.console import print_console
from commons.constants import STAGING, URL_GENERIC_ITEM
from commons.handlers import handle_and_continue
from commons.http_requests import get


def run(config: dict, bucket: str, item_ids: list[str], content_type: str, is_json: bool):
  """
  Retrieves the information of the given item.
  :param dict config: The configuration containing the pdp products' url.
  :param str bucket: The name of the bucket to get the item.
  :param str item_ids: The list of ids of the items that will be shown.
  :param str content_type: Defines which data do you want to get. CONTENT, METADATA, BOTH.
  :param bool is_json: This is a boolean flag. It will print the results in JSON format.
  """
  contents = {}
  for item_id in item_ids:
    _, res = handle_and_continue(
      get, {'show_exception': True},
      URL_GENERIC_ITEM.format(config[STAGING], bucket=bucket, content_id=item_id),
      params={'contentType': content_type}
    )
    if res is None:
      continue
    contents[item_id] = json.loads(res)

  if is_json:
    return print_console(contents)

  if len(contents.keys()) <= 0:
    return print_console("No items to show.")
  
  print_console("Items content: ")
  print_console(json.dumps(contents, indent=2))
