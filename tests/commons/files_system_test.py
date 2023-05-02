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
from unittest.mock import mock_open

import pytest

from commons.custom_classes import PdpException
from commons.file_system import has_pdp_project_structure, list_files, read_binary_file, read_entities, write_entities


def test_list_files(test_project_path):
  """
  Test the function defined in :func:`commons.files_system.list_files`.
  """
  path = test_project_path()
  files = list_files(path)
  assert len(files) == 2
  assert files == ['custom_pipeline.json', 'pdp.ini']


def test_read_entities(test_project_path, mocker):
  """
  Test the function defined in :func:`commons.files_system.read_entities`.
  """
  mocker.patch('builtins.open', mock_open(read_data='[{"id":"fake-id","name":"fake-name"}]'))
  path = test_project_path('Ingestion', 'processors.json')
  files = read_entities(path)
  assert len(files) == 1
  assert files == [{"id": "fake-id", "name": "fake-name"}]


def test_read_entities_with_just_one_entity(test_project_path, mocker):
  """
  Test the function defined in :func:`commons.files_system.read_entities`,
  when the file do not contain a list but an entity.
  """
  mocker.patch('builtins.open', mock_open(read_data='{"id":"fake-id","name":"fake-name"}'))
  path = test_project_path('Ingestion', 'processors.json')
  files = read_entities(path)
  assert len(files) == 1
  assert files == [{"id": "fake-id", "name": "fake-name"}]


def test_read_entities_not_a_file(test_project_path, mocker):
  """
  Test the function defined in :func:`commons.files_system.read_entities`,
  when the given path is not a file.
  """
  mocker.patch('commons.file_system.os.path.isfile', return_value=False)
  path = test_project_path('Ingestion', 'processors.json')
  with pytest.raises(PdpException) as exception:
    read_entities(path)
  assert exception.value.message == f'Path "{path}" is not a file.'


def test_write_entities(test_project_path, mocker):
  """
  Test the function defined in :func:`commons.files_system.write_entities`.
  """
  mocker.patch('builtins.open', mock_open(read_data=''))
  json_dump_mock = mocker.patch("commons.file_system.json.dump")
  path = test_project_path('Ingestion', 'cron_jobs.json')
  entities = [{'fake-property': 'fake-value'}]
  write_entities(path, entities)
  json_dump_mock.assert_called_once()


def test_has_pdp_project_structure_successfully(mocker, test_project_path):
  """
  Test the function defined in :func:`commons.files_system.has_pdp_project_structure`.
  """
  folder = test_project_path()
  has_structure = has_pdp_project_structure(folder)
  assert has_structure


def test_has_pdp_project_structure_no_print(mocker, test_project_path):
  """
  Test the function defined in :func:`commons.files_system.has_pdp_project_structure`,
  without print anything and the give path is not a folder.
  """
  mocker.patch("commons.file_system.os.path.isdir", return_value=False)
  folder = test_project_path()
  has_structure = has_pdp_project_structure(folder)
  assert not has_structure


def test_has_pdp_project_structure_not_a_folder(mocker, test_project_path):
  """
  Test the function defined in :func:`commons.files_system.has_pdp_project_structure`,
  printing warnings and the give path is not a folder.
  """
  mocker.patch("commons.file_system.os.path.isdir", return_value=False)
  print_warning_mock = mocker.patch("commons.file_system.print_warning")
  folder = test_project_path()
  has_structure = has_pdp_project_structure(folder, 'warning')
  assert not has_structure
  print_warning_mock.assert_called_once_with("The path provided is not a directory.")


def test_has_pdp_project_structure_missing_files(mocker, test_project_path):
  """
  Test the function defined in :func:`commons.files_system.has_pdp_project_structure`,
  printing errors and when there are missing files.
  """
  mocker.patch("commons.file_system.list_files", return_value=[])
  print_error_mock = mocker.patch("commons.file_system.print_error")
  path = test_project_path()
  has_structure = has_pdp_project_structure(path, 'error')
  assert not has_structure
  print_error_mock.assert_called()


def test_has_pdp_project_structure_missing_folder(mocker, test_project_path):
  """
  Test the function defined in :func:`commons.files_system.has_pdp_project_structure`,
  when an expected folder is not found.
  """
  mocker.patch("commons.file_system.list_directories", return_value=[])
  print_error_mock = mocker.patch("commons.file_system.print_error")
  path = test_project_path()
  has_structure = has_pdp_project_structure(path, 'error')
  assert not has_structure
  print_error_mock.assert_called()


def test_read_binary_file(mocker):
  """
  Test the function defined in :func:`commons.files_system.read_binary_file`.
  """
  mocker.patch('commons.file_system.open', mock_open(read_data=b"fakedata"))
  mocker.patch("commons.file_system.raise_file_not_found_error")
  bytes_read = read_binary_file('fake_path')
  assert bytes_read == b"fakedata"
