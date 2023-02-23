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

import pyfiglet
from click.testing import CliRunner
from pdp import pdp, health

cli = CliRunner()


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
