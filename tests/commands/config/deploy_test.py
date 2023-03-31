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
import commons.console
from commands.config.deploy import run as run_deploy
from commons.console import suppress_errors, suppress_warnings
from commons.constants import CORE, DEFAULT_CONFIG, INGESTION, PRODUCTS, STAGING


def test_run_deploy(test_project_path, mocker):
  """
  Test the command defined in :func:`commands.config.deploy.run`.
  """
  mocker.patch("commands.config.deploy.write_entities")
  mocker.patch("commands.config.deploy.create_spinner")
  mocker.patch("commands.config.deploy.print_console")
  create_update_mock = mocker.patch("commands.config.deploy.create_or_update_entity", return_value='fakeid')
  suppress_errors(True)
  suppress_warnings(True)
  target_products = [product for product in PRODUCTS['list'] if product != STAGING]
  run_deploy(DEFAULT_CONFIG, test_project_path(), target_products)
  commons.console.is_errors_suppressed = False
  commons.console.is_warnings_suppressed = False
  assert create_update_mock.call_count == 6


def test_run_deploy_quiet_mode(test_project_path, mocker):
  """
  Test the command defined in :func:`commands.config.deploy.run`,
  when quiet mode is activated.
  """
  mocker.patch("commands.config.deploy.get_number_errors_exceptions", return_value=0)
  mocker.patch("commands.config.deploy.write_entities")
  mocker.patch("commands.config.deploy.create_spinner")
  print_mock = mocker.patch("commands.config.deploy.print_console")
  create_update_mock = mocker.patch("commands.config.deploy.create_or_update_entity", return_value='fakeid')
  suppress_errors(True)
  suppress_warnings(True)
  target_products = [product for product in PRODUCTS['list'] if product != STAGING and product != CORE]
  run_deploy(DEFAULT_CONFIG, test_project_path(), target_products, False, False, True)
  commons.console.is_errors_suppressed = False
  commons.console.is_warnings_suppressed = False
  assert create_update_mock.call_count == 6
  print_mock.assert_called_with("fakeid")


def test_run_deploy_verbose_mode(test_project_path, mocker):
  """
  Test the command defined in :func:`commands.config.deploy.run`,
  when verbose mode is activated.
  """
  mocker.patch("commands.config.deploy.write_entities")
  mocker.patch("commands.config.deploy.create_spinner")
  mocker.patch("commands.config.deploy.print_console")
  create_update_mock = mocker.patch("commands.config.deploy.create_or_update_entity", return_value='fakeid')
  suppress_errors(True)
  suppress_warnings(True)
  target_products = [product for product in PRODUCTS['list'] if product != STAGING and product != INGESTION]
  run_deploy(DEFAULT_CONFIG, test_project_path(), target_products, True, False, False)
  commons.console.is_errors_suppressed = False
  commons.console.is_warnings_suppressed = False
  assert create_update_mock.call_count == 2


def test_run_deploy_ignoring_ids(test_project_path, mocker):
  """
  Test the command defined in :func:`commands.config.deploy.run`,
  when ignore_ids mode is activated.
  """
  mocker.patch("commands.config.deploy.write_entities")
  mocker.patch("commands.config.deploy.create_spinner")
  mocker.patch("commands.config.deploy.print_console")
  create_update_mock = mocker.patch("commands.config.deploy.create_or_update_entity", return_value=None)
  suppress_errors(True)
  suppress_warnings(True)
  target_products = [product for product in PRODUCTS['list'] if product != STAGING]
  run_deploy(DEFAULT_CONFIG, test_project_path(), target_products, True, True, False)
  commons.console.is_errors_suppressed = False
  commons.console.is_warnings_suppressed = False
  assert create_update_mock.call_count == 6
