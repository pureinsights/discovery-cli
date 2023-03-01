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
#
#  Permission to use, copy, modify or distribute this software and its
#  documentation for any purpose is subject to a licensing agreement with
#  Pureinsights Technology Ltd.
#
#  All information contained within this file is the property of
#  Pureinsights Technology Ltd. The distribution or reproduction of this
#  file or any information contained within is strictly forbidden unless
#  prior written permission has been granted by Pureinsights Technology Ltd.

from pdp_test import cli
from src.pdp import pdp


def test_config():
  """
  Should end with an exit code 0.
  """
  response = cli.invoke(pdp, ["config"])
  assert response.exit_code == 0


def test_init_success(mocker):
  """
  Should show a successful message to the user. And use default values.
  """
  init_run_mocked = mocker.patch('commands.config.command.run_init', return_value=True)
  project_name = "my-pdp-project"
  response = cli.invoke(pdp, ["config", "init"])
  init_run_mocked.assert_called()
  assert response.exit_code == 0
  assert f"Project {project_name} created successfully." in response.output


def test_init_could_not_create(mocker):
  """
  Should show an error message to the user.
  """
  init_run_mocked = mocker.patch('commands.config.command.run_init', return_value=False)
  project_name = "my-pdp-project"
  response = cli.invoke(pdp, ["config", "init"])
  init_run_mocked.assert_called()
  assert response.exit_code == 0
  assert 'Could not create the project {0}.\n'.format(project_name) in response.output


def test_init_parse_options(mocker):
  """
  Should show a successful message to the user. But parse correctly all the options provided.
  """
  init_run_mocked = mocker.patch('commands.config.command.run_init', return_value=True)
  project_name = "my-pdp-project"
  no_empty = '--no-empty'
  expected_config = {
    'ingestion': 'http://ingestion-fake',
    'discovery': 'http://ingestion-fake',
    'core': 'http://ingestion-fake',
    'staging': 'http://ingestion-fake'
  }
  force = '--force'
  response = cli.invoke(pdp,
                        ["config", "init", "-n", project_name, no_empty, force, '-u', 'ingestion',
                         'http://ingestion-fake', '-u', 'discovery', 'http://ingestion-fake', '-u', 'core',
                         'http://ingestion-fake', '-u', 'staging', 'http://ingestion-fake'])

  init_run_mocked.assert_called_once_with(project_name, False, expected_config, True)
  assert response.exit_code == 0
  assert f"Project {project_name} created successfully." in response.output


def test_init_incorrect_option_product():
  """
  Should show an error message to the user, about the unrecognized product.
  """
  project_name = "my-pdp-project"
  response = cli.invoke(pdp, ["config", "init", '-u', 'fake-product',
                              'http://ingestion-fake'])
  assert response.exit_code == 1
  assert f'Unrecognized product "fake-product".' in response.output
