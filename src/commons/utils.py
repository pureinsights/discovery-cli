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

def flat_list(_list: list[any]):
  """
  Takes a list of lists a returns a flatted version.
  That means take out all the elements of the nested lists and returns a list just one level.
  :param list[any] _list: The list to be flatted.
  :rtype: list[any]
  :return: The flatted version of the given list.
  """
  if not isinstance(_list, list):
    return [_list]
  flatlist = []
  # Cut off all the elements which are not a list
  for element in _list:
    if type(element) is not list:
      index = _list.index(element)
      flatlist += [element]
      _list = _list[:index] + _list[index + 1:]

  for nested_list in _list:
    flatlist += flat_list(nested_list)

  return flatlist
