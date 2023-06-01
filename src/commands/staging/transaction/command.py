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

from commands.staging.transaction.delete import run as run_delete
from commands.staging.transaction.get import run as run_get
from commons.custom_classes import PdpException


@click.group()
@click.pass_context
def transaction(ctx):
  """
  Encloses all commands that let you perform actions on an item on the staging API.
  """


@transaction.command()
@click.pass_obj
@click.option('--bucket', required=True,
              help='The name for the bucket to get the transactions.')
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


@transaction.command()
@click.pass_obj
@click.option('--bucket', required=True,
              help='The name for the bucket to delete the transactions.')
@click.option('-i', '--id', 'transaction_ids', default=[], multiple=True,
              help='Will delete the transaction with the specific id. Default is []. '
                   'The command allows multiple flags of -i.')
@click.option('-a', '--all', '_all', is_flag=True,
              help='Will try to delete all the transactions if the -i flag was not provided. If neither of them are '
                   'provided an error will be raised. Default is False.')
@click.option('--purge', is_flag=True, default=False,
              help='Will delete all the transactions starting from the first transaction until the one specified by the'
                   ' -i flag. Default is False.')
def delete(obj, bucket: str, transaction_ids: list[str], _all: bool, purge: bool):
  """
  Deletes one or more transactions for a given bucket.
  """
  if len(transaction_ids) <= 0 and not _all:
    raise PdpException(message="You must to provide the --all flag if you want to delete all the entities.")

  configuration = obj['configuration']
  run_delete(configuration, bucket, transaction_ids, purge)
