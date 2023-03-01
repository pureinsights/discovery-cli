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
def create_file(tmp_path, request):
  params = request.node.get_closest_marker("params")
  if params is None:
    return

  path = os.path.join(tmp_path, params[0])
  with open(path, mode='w') as file:
    file.write(params[1])
  return path


@pytest.fixture
@pytest.mark.params(config_file_test_name, config_file_test_content)
def create_test_config_file(create_file):
  return create_file
