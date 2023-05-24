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
from commons.constants import STAGING, URL_GENERIC_BUCKET
from commons.http_requests import get, post


def run(config: dict, bucket: str, query_params: dict, filter_body: dict, query: str, is_json: bool):
  """
  Retrieves the information of all the items of the given bucket.
  :param dict config: The configuration containing the pdp products' url.
  :param dict query_params: A dict containing the query params for the endpoint.
  :param dict filter_body: A dict containing the filter criteria.
  :param str query: A str that determine to which endpoint will query the data.
                    If is distinct to '' will use URL/content/{bucket}/{query}
  :param str bucket: The name of the bucket to get the item.
  :param bool is_json: This is a boolean flag. It will print the results in JSON format.
  """
  items = {}
  if query != '':
    items = post(f'{URL_GENERIC_BUCKET.format(config[STAGING], bucket=bucket)}/{query}', params=query_params,
                 json=filter_body)
  else:
    items = get(URL_GENERIC_BUCKET.format(config[STAGING], bucket=bucket), params=query_params)

  items = json.loads(items)

  if is_json:
    return print_console(items)

  print_console(json.dumps(items, indent=2))
