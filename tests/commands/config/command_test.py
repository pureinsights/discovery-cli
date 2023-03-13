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
from pdp import pdp
from pdp_test import cli


def test_config(snapshot):
  """
  Test the command defined in :func:`src.commands.config.command.config`.
  """
  response = cli.invoke(pdp, ["config"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_config.snapshot')


def test_init_success(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.config.command.init`,
  without arguments.
  """
  init_run_mocked = mocker.patch('commands.config.command.run_init', return_value=True)
  project_name = "my-pdp-project"
  response = cli.invoke(pdp, ["config", "init", "--empty", "--template", "empty"])
  init_run_mocked.assert_called()
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_init_success.snapshot')


def test_init_could_not_create(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.config.command.init`,
  when some error happens.
  """
  init_run_mocked = mocker.patch('commands.config.command.run_init', return_value=False)
  response = cli.invoke(pdp, ["config", "init"])
  init_run_mocked.assert_called()
  assert response.exit_code == 1
  snapshot.assert_match(response.output, 'test_init_could_not_create.snapshot')


def test_init_parse_options(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.config.command.init`,
  with all the arguments provided.
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
                        ["config", "init", "-n", project_name, no_empty, force, '--template', 'empty', '-u',
                         'ingestion',
                         'http://ingestion-fake', '-u', 'discovery', 'http://ingestion-fake', '-u', 'core',
                         'http://ingestion-fake', '-u', 'staging', 'http://ingestion-fake'])

  init_run_mocked.assert_called_once_with(project_name, expected_config, True, None)
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_init_parse_options.snapshot')


def test_init_incorrect_option_product(snapshot):
  """
  Test the command defined in :func:`src.commands.config.command.init`,
  with an unrecognized product of the argument product-url.
  """
  project_name = "my-pdp-project"
  response = cli.invoke(pdp, ["config", "init", '-u', 'fake-product',
                              'http://ingestion-fake'])
  assert response.exit_code == 1
  snapshot.assert_match(response.output, 'test_init_incorrect_option_product.snapshot')
