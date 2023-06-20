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

from commons.console import print_console, print_error
from commons.constants import CORE, URL_LOG_LEVEL
from commons.http_requests import post


def run(config: dict, component: str, level: str, logger_name: str):
  """
  Change the logging level of a component.
  :param dict config: The configuration containing the url products.
  :param str component: The name of the component that you want to change the log level.
  :param str level: The level log you want to change to.
  :param str logger_name: The name of the logger.
  """
  res = post(
    URL_LOG_LEVEL.format(config[CORE]),
    params={'componentName': component, 'level': level, 'loggerName': logger_name}
  )
  if res is None or not json.loads(res).get('acknowledged', False):
    return print_error(f'Could not change the level log of "{component}".')

  print_console(
    f"The level log of \"{click.style(component, fg='green')}\" was changed to {click.style(level, fg='blue')}."
  )
