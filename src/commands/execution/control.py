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
from commons.constants import INGESTION, URL_SEED_CONTROL
from commons.http_requests import put


def run(config: dict, seed: str, action: str):
  """
  Will control the execution for the given seed id.
  :param dict config: The configuration containing the url products.
  :param str seed: The id of the seed to start the execution.
  :param str action: The action to trigger.
  """
  res = put(URL_SEED_CONTROL.format(config[INGESTION], id=seed), params={"action": action})
  if json.loads(res).get("acknowledged", False):
    print_console(
      f"The seed {click.style(seed, fg='green')} was {click.style(f'{action.upper()}', fg='magenta')} successfully."
    )
    return
  print_console(f"Couldn't {click.style(action.upper(), fg='magenta')} the seed {click.style(seed, fg='green')}")
