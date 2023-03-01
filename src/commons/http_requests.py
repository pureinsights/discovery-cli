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

import requests as req

from commons.handlers import handle_http_response


def get(url: str):
  """
  Performs a http(s) call, method GET.

  :param str url: The url where the http request will be sent.
  :rtype: any
  :return: Returns a class Response containing the http response.
  """
  res = req.get(url)
  return handle_http_response(res)  # Calls to handle_http_response to handle any exception related with the status code
