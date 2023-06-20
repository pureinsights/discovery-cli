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
from commons.utils import flat_list


def test_flat_list():
  """
  Test the function defined in :func:`commons.utils.flat_list`,
  when the param is not a list.
  """
  assert flat_list([2, 8, ['a', [{}]], [[[], [False]]]]) == [2, 8, 'a', {}, False]


def test_flat_list_not_a_list():
  """
  Test the function defined in :func:`commons.utils.flat_list`,
  when the param is not a list.
  """
  assert flat_list(8) == [8]


