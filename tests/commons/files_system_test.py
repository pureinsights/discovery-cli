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

from commons.file_system import list_files, read_entities, write_entities


def test_list_files(test_project_path):
  """
  Test the function defined in :func:`commons.files_system.list_files`.
  """
  path = test_project_path()
  files = list_files(path)
  assert len(files) == 1
  assert files == ['pdp.ini']


def test_read_entities(test_project_path, mocker):
  """
  Test the function defined in :func:`commons.files_system.read_entities`.
  """
  mocker.patch('builtins.open', mock_open(read_data='[{"id":"fake-id","name":"fake-name"}]'))
  path = test_project_path('Ingestion')
  files = read_entities(path)
  assert len(files) == 1
  assert files == [{"id": "fake-id", "name": "fake-name"}]


def test_read_entities_with_just_one_entity(test_project_path, mocker):
  """
  Test the function defined in :func:`commons.files_system.read_entities`,
  when the file do not contain a list but an entity.
  """
  mocker.patch('builtins.open', mock_open(read_data='{"id":"fake-id","name":"fake-name"}'))
  path = test_project_path('Ingestion')
  files = read_entities(path)
  assert len(files) == 1
  assert files == [{"id": "fake-id", "name": "fake-name"}]


def test_write_entities(test_project_path, mocker):
  """
  Test the function defined in :func:`commons.files_system.write_entities`.
  """
  json_dump_mock = mocker.patch("commons.file_system.json.dump")
  path = test_project_path('Ingestion', 'cron_jobs.json')
  entities = [{'fake-property': 'fake-value'}]
  write_entities(path, entities)
  json_dump_mock.assert_called_once()
