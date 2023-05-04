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

from commands.execution.control import run as run_control
from commands.execution.reset import run as run_reset
from commands.execution.start import run as run_start


@click.group("seed-exec")
@click.pass_context
def seed_exec(ctx):
  """
  Encloses all the commands related with the seed execution.
  """


@seed_exec.command()
@click.pass_obj
@click.option('--seed', required=True,
              help='The id of the seed to start the scanning process.')
@click.option('--scan-type', 'scan_type', default='FULL',
              type=click.Choice(['FULL', 'INCREMENTAL'], case_sensitive=False),
              help='The strategy to apply during the scan phase. '
                   'Values supported INCREMENTAL and FULL. Default is FULL.')
def start(obj, seed, scan_type):
  """
  Try to start the scanning process of a seed. Note that a seed can only have one active execution at a time.
  """
  configuration = obj['configuration']
  run_start(configuration, seed, scan_type)


@seed_exec.command()
@click.pass_obj
@click.option('--seed', required=True,
              help='The id of the seed to reset the associated data.')
def reset(obj, seed):
  """
  Reset all the associated data of the given seed.
  """
  configuration = obj['configuration']
  run_reset(configuration, seed)


@seed_exec.command()
@click.pass_obj
@click.option('--seed', required=True,
              help='The id of the seed to trigger the action.')
@click.option('--action', default='HALT',
              type=click.Choice(['HALT', 'PAUSE', 'RESUME'], case_sensitive=False),
              help='The action you want to trigger. Values supported HALT, PAUSE and RESUME. Default is HALT.')
def control(obj, seed, action):
  """
  Triggers and action on all active executions for the given seed.
  """
  configuration = obj['configuration']
  run_control(configuration, seed, action)
