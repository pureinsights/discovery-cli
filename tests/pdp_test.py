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

import pyfiglet
import pytest
from click.testing import CliRunner

from commons.constants import DEFAULT_CONFIG
from src.pdp import ensure_configurations, health, load_config, pdp

cli = CliRunner()


def test_ensure_configurations_all_configurations_right():
  """
  If all configurations are present, should not change anything.
  """
  expected_config = {
    'ingestion': 'https://ingestion-fake-url',
    'core': 'https://core-fake-url',
    'staging': 'https://staging-fake-url',
    'discovery': 'https://discovery-fake-url'
  }
  config_result = ensure_configurations(expected_config)
  assert expected_config == config_result


def test_ensure_configurations_missing_configurations():
  """
  Should use default values for those missing values.
  """
  config = {
    'ingestion': 'https://ingestion-fake-url',
    'discovery': 'https://discovery-fake-url'
  }
  expected_config = { **DEFAULT_CONFIG, **config }
  config_result = ensure_configurations(config)
  assert config_result == expected_config


def test_ensure_configurations_all_configuration_missing():
  """
  Should use default values for those missing values.
  """
  config = {
  }
  config_result = ensure_configurations(config)
  assert config_result == DEFAULT_CONFIG


config_return_fixture = {
  'DEFAULT': { **DEFAULT_CONFIG },
  'FAKE': {
    'ingestion': 'https://ingestion-fake-url',
    'core': 'https://core-fake-url',
    'staging': 'https://staging-fake-url',
    'discovery': 'https://discovery-fake-url'
  }
}


def test_load_config_default_profile():
  """
  Should read the default profile correctly if a profile was not provided.
  """
  config_name = "pdp_test.ini"  # The same name must be defined in conftest.py mock_os_path_exists
  config_result = load_config(config_name, 'DEFAULT')
  assert config_result == DEFAULT_CONFIG


def test_load_config_fake_profile():
  """
  Should read the fake profile correctly.
  """
  expected_config = {
    'ingestion': 'http://ingestion-fake',
    'discovery': 'http://discovery-fake/admin',
    'core': 'http://core-fake',
    'staging': 'http://staging-fake'
  }
  config_name = "pdp_test.ini"  # The same name must be defined in conftest.py mock_os_path_exists
  config_result = load_config(config_name, 'FAKE')
  assert { **config_result } == expected_config


def test_load_config_invalid_profile():
  with pytest.raises(Exception) as exception:
    load_config("pdp_test.ini", 'NOT_EXISTS')
  assert str(exception.value) == 'Configuration profile NOT_EXISTS was not found.'


def test_pdp():
  """
  Should end with an exit code 0.
  """
  response = cli.invoke(pdp, [])
  assert response.exit_code == 0


def test_health():
  """
  Should show an styled message with information
  about the version and a link to the web page of
  pureinsights.
  """
  response = cli.invoke(health)
  ascii_art_pdp_cli = pyfiglet.figlet_format("PDP - CLI")
  assert response.exit_code == 0
  assert f"{ascii_art_pdp_cli}Pureinsights Discovery Platform: Command Line Interface\nv1.5.0\nhttps://pureinsights.com/" in response.output
