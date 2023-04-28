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

from commands.core.file.download import run as run_download
from commands.core.file.upload import run as run_upload


@click.group()
@click.pass_context
def file(ctx):
  """
  Encloses all the commands related with files within the Core API.
  """


@file.command()
@click.pass_obj
@click.option('--name', default=None,
              help='The name of the file, if no name is provided, then the name will be the name found in the path.')
@click.option('--path', required=True,
              help='The path where the file is located. If just a name is passed instead of a path to the '
                   'file the cli will try to find the file in the ./Core/files/.')
def upload(obj, path: str, name: str):
  """
  Try to upload the provided file to the Core API.
  """
  configuration = obj['configuration']
  run_upload(configuration, name, path)


@file.command()
@click.pass_obj
@click.option('--name', required=True,
              help='The name of the file to download form Core API.')
@click.option('--path', default=None,
              help='The path where the file will be written. Default is ./Core/files/ if you are in a PDP project, '
                   'if not, default is ./.')
def download(obj, name: str, path: str):
  """
  Will try to download a file previously uploaded to the Core API.
  """
  configuration = obj['configuration']
  configuration['project_path'] = obj['project_path']
  run_download(configuration, name, path)
