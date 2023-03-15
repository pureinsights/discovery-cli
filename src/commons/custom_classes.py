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


class PdpEntity:
  associated_file_name: str = ''
  type: str = ''
  product: str = ''
  reference_field: str = ''

  def __init__(self, product: str, type: str, file: str, reference_field=None):
    self.associated_file_name = file
    self.type = type
    self.product = product
    if reference_field is None:
      self.reference_field = f'{type}Id'
    else:
      self.reference_field = reference_field


class DataInconsistency(Exception):
  """
  Raised when data does not match or expected fields are missing.
  """

  def __init__(self, **kwargs):
    from commons.constants import ERROR_SEVERITY
    self.message = kwargs.get('message', None)
    if self.message is not None:
      super().__init__(self.message)
    self.handled = kwargs.get('handled', True)
    self.severity = kwargs.get('severity', ERROR_SEVERITY)
    self.content = kwargs.get('content', {})


class PdpException(Exception):
  """
  Raised to force handlers to manage the error. It's a general exception controlled by a PDP developer.
  """

  def __init__(self, **kwargs):
    self.message = kwargs.get('message', None)
    if self.message is not None:
      super().__init__(self.message)
    self.handled = kwargs.get('handled', False)
    self.content = kwargs.get('content', {})
