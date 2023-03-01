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

from commons.http_requests import get


def test_get_success(mocker):  # All the possible errors are tested in test_handle_http_response
  expected_response = { 'status_code': 200, 'content': 'fake-content' }
  mocker.patch('requests.get', return_value=expected_response)
  mock_handler = mocker.patch('commons.http_requests.handle_http_response')
  get('http://fake-url')
  mock_handler.assert_called_once_with(expected_response)
