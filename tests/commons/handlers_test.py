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
from commons.constants import EXCEPTION_FORMAT
from commons.handlers import handle_exceptions


def mock_custom_exception(exception):
  raise exception


def test_handle_exceptions(capsys):
  """
  Should show a specific message for each exception handled.
  """
  exception = Exception("Unknown")
  handle_exceptions(mock_custom_exception, exception)
  captured = capsys.readouterr()
  assert EXCEPTION_FORMAT.format(exception=type(exception).__name__, error='') in captured.err
