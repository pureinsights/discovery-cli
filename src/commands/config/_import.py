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
from commons.constants import URL_IMPORT
from commons.custom_classes import DataInconsistency
from commons.file_system import read_binary_file, replace_file_extension
from commons.handlers import handle_and_continue
from commons.http_requests import post
from commons.raisers import raise_file_not_found_error


def print_stage(entities_imported: dict):
  """
  Shows the entities that were imported correctly.
  :param dict entities_imported: The entities imported classified by entity type.
  """
  print_console('Imported entities:')
  for key in entities_imported.keys():
    print_console(f'{click.style(key.title(), fg="cyan")}: ', prefix='  ')
    for entity in entities_imported[key]:
      _id = entity.get("id", None)
      print_console(f'Entity imported with id {click.style(_id.title(), fg="green")}.', prefix='    ')


def run(config: dict, target: str, file_path: str):
  """
  Imports the entities on the giving product.
  :param dict config: The configuration containing the product's url.
  :param str target: The name of the product where the entities will be imported.
  :param str file_path: The path to the .zip that will be imported.
  """
  raise_file_not_found_error(file_path)
  if os.path.isdir(file_path):
    raise DataInconsistency(message=f'The path "{file_path}" is a folder, not a file.', content={"path": file_path})

  binary_data = read_binary_file(file_path)
  handle_configuration = {
    'message': f'Could not import the file "{file_path}" to {target}.',
    'show_exception': True
  }
  split_path = os.path.split(file_path)
  file_name = replace_file_extension(split_path[len(split_path) - 1], '')
  success, result = handle_and_continue(
    post, handle_configuration, URL_IMPORT.format(config[target]),
    files={'file': (file_name, binary_data, 'multipart/form-data')}
  )

  if not success:
    return

  entities_imported = json.loads(result)
  print_stage(entities_imported)
