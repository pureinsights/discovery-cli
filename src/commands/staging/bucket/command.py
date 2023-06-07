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

from commands.staging.bucket.batch import run as run_batch
from commands.staging.bucket.delete import run as run_delete
from commands.staging.bucket.get import run as run_get
from commands.staging.bucket.status import run as run_status
from commons.console import print_warning
from commons.custom_classes import DataInconsistency, PdpException
from commons.handlers import handle_and_exit


@click.group('bucket')
@click.pass_context
def bucket_command(ctx):
  """
  Encloses all commands that let you perform actions on a bucket on the staging API.
  """


@bucket_command.command()
@click.pass_obj
@click.option('--bucket', required=True,
              help='The name for the bucket to get the items.')
@click.option('--token', default=None,
              help='The token of the contents you want to filter.')
@click.option('--content-type', 'content_type', default=None,
              type=click.Choice(['CONTENT', 'METADATA', 'BOTH'], case_sensitive=False),
              help='The content-type of the query. Default is CONTENT.')
@click.option('--filter', 'filter', is_flag=True, default=False,
              help='Will open a text editor to capture the query to filter the data.')
@click.option('--page', default=None,
              help='The number of the page to query.')
@click.option('--size', default=None,
              help='The size of the page to query.')
@click.option('--asc', default=[], multiple=True,
              help='The name of the property to sort in ascending order. Multiple flags are supported. Default is [].')
@click.option('--desc', default=[], multiple=True,
              help='The name of the property to sort in descending order. Multiple flags are supported. Default is [].')
@click.option('-j', '--json', '_json', is_flag=True, default=False,
              help='This is a boolean flag. It will print the results in JSON format. Default is False.')
def get(obj: dict, bucket: str, token: str, content_type: str, page: int, size: int, asc: list[str], desc: list[str],
        _json: bool, filter: bool):
  """
  Retrieves all the items for a given bucket. You can filter or use pagination, but just once at time.
  If you provide --token or --content-type will prioritize the filter over pagination.
  """
  query_params = {
    'token': token,
    'contentType': content_type,
    'size': size
  }
  filter_body = {}
  # This could be '', 'query' or 'filter' and is the last segment of the url URL/content/{bucket}/<query_value>
  # Help to decide which endpoint to call
  query = ''
  if filter:
    query = 'filter'
    filter_criteria: str | None = click.edit('{\n\n}')
    if filter_criteria is None:
      print_warning('If you wrote some criteria, please be sure you save before close the text editor.')
      raise PdpException(message="The filter criteria can't be empty.")
    _, filter_body = handle_and_exit(
      json.loads,
      {'message': "The filter criteria entered doesn't have a valid JSON format."},
      filter_criteria
    )
  elif page is not None or len(asc) > 0 or len(desc) > 0:
    sort = []
    for asc_property in asc:
      sort += [f'{asc_property},asc']
    for desc_property in desc:
      sort += [f'{desc_property},desc']
    query = 'query'
    query_params = {
      "page": page,
      "size": size,
      "sort": sort
    }

  configuration = obj['configuration']
  run_get(configuration, bucket, query_params, filter_body, query, _json)


@bucket_command.command()
@click.pass_obj
@click.option('--bucket', required=True,
              help='The name for the bucket to get the items.')
@click.option('--path', 'file', default=None,
              help='The path to the file that contains the body for the query on a JSON format.')
@click.option('--interactive', default=False, is_flag=True,
              help='Will open a text editor to let you write the body for the request.')
@click.option('-j', '--json', 'is_json', is_flag=True, default=False,
              help='This is a boolean flag. It will print the results in JSON format. Default is False.')
def batch(obj, bucket: str, file: str, interactive: bool, is_json: bool):
  """
  Performs a list of actions such as ADD and DELETE to a given bucket within the Staging API.
  """
  if file is None and not interactive:
    raise DataInconsistency(message='You must to provide the --path or --interactive flag.')

  configuration = obj['configuration']
  run_batch(configuration, bucket, file, interactive, is_json)


@bucket_command.command()
@click.pass_obj
@click.option('--bucket', 'buckets', required=True, multiple=True, default=[],
              help='The name of the bucket to delete. If this flag is not provided the flag --all must be provided, '
                   'otherwise an error will be raised. The command allows multiple flags of --bucket. Default is [].')
def delete(obj, buckets: list[str]):
  """
  Will delete the given bucket from the Staging API.
  """
  configuration = obj['configuration']
  run_delete(configuration, buckets)


@bucket_command.command()
@click.pass_obj
@click.option('--bucket', default='',
              help='The name of the bucket to get the status. Default is None.')
@click.option('-p', '--page', 'page', default=0, type=int,
              help='The number of the page to show. Min 0. Default is 0.')
@click.option('-s', '--size', 'size', default=25, type=int,
              help='The size of the page to show. Range 1 - 100. Default is 25.')
@click.option('--asc', default=[], multiple=True,
              help='The name of the property to sort in ascending order. Multiple flags are supported. Default is [].')
@click.option('--desc', default=[], multiple=True,
              help='The name of the property to sort in descending order. Multiple flags are supported. Default is [].')
@click.option('-j', '--json', 'is_json', is_flag=True, default=False,
              help="This is a boolean flag. Will print the results in JSON format. Default is False.")
def status(obj, bucket: str, page: int, size: int, asc: list[str], desc: list[str], is_json: bool):
  """
  Retrieves the status for all the buckets or a given bucket by the user.
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
  run_status(configuration, bucket, query_params, is_json)
