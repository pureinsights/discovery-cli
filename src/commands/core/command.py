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

import click

from commands.core.log_level import run as run_log_level
from commons.constants import LOGGER_LEVELS


@click.group()
@click.pass_context
def core(ctx):
  """
  Encloses all the commands related with the Core API.
  """


@core.command('log-level')
@click.pass_obj
@click.option('--component', required=True,
              help='The name of the component that you want to change the log level.')
@click.option('--level', required=True, type=click.Choice(LOGGER_LEVELS,
                                                          case_sensitive=False),
              help='The level log you want to change to. Values supported ERROR,WARN, INFO,DEBUG and TRACE.')
@click.option('--logger',
              help='The of the logger. Default is None.')
def log_level(obj, component: str, level: str, logger: str):
  """
  Change the logging level of a component.
  """
  configuration = obj['configuration']
  run_log_level(configuration, component, level, logger)
