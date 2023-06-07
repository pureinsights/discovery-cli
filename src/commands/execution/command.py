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
from commands.execution.get import run as run_get
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
@click.option('-i', '--seed-id', 'seed', required=True,
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
@click.option('-i', '--seed-id', 'seed', required=True,
              help='The id of the seed to reset the associated data.')
def reset(obj, seed):
  """
  Reset all the associated data of the given seed.
  """
  configuration = obj['configuration']
  run_reset(configuration, seed)


@seed_exec.command()
@click.pass_obj
@click.option('-i', '--seed-id', 'seed', required=True,
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


@seed_exec.command()
@click.pass_obj
@click.option('--seed-id', 'seed', required=True,
              help='The id of the seed you want to get the active executions.')
@click.option('--execution-id', 'executions', default=[], multiple=True,
              help='The id of the execution you want to get information. Default is None. The command allows '
                   'multiple flags of --execution.')
@click.option('-j', '--json', 'is_json', is_flag=True, default=False,
              help='This is a boolean flag. It will print the results in JSON format. Default is False.')
@click.option('-v', '--verbose', 'is_verbose', is_flag=True, default=False,
              help='It will show more information about the deploy results. Default is False.')
@click.option('-p', '--page', 'page', default=0, type=int,
              help='The number of the page to show. Min 0. Default is 0.')
@click.option('-s', '--size', 'size', default=25, type=int,
              help='The size of the page to show. Range 1 - 100. Default is 25.')
@click.option('--asc', default=[], multiple=True,
              help='The name of the property to sort in ascending order. Multiple flags are supported. Default is [].')
@click.option('--desc', default=[], multiple=True,
              help='The name of the property to sort in descending order. Multiple flags are supported. Default is [].')
def get(obj, seed, executions: list[str], is_json: bool, is_verbose: bool, page: int, size: int, asc: list[str],
        desc: list[str]):
  """
  Retrieves the executions of a given seed.
  """
  configuration = obj['configuration']
  sort = []
  for asc_property in asc:
    sort += [f'{asc_property},asc']
  for desc_property in desc:
    sort += [f'{desc_property},desc']
  query_params = {
    "page": page,
    "size": size,
    "sort": sort
  }
  run_get(configuration, seed, executions, is_json, is_verbose and not is_json, query_params)
