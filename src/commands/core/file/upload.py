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
import os.path

import click

from commons.console import print_console
from commons.constants import CORE, URL_GENERIC_FILE
from commons.file_system import read_binary_file
from commons.handlers import handle_and_continue
from commons.http_requests import put


def run(config: dict, name: str, file_path: str):
  """
  Try to upload a file to the files within the Core API.
  :param dict config: The configuration that contains the product's url.
  :param str name: The name of the file to be uploaded.
  :param str file_path: The path to the file which will be uploaded.
  """
  binary_data = read_binary_file(file_path)
  file_name = name
  if name is None:
    split_path = os.path.split(file_path)
    file_name = split_path[len(split_path) - 1]

  handle_configuration = {
    'message': f'Could not upload the file "{file_path}".',
    'show_exception': True
  }
  _, acknowledged = handle_and_continue(
    put, handle_configuration, URL_GENERIC_FILE.format(config[CORE]),
    params={"name": file_name}, files={'file': (file_name, binary_data, 'multipart/form-data')}
  )
  if acknowledged is None:
    return

  if not json.loads(acknowledged).get('acknowledged'):
    print_console(f'Could not upload the file "{file_path}".')
    return

  print_console(f'The file "{click.style(file_path, fg="cyan")}" was uploaded successfully '
                f'as "{click.style(file_name, fg="blue")}".')
