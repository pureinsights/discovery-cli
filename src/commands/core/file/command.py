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

from commands.core.file.delete import run as run_delete
from commands.core.file.download import run as run_download
from commands.core.file.list import run as run_ls
from commands.core.file.upload import run as run_upload


@click.group()
@click.pass_context
def file(ctx):
  """
  Encloses all the commands related with files within the Core API.
  """


@file.command()
@click.pass_obj
@click.option('-n', '--name', default=None,
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
@click.option('-n', '--name', required=True,
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


@file.command()
@click.pass_obj
@click.option('-n', '--name', 'names', required=True, multiple=True,
              help='The name of the file you want to delete. You can provide a full path too to use it with the --local'
                   ' flag. The command allows multiple flags of -n.')
@click.option('--local', default=False, is_flag=True,
              help='This is a boolean flag, it will try to delete the file from your pc too. It will use the path '
                   'provided by the flag name, if just a name was passed and not a path it will search for the file on '
                   'the ./core/files. Default is False.')
def delete(obj, names: list[str], local: bool):
  """
  Will delete the files from the Core API.
  """
  configuration = obj['configuration']
  configuration['project_path'] = obj['project_path']
  run_delete(configuration, names, local)


@file.command()
@click.pass_obj
@click.option('--json', 'is_json', default=False, is_flag=True,
              help='This is a boolean flag. Will print the results in JSON format. Default is False.')
@click.option('--pretty', is_flag=True, default=False,
              help='This is a Boolean flag. Will print the results in human readable JSON format. Default is False.')
def ls(obj, is_json: bool, pretty: bool):
  """
  Show the list of files uploaded to the Core API.
  """
  configuration = obj['configuration']
  run_ls(configuration, is_json, pretty)
