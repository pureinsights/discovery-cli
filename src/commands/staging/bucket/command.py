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
@click.option('-j', '--json', '_json', is_flag=True, default=False,
              help='This is a boolean flag. It will print the results in JSON format. Default is False.')
def get(obj: dict, bucket: str, _json: bool):
  """
  Retrieves all the items for a given bucket.
  """
  configuration = obj['configuration']
  run_get(configuration, bucket, _json)
