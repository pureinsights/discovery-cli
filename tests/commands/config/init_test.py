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
import pytest

from commands.config.init import create_project_from_template, run
from commons.constants import DEFAULT_CONFIG


def test_run_init_created_successfully(mocker, mock_path_exists):
  mock_path_exists(False)
  mocker.patch('commands.config.init.open')
  mocker.patch('commands.config.init.create_project_from_template', return_value=True)
  mocker_configuration_write = mocker.patch('commands.config.init.configparser.RawConfigParser.write')
  project_name = 'my-pdp-project'
  success = run(project_name, True, DEFAULT_CONFIG, False)
  mocker_configuration_write.assert_called_once()
  assert success


def test_run_init_failed(mocker, mock_path_exists):
  mock_path_exists(False)
  mocker.patch('commands.config.init.open')
  mocker.patch('commands.config.init.create_project_from_template', return_value=False)
  project_name = 'my-pdp-project'
  success = run(project_name, True, DEFAULT_CONFIG, False)
  assert not success


def test_run_init_project_already_exists(mock_path_exists):
  mock_path_exists(True)
  project_name = 'my-pdp-project'
  success = False
  with pytest.raises(Exception) as exception:
    success = run(project_name, False, DEFAULT_CONFIG, False)
  assert 'Project {0} already exists.'.format(project_name) in str(exception.value)
  assert not success


def test_run_init_project_already_exists_and_cant_force(mocker, mock_path_exists):
  mock_path_exists(True)
  mocker.patch('commands.config.init.shutil.rmtree', side_effect=Exception)
  project_name = 'my-pdp-project'
  success = False
  with pytest.raises(Exception) as exception:
    success = run(project_name, False, DEFAULT_CONFIG, True)
  assert 'Can not remove {project_name}.'.format(project_name=project_name.title()) in str(exception.value)
  assert not success


def test_create_project_from_template_project_already_exists(mock_path_exists):
  mock_path_exists(True)
  project_name = 'my-pdp-project'
  success = False
  with pytest.raises(Exception) as exception:
    success = create_project_from_template(project_name)
  assert 'Project {0} already exists.'.format(project_name) in str(exception.value)
  assert not success


# TODO: Comment more this test
def test_create_project_from_template_project_successfully(mocker, mock_path_exists):
  """
  with default values.
  """
  mock_path_exists(True)
  mocker.patch('commands.config.init.os.path.abspath', return_value='')
  mocker.patch('commands.config.init.os.path.join', return_value='')
  mocker.patch('commands.config.init.os.path.dirname', return_value='')
  mocker.patch('commands.config.init.shutil.copytree')
  project_name = 'my-pdp-project'
  success = False
  with pytest.raises(Exception) as exception:
    success = create_project_from_template(project_name)
  assert 'Project {0} already exists.'.format(project_name) in str(exception.value)
  assert not success
