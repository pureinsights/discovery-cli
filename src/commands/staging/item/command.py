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
import uuid

import click

from commands.staging.item.add import run as run_add
from commands.staging.item.delete import run as run_delete
from commands.staging.item.get import run as run_get
from commons.custom_classes import PdpException


@click.group()
@click.pass_context
def item(ctx):
  """
  Encloses all commands that let you perform actions on an item on the staging API.
  """


@item.command()
@click.pass_obj
@click.option('--bucket', required=True,
              help='The name of the bucket where the item will be added.')
@click.option('-i', '--item-id', 'item_id',
              help='The id of the new item. If no id is provided, then an auto-generated hash will be set. '
                   'Default is None.')
@click.option('--parent', default=None,
              help='This allows you to add an item within an existing item. Default is None.')
@click.option('--interactive', is_flag=True, default=False,
              help='This is a Boolean flag. Will launch your default text editor to allow you to modify the entity '
                   'configuration. Default is False.')
@click.option('--path', '_file', default=None,
              help='The path to the file that contains the content of the item. Default is None.')
@click.option('-v', '--verbose', is_flag=True, default=False,
              help='Will show more information about the item upload. Default is False.')
@click.option('-j', '--json', '_json', is_flag=True, default=False,
              help='This is a boolean flag. It will print the results in JSON format. Default is False.')
def add(obj: dict, bucket: str, item_id: str, parent: str, _file: str, interactive: bool, _json: bool, verbose: bool):
  """
  Adds a new item to a given bucket within the staging API. If the bucket doesn't exist, will be created.
  """
  configuration = obj['configuration']
  if item_id is None:
    item_id = str(uuid.uuid4())
  run_add(configuration, bucket, item_id, _file, interactive, parent, _json, verbose)


@item.command()
@click.pass_obj
@click.option('--bucket', required=True,
              help='The name of the bucket where the item will be added.')
@click.option('-i', '--item-id', 'item_id', required=True, multiple=True,
              help='The id of the item to show. Default is []. The command allows multiple flags of -i.')
@click.option('--content-type', 'content_type', default='CONTENT',
              type=click.Choice(['CONTENT', 'METADATA', 'BOTH'], case_sensitive=False),
              help='The content-type of the query. Default is CONTENT.')
@click.option('-j', '--json', 'is_json', is_flag=True, default=False,
              help='This is a boolean flag. It will print the results in JSON format. Default is False.')
def get(obj: dict, bucket: str, item_id: list[str], content_type: str, is_json: bool):
  """
  Retrieves the information of the given item.
  """
  configuration = obj['configuration']
  run_get(configuration, bucket, item_id, content_type, is_json)


@item.command()
@click.pass_obj
@click.option('--bucket', required=True,
              help='The name of the bucket where the item will be added.')
@click.option('-i', '--item-id', 'item_ids', multiple=True,
              help='The id of the item that you want to delete. Default is []. '
                   'The command allows multiple flags of -i.')
@click.option('-a', '--all', '_all', is_flag=True,
              help='Will try to delete all the items if the -i flag was not provided. If neither of them are provided '
                   'an error will be raised. Default is False.')
@click.option('--filter', 'filter', is_flag=True, default=False,
              help='Will open a text editor to capture the query to filter the data.')
def delete(obj: dict, bucket: str, item_ids: list[str], _all: bool, filter: bool):
  """
  Will delete a given item or all items in case that you don’t provide one or more item ids.
  """
  if len(item_ids) <= 0 and not _all and not filter:
    raise PdpException(message="You must to provide the --all flag if you want to delete all the entities.")

  configuration = obj['configuration']
  run_delete(configuration, bucket, item_ids, filter)
