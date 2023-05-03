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

from commons.console import create_spinner, spinner_change_text, spinner_fail, spinner_ok
from commons.constants import CORE, FILES_FOLDER, URL_GENERIC_FILE
from commons.file_system import has_pdp_project_structure
from commons.handlers import handle_and_continue
from commons.http_requests import delete


def delete_file(config: dict, file_name: str, local: bool):
  """
  Deletes a file from the Core API.
  :param dict config: A dictionary containing the url of the products and the path of the project.
  :param str file_name: The file name to delete.
  :param bool local: Deletes the file from your pc too.
  """
  project_path = config['project_path']
  files_folder = os.path.join(project_path, CORE.title(), FILES_FOLDER)
  if not has_pdp_project_structure(project_path):
    files_folder = ""
  file_path = os.path.join(files_folder, file_name)
  name = os.path.split(file_name)
  name = name[len(name) - 1]
  _, res = handle_and_continue(delete, {'show_exception': True}, URL_GENERIC_FILE.format(config[CORE]),
                               params={'name': name})
  if res is None:
    return False

  acknowledged = json.loads(res).get("acknowledged", False)
  if not acknowledged:
    return False

  if local:
    os.remove(file_path)

  return True


def run(config: dict, files: list[str], local: bool):
  """
  Deletes files from the Core API and from local.
  :param dict config: A dictionary containing the url of the products and the path of the project.
  :param str files: A list with the file names to delete.
  :param bool local: Deletes the file from your pc too.
  """
  for file in [*files]:
    create_spinner()
    spinner_change_text(f"Deleting file {click.style(file, fg='cyan')}...")
    if delete_file(config, file, local):
      spinner_ok(f"File {click.style(file, fg='cyan')} was deleted.")
    else:
      spinner_fail(f"File {click.style(file, fg='cyan')} couldn't be deleted.")
