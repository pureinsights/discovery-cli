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


def read_entities(path: str):
  """
  Reads a file in format json that contains the entities of PDP.
  :param str path: The path to the file.
  :rtype: list[dict]
  :return: A list of entities. If the file contains just one entity it still will be returned as a list.
  """
  raise_file_not_found_error(path)
  with open(path, 'r+') as file:
    entities = json.load(file)
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
