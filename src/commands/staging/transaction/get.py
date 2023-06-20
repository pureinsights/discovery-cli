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
from commons.constants import STAGING, URL_GENERIC_TRANSACTION
from commons.http_requests import get


def run(config: dict, bucket: str, query_params: dict, is_json: bool):
  """
  Retrieves all the transactions for a given bucket.
  :param dict config: The configuration containing the pdp products' url.
  :param str bucket: The name of the bucket to get the item.
  :param str query_params: The query_params to get the transaction.
  :param bool is_json: This is a boolean flag. It will print the results in JSON format.
  """
  res = get(URL_GENERIC_TRANSACTION.format(config[STAGING], bucket=bucket), params=query_params)
  contents = json.loads(res)

  if is_json:
    return print_console(contents)

  print_console("Items content: ")
  print_console(json.dumps(contents, indent=2))
