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

from commands.staging.bucket.get import run as run_get


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
        _json: bool):
  """
  Retrieves all the items for a given bucket. You can filter or use pagination, but just once at time.
  If you provide --token or --content-type will prioritize the filter over pagination.
  """
  query_params = {}
  # This could be '', 'query' or 'filter' and is the last segment of the url URL/content/{bucket}/<query_value>
  # Help to decide which endpoint to call
  query = ''
  if token is not None or content_type is not None:
    query = 'filter'
    query_params = {
      'token': token,
      'contentType': content_type,
      'size': size
    }
  elif page is not None or size is not None or len(asc) > 0 or len(desc) > 0:
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
  run_get(configuration, bucket, query_params, query, _json)
