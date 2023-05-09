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

from commands.core.search import run as run_search


@click.group()
@click.pass_context
def core(ctx):
  """
  Encloses all the commands related with the Core API.
  """


@core.command()
@click.pass_obj
@click.option('-l', '--label', default=None, multiple=True,
              help='Label key or label key and value of the entity. Format: <key>:<value> | <key>. '
                   'Multiple label flags are supported. Default is None.')
@click.option('-t', '--entity-type', 'entity_types', default=None, multiple=True,
              help='Type of the entity. Format <product>:<entityType>. Product values supported Ingestion '
                   'and Discovery. Default is None.')
@click.option('-q', default=None,
              help='The name or description of the entity. Default is None.')
@click.option('-p', '--page', default=0,
              help='The number of the page to show. Min 0. Default is 0.')
@click.option('-s', '--size', default=25,
              help='The size of the page to show. Range 1 - 100. Default is 25.')
@click.option('--asc', default=[], multiple=True,
              help='The name of the property to sort in ascending order. Multiple flags are supported. Default is [].')
@click.option('--desc', default=[], multiple=True,
              help='The name of the property to sort in descending order. Multiple flags are supported. Default is [].')
@click.option('-j', '--json', 'is_json', default=False, is_flag=True,
              help='This is a boolean flag. Will print the results in JSON format. Default is False.')
def search(obj, label: list[str], entity_types: list[str], q: str, page: int, size: int, asc: list[str],
           desc: list[str], is_json: bool):
  """
  Search for entities of all products Ingestion, Core and Discovery. And also, is a group command to chain the 'replace'
  command and replace the entities on the search results.
  """
  configuration = obj['configuration']
  sort = []
  for asc_property in asc:
    sort += [f'{asc_property},asc']
  for desc_property in desc:
    sort += [f'{desc_property},desc']
  query_params = {
    "label": label,
    "type": entity_types,
    "q": q,
    "page": page,
    "size": size,
    "sort": sort
  }
  run_search(configuration, query_params, is_json)
