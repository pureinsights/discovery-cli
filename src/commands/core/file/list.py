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
from commons.constants import CORE, URL_GENERIC_FILE
from commons.http_requests import get


def run(config: dict, is_json: bool):
  """
  Show all the files on the Core API.
  :param dict config: The configuration containing the url of the products.
  :param bool is_json: Will show the entities in JSON format.
  """
  files = get(URL_GENERIC_FILE.format(config[CORE]))
  files = json.loads(files)
  
  if is_json:
    print_console(files)
    return

  if len(files) <= 0:
    print_console("There are not files on the Core API.")
    return

  print_console("Files: ")
  for file in files:
    print_console(file, prefix='  ')
