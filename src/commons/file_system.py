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
import os
from pathlib import Path

from commons.console import print_error, print_warning
from commons.constants import PRODUCTS, STAGING
from commons.custom_classes import PdpException
from commons.handlers import handle_and_exit
from commons.raisers import raise_file_not_found_error


def list_directories(abs_path: str):
  """
  Get a list of names of directories on a given path.
  :param str abs_path: The path to the folder to read.
  :rtype: list[str]
  :return: A list of names of folders.
  """
  raise_file_not_found_error(abs_path)
  return [directory for directory in os.listdir(abs_path) if os.path.isdir(os.path.join(abs_path, directory))]


def list_files(path: str):
  """
  Get a list of names of files on a given path.
  :param str path: The path to the folder to read.
  :rtype: list[str]
  :return: A list of file names.
  """
  raise_file_not_found_error(path)
  return [file for file in os.listdir(path) if not os.path.isdir(os.path.join(path, file))]


def replace_file_extension(file: str, extension: str):
  """
  Remove the file extension of a given file path.
  :param str file: The path to the file to remove the extension.
  :param str extension: The new extension of the file.
  :rtype: str
  :return: The new file path with the file extension replaced.
  """
  return Path(file).with_suffix(extension).name


def read_entities(path: str):
  """
  Reads a file in format json that contains the entities of PDP.
  :param str path: The path to the file.
  :rtype: list[dict]
  :return: A list of entities. If the file contains just one entity it still will be returned as a list.
  :raises FileNotFoundError: If the file does not exist.
  """
  raise_file_not_found_error(path)
  if not os.path.isfile(path):
    raise PdpException(message=f'Path "{path}" is not a file.')
  with open(path, 'r+') as file:
    _, entities = handle_and_exit(json.load,
                                  {
                                    'show_exception': True,
                                    'exception': PdpException(
                                      message=f'JSONDecodeError: Could not parse the file {path}. '
                                              f'Please check the file has a valid JSON format.')
                                  },
                                  file)
    if type(entities) is not list:
      entities = [entities]
    return entities


def write_entities(path: str, entities: list[dict]):
  """
  Write the given entities in json format on a file.
  :param str path: The path to the file where the entities will be written.
  :param list[dict] entities: A list of entities that will be written.
  """
  with open(path, 'r+') as file:
    file.seek(0)
    json.dump(entities, file, indent=2)
    file.truncate()


def has_pdp_project_structure(path: str, show: str = None):
  """
  Validates if a given directory has a pdp project structure.
  :param str path: The path to the directory to validate.
  :param str show: The mode to print messages, warning, error, None. None will no print nothing and
                   error will raise an error.
  :rtype: bool
  :return: True if the folder has the structure correctly, False in other case.
  :raises FileNotFoundError: If path doesn't exist.
  """
  raise_file_not_found_error(path)
  print_aux = lambda message: None
  if show == 'warning':
    print_aux = print_warning
  elif show == 'error':
    print_aux = lambda message: print_error(message, True)

  if not os.path.isdir(path):
    print_aux("The path provided is not a directory.")
    return False

  directories = list_directories(path)
  files = list_files(path)
  has_structure = True
  if 'pdp.ini' not in files:
    print_aux(f"The file pdp.ini is missing on {path}. (Case sensitive).")
    has_structure = False

  for product in PRODUCTS['list']:
    if product == STAGING: continue
    if product.title() not in directories:
      print_aux(f"The folder {product.title()} missing on {path}. (Case sensitive).")
      has_structure = False
      continue
    product_path = os.path.join(path, product.title())
    product_files = list_files(product_path)
    for entity_type in PRODUCTS[product]['entities']:
      if entity_type.associated_file_name not in product_files:
        print_aux(f"The file {entity_type.associated_file_name} missing on {product_path}. (Case sensitive).")
        has_structure = False

  return has_structure


def read_binary_file(path: str):
  """
  Reads a file from the file system as binary data.
  :param str path: The path of the file to read.
  """
  raise_file_not_found_error(path)
  with open(path, mode='rb') as file:
    return file.read()


def write_binary_file(path: str, data: bytes):
  """
  Writes binary data to a file to the file system.
  :param str path: The path where the file will be written.
  :param bytes data: The data to be written in the files.
  """
  with open(path, mode='wb') as file:
    file.write(data)
