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
from tabulate import tabulate

from commons.console import print_console
from commons.constants import INGESTION, URL_SEED_EXECUTION
from commons.handlers import handle_and_continue
from commons.http_requests import get


def get_all_executions(config: dict, seed: str, query_params: dict) -> list[dict]:
  """
  Returns all the executions of a seed.
  :param dict config: The configuration containing the product's url.
  :param str seed: The seed to get the executions.
  :param dict query_params: Query params used when no ids where passed.
  """
  res = get(URL_SEED_EXECUTION.format(config[INGESTION], id=seed, execution=''), params=query_params)
  if res is None:
    return []
  return json.loads(res).get('content', [])


def get_executions_by_ids(config: dict, seed: str, ids: list[str]):
  """
  Returns the information of the executions by ids.
  :param dict config: The configuration containing the product's url.
  :param str seed: The seed to get the executions.
  :param list[str] ids: A list of ids of executions to get information.
  """
  executions = []
  for execution_id in ids:
    _, execution = handle_and_continue(get_execution_by_id, {}, config, seed, execution_id)
    if execution is not None:
      executions += [execution]
  return executions


def get_execution_by_id(config: dict, seed: str, execution: str):
  """
  Returns the information of the execution by id.
  :param dict config: The configuration containing the product's url.
  :param str seed: The seed to get the executions.
  :param str execution: The id of the execution to search.
  """
  res = get(URL_SEED_EXECUTION.format(config[INGESTION], id=seed, execution=execution))
  return json.loads(res)


def print_stage(seed: str, executions: list[dict], given_executions: list[str], is_json: bool, pretty: bool,
                is_verbose: bool):
  """
  Shows the information from active executions of a given seed.
  :param str seed: The seed to get the executions.
  :param list[dict] executions: A list of executions to show the information.
  :param list[str] given_executions: A list of executions ids given by the user to search for.
  :param bool is_json: Will show the executions in JSON format.
  :param bool pretty: If True, the result will be showed in  a human-readable JSON format.
  :param bool is_verbose: Will show more information about the executions.
  """
  if pretty:
    return print_console(json.dumps(executions, indent=2))

  if is_json:
    return print_console(executions)

  if len(executions) <= 0:
    if len(given_executions) <= 0:
      return print_console(f"The seed {seed} doesn't have executions.")

    return print_console(f"The seed {seed} doesn't have the executions: {','.join(given_executions)}")

  if is_verbose:
    return print_executions_verbose(executions, seed)

  executions_str = ""
  for index, execution in enumerate(executions):
    executions_str += (', ' if index != 0 else '') + execution.get('id', '')

  print_console(f"Executions of seed {seed}: {executions_str}")


def print_executions_verbose(executions: list[dict], seed: str):
  """
  Shows the information from active executions of a given seed, shows more information as tables.
  :param str seed: The seed to get the executions.
  :param list[dict] executions: A list of executions to show the information.
  """
  print_console(f"Executions of seed {click.style(seed, fg='green')}: ")
  headers = ['id', 'status', 'pipelineId', 'jobId', 'steps']
  table_values = [headers]
  for execution in executions:
    values = []
    for header in headers:
      if header == 'steps':
        values += [len(execution.get(header, '---'))]
      else:
        values += [execution.get(header, '---')]
    table_values += [values]
  table = tabulate(table_values, headers='firstrow', showindex='always', tablefmt='presto', missingval='---')
  print_console(table)


def run(config: dict, seed: str, execution: list[str], is_json: bool, pretty: bool, is_verbose: bool,
        query_params: dict):
  """
  Retrieves the information from active executions of a given seed.
  :param dict config: The configuration containing the product's url.
  :param str seed: The seed to get the executions.
  :param list[str] execution: A list of ids of executions to get information.
  :param bool is_json: Will show the executions in JSON format.
  :param bool pretty: If True, the result will be showed in  a human-readable JSON format.
  :param bool is_verbose: Will show more information about the executions.
  :param dict query_params: Query params used when no ids where passed.
  """
  executions = []
  if len(execution) > 0:
    executions = get_executions_by_ids(config, seed, execution)
  else:
    executions = get_all_executions(config, seed, query_params)

  print_stage(seed, executions, execution, is_json, pretty, is_verbose)
