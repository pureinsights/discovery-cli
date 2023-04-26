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

import pytest

config_file_test_name = "pdp_test.py"
config_file_test_content = '[DEFAULT]' \
                           'ingestion = http://localhost:8080' \
                           'discovery = http://localhost:8088/admin' \
                           'core = http://localhost:8082' \
                           'staging = http://localhost:8081' \
                           '[FAKE]' \
                           'ingestion = http://ingestion-fake' \
                           'discovery = http://discovery-fake/admin' \
                           'core = http://core-fake' \
                           'staging = http://staging-fake'


@pytest.fixture
def mock_path_exists(mocker):
  def _mock_path_exists(ret: bool):
    path_exists = mocker.patch('os.path.exists')
    return_value = False
    path_exists.return_value = ret

  return _mock_path_exists


@pytest.fixture
def test_project_path():
  def joinPath(*args):
    return os.path.join('.', 'test_project', *args)

  return joinPath


@pytest.fixture
def mock_custom_exception():
  def raise_custom_exception(exception):
    if exception is not None:
      raise exception

  return raise_custom_exception
