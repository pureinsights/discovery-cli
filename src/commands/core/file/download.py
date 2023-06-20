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
import os.path

from commons.console import print_console
from commons.constants import CORE, FILES_FOLDER, URL_DOWNLOAD_FILE
from commons.file_system import has_pdp_project_structure, write_binary_file
from commons.handlers import handle_and_exit
from commons.http_requests import get


def run(config: dict, name: str, path: str | None):
  """
  Will try to download a file previously uploaded to the Core API.
  :param dict config: The configuration that contains the product's url.
  :param str name: The name of the file to be downloaded.
  :param str path: The path where the file will be downloaded.
  """
  if path is None:
    path = config['project_path']
    if has_pdp_project_structure(path):
      path = os.path.join(path, CORE.title(), FILES_FOLDER)

  if os.path.isdir(path):
    path = os.path.join(path, name)

  _, data = handle_and_exit(get, {}, URL_DOWNLOAD_FILE.format(config[CORE]), params={'name': name})

  write_binary_file(path, data)
  print_console(f'The file "{name}" has been written correctly.')
