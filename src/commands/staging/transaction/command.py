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

from commands.staging.transaction.get import run as run_get


@click.group()
@click.pass_context
def transaction(ctx):
  """
  Encloses all commands that let you perform actions on an item on the staging API.
  """


@transaction.command()
@click.pass_obj
@click.option('--bucket', required=True,
              help='The name for the bucket to get the items.')
@click.option('-i', '--id', '_id', default=None,
              help='Will retrieve all the transactions after the given one. Useful to make pagination. '
                   'Default is None.')
@click.option('--size', default=100,
              help='The number of transactions to fetch. Min 1, Max 1000. Default is 100.')
@click.option('-j', '--json', '_json', is_flag=True, default=False,
              help='This is a boolean flag. It will print the results in JSON format. Default is False.')
def get(obj, bucket: str, _id: str, size: int, _json: bool):
  """
  Retrieves all the transactions for a given bucket.
  """
  configuration = obj['configuration']
  query_params = {
    'transactionId': _id,
    'size': size
  }
  run_get(configuration, bucket, query_params, _json)
