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
from commons.constants import INGESTION, URL_SEED_RESTART
from commons.http_requests import post


def run(config: dict, seed: str):
  """
  Will reset the given seed.
  :param dict config: The configuration containing the url products.
  :param str seed: The id of the seed to reset.
  """
  res = post(URL_SEED_RESTART.format(config[INGESTION], id=seed))
  if json.loads(res).get("acknowledged", False):
    print_console(f"The seed {click.style(seed, fg='green')} was reset successfully.")
    return
  print_console(f"Couldn't reset the seed {click.style(seed, fg='green')}")
