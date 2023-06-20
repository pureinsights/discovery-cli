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

from commons.uuid_commons import is_hex_uuid, is_valid_uuid


def test_is_valid_uuid_invalid():
  """
  Test the function defined in :func:`commons.uuid_commons.is_valid_uuid`,
  when the given uuid is invalid.
  """
  assert not is_valid_uuid("fake-id")


def test_is_hex_uuid():
  """
  Test the function defined in :func:`commons.uuid_commons.is_hex_uuid`.
  """
  assert is_hex_uuid("29a9b5e600704853983b0dd855a11cc6")


def test_is_hex_uuid_valid_uuid_but_no_hex():
  """
  Test the function defined in :func:`commons.uuid_commons.is_hex_uuid`,
  when the given value is a valid uuid but is not hex.
  """
  assert not is_hex_uuid("70ae3bac-305a-4e00-9216-e76fb5b41410")


def test_is_hex_uuid_invalid_uuid():
  """
  Test the function defined in :func:`commons.uuid_commons.is_hex_uuid`,
  when the given value is a valid uuid but is not hex.
  """
  assert not is_hex_uuid("fake")
