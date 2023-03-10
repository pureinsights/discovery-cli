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

from commands.config.init import create_project_folder_structure, create_project_from_existing_sources, \
  create_project_from_template, run
from commons.constants import DEFAULT_CONFIG


def test_run_init_created_successfully(mocker, mock_path_exists):
  """
  Test the command defined in :func:`commands.config.init.run`.
  """
  mock_path_exists(False)
  mocker.patch('commands.config.init.open')
  mocker.patch('commands.config.init.create_project_from_template', return_value=True)
  mocker_configuration_write = mocker.patch('commands.config.init.configparser.RawConfigParser.write')
  project_name = 'my-pdp-project'
  success = run(project_name, DEFAULT_CONFIG, False, 'empty')
  mocker_configuration_write.assert_called_once()
  assert success


def test_run_init_failed(mocker, mock_path_exists):
  """
  Test the command defined in :func:`commands.config.init.run`,
  when can not create the project from a template.
  """
  mock_path_exists(False)
  mocker.patch('commands.config.init.open')
  mocker.patch('commands.config.init.create_project_from_template', return_value=False)
  mocker_configuration_write = mocker.patch('commands.config.init.configparser.RawConfigParser.write')
  project_name = 'my-pdp-project'
  success = run(project_name, DEFAULT_CONFIG, False)
  assert mocker_configuration_write.call_count == 0
  assert not success


def test_run_init_project_already_exists(mocker, mock_path_exists):
  """
  Test the command defined in :func:`commands.config.init.run`,
  when the project name already exists.
  """
  mock_print_error = mocker.patch('commons.console.print_error')
  mock_path_exists(True)
  project_name = 'my-pdp-project'
  success = run(project_name, DEFAULT_CONFIG, False)
  mock_print_error.assert_called_once_with(
    'Project {0} already exists.\n\tUse --force flag to override the project.'.format(project_name),
    False
  )
  assert not success


def test_run_init_project_already_exists_and_cant_force(mocker, mock_path_exists):
  """
  Test the command defined in :func:`commands.config.init.run`,
  when the project name already exists and can not be removed.
  """
  mock_path_exists(True)
  mocker.patch('commands.config.init.shutil.rmtree', side_effect=Exception)
  project_name = 'my-pdp-project'
  success = False
  with pytest.raises(Exception) as exception:
    success = run(project_name, DEFAULT_CONFIG, True)
  assert 'Can not remove {project_name}.'.format(project_name=project_name.title()) in str(exception.value)
  assert not success


def test_create_project_from_template_project_already_exists(mock_path_exists):
  """
  Test the command defined in :func:`commands.config.init.create_project_from_template`,
  when the project already exists
  """
  mock_path_exists(True)
  project_name = 'my-pdp-project'
  success = create_project_from_template(project_name)
  assert not success


def test_create_project_from_template_project_could_not_copy(mocker, mock_path_exists):
  """
  Test the command defined in :func:`commands.config.init.create_project_from_template`,
  when can not copy the template.
  """
  mock_path_exists(False)
  mocker.patch('commands.config.init.shutil.copytree', side_effect=Exception)
  project_name = 'my-pdp-project'
  success = create_project_from_template(project_name)
  assert not success


def test_create_project_from_template_project_successfully(mocker, mock_path_exists):
  """
  Test the command defined in :func:`commands.config.init.create_project_from_template`.
  """
  mock_path_exists(False)
  mock_copytree = mocker.patch('commands.config.init.shutil.copytree')
  project_name = 'my-pdp-project'
  success = create_project_from_template(project_name)
  abs_path = os.path.abspath(
    os.path.join(os.path.dirname(__file__), os.path.join('templates', 'projects', 'random_generator'))).replace('tests',
                                                                                                                'src')
  mock_copytree.assert_called_once_with(abs_path, project_name)
  assert success


def test_create_project_folder_structure_successful(mocker, mock_path_exists):
  """
  Test the command defined in :func:`commands.config.init.create_project_folder_structure`.
  """
  mock_path_exists(False)
  mock_mkdir = mocker.patch('commands.config.init.os.mkdir')
  project_path = 'fake_path'
  folders = ['Discovery', 'Core', 'Core/Files', 'Ingestion']
  create_project_folder_structure(project_path)
  calls = []
  for folder in folders:
    path = os.path.join(project_path, folder)
    calls += [mocker.call(path)]
  mock_mkdir.assert_has_calls(calls)


def test_create_project_from_existing_sources_successful(mocker):
  """
  Test the command defined in :func:`commands.config.init.create_project_from_existing_sources`.
  """
  mocker.patch('commands.config.init.create_project_folder_structure')
  mocker.patch('commands.config.init.export_all_entities', return_value=(True, { }))
  mocker.patch('commands.config.init.create_project_folder_structure')
  project_name = 'project-name'
  success = create_project_from_existing_sources(project_name, { })
  assert success


def test_create_project_from_existing_sources_failed(mocker):
  """
  Test the command defined in :func:`commands.config.init.create_project_from_existing_sources`.
  """
  mocker.patch('commands.config.init.create_project_folder_structure')
  mocker.patch('commands.config.init.export_all_entities', return_value=(False, None))
  mocker.patch('commands.config.init.handle_and_continue', return_value=(False, None))
  mocker.patch('commands.config.init.create_project_folder_structure')
  project_name = 'project-name'
  success = create_project_from_existing_sources(project_name, { })
  assert not success
