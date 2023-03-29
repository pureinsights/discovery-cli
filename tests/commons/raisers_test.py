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
import pytest

from commons.raisers import raise_file_not_found_error


def test_raise_file_not_found_error():
  """
  Test the function defined in :func:`src.commons.raisers.test_raise_file_not_found_error`,
  when the file not exists.
  """
  fake_path = "./fake/path"
  with pytest.raises(FileNotFoundError) as error:
    raise_file_not_found_error(fake_path)
  assert error.value.filename == fake_path


def test_raise_file_not_found_error_file_exists(mocker):
  """
  Test the function defined in :func:`src.commons.raisers.test_raise_file_not_found_error`,
  when the file exists.
  """
  fake_path = "fake/path"
  path_exists_mock = mocker.patch("commons.raisers.os.path.exists", returned_value=True)
  raise_file_not_found_error(fake_path)
  path_exists_mock.assert_called_with(fake_path)
