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
@click.option('--item-id', 'item_id',
              help='The id of the new item. If no id is provided, then an auto-generated hash will be set. '
                   'Default is None.')
@click.option('--parent', default=None,
              help='This allows you to add an item within an existing item. Default is None.')
@click.option('--interactive', is_flag=True, default=False,
              help='This is a Boolean flag. Will launch your default text editor to allow you to modify the entity '
                   'configuration. Default is False.')
@click.option('--file', '_file', default=None,
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
