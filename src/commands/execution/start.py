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
from commons.constants import INGESTION, URL_SEED_START
from commons.http_requests import post


def run(config: dict, seed: str, scan_type: str):
  """
  Will start the execution for the given seed id.
  :param dict config: The configuration containing the url products.
  :param str seed: The id of the seed to start the execution.
  :param str scan_type: The strategy to use in the scan phase of the execution.
  """
  res = post(URL_SEED_START.format(config[INGESTION], id=seed), params={"scanType": scan_type})
  execution_id = json.loads(res).get("id")
  print_console(f"The execution was started with id {click.style(execution_id, fg='green')}.")
