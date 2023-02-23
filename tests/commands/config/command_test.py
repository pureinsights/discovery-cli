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

from commands.config.command import config, init
from pdp_test import cli


def test_config():
    """
    Should end with an exit code 0.
    """
    response = cli.invoke(config, [])
    assert response.exit_code == 0


def test_init():
    project_name = "Hello world"
    response = cli.invoke(init, ["-n", project_name])
    assert response.exit_code == 0
    assert f"Project {project_name} creted successfully." in response.output
