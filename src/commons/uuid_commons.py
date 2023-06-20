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
import uuid


def is_valid_uuid(_id: str) -> bool:
  """
  Check if a str is a valid UUID.
  :param str _id: The str to check.
  :rtype: bool
  :return: True if is a valid UUID, False in other case.
  """
  try:
    uuid.UUID(_id)
    return True
  except ValueError:
    return False


def is_hex_uuid(_id: str) -> bool:
  """
  Check if a str is a valid UUID and if is a hex str.
  :param str _id: The str to check.
  :rtype: bool
  :return: True if is a valid UUID and valid hex str, False in other case.
  """
  if not is_valid_uuid(_id):
    return False

  try:
    int(_id, 16)
    return True
  except ValueError:
    return False
