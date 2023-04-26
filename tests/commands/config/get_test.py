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
from commands.config.get import get_all_entities, get_entities_by_ids, print_stage
from commons.constants import ENDPOINT, PIPELINE


def test_get_all_entities_entities_not_found(mocker):
  """
  Test the function defined in :func:`commands.config.get.get_all_entities`,
  when there are no entities on the product.
  """
  mocker.patch("commands.config.get.get", return_value=None)
  assert get_all_entities({'ingestion': {}}, [PIPELINE], {}, False) == {}


def test_get_all_entities_entities_not_found_verbose(mocker):
  """
  Test the function defined in :func:`commands.config.get.get_all_entities`,
  when there are no entities on the product with verbose.
  """
  mocker.patch("commands.config.get.create_spinner")
  mocker.patch("commands.config.get.get", return_value=None)
  assert get_all_entities({'ingestion': {}}, [PIPELINE], {}, True) == {}


def test_get_entities_by_ids_invalid(mocker):
  """
  Test the function defined in :func:`commands.config.get.get_entities_by_ids`,
  when there are no entities on the product with verbose.
  """
  mocker.patch("commands.config.get.create_spinner")
  mocker.patch("commands.config.get.get", return_value=None)
  assert get_entities_by_ids({'discovery': {}}, ['70ae3bac-305a-4e00-9216-e76fb5b41410'], [ENDPOINT], {}, True) \
         == {}


def test_get_entities_by_ids_valid_hex_and_not_discovery(mocker):
  """
  Test the function defined in :func:`commands.config.get.get_entities_by_ids`,
  when there are no entities on the product with verbose.
  """
  mocker.patch("commands.config.get.create_spinner")
  mocker.patch("commands.config.get.get", return_value=None)
  assert get_entities_by_ids({'ingestion': {}}, ['29a9b5e600704853983b0dd855a11cc6'], [PIPELINE], {}, False) \
         == {}


def test_get_entities_by_ids_error_curred(mocker, mock_custom_exception):
  """
  Test the function defined in :func:`commands.config.get.get_entities_by_ids`,
  when there are no entities on the product with verbose.
  """
  mocker.patch("commands.config.get.create_spinner")
  mocker.patch("commands.config.get.get_entity_by_id", side_effect=lambda: mock_custom_exception(Exception))
  assert get_entities_by_ids({'ingestion': {}}, ['29a9b5e600704853983b0dd855a11cc6'], [PIPELINE], {}, False) \
         == {}


def test_print_stage_empty_entities(mocker):
  """
  Test the function defined in :func:`commands.config.get.print_stage`,
  when a product doesn't have entities.
  """
  print_console_mock = mocker.patch("commands.config.get.print_console")
  print_stage({'discovery': {}}, False, False)
  assert print_console_mock.call_count == 2
