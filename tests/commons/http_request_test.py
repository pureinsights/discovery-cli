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

from commons.http_requests import delete, get, post, put


def test_get_success(mocker):  # All the possible errors are tested in test_handle_http_response
  """
  Test the function defined in :func:`commons.http_requests.get`.
  """
  expected_response = {'status_code': 200, 'content': 'fake-content'}
  mocker.patch('requests.get', return_value=expected_response)
  mock_handler = mocker.patch('commons.http_requests.handle_http_response')
  get('http://fake-url')
  mock_handler.assert_called_once_with(expected_response, True)


def test_post_success(mocker):  # All the possible errors are tested in test_handle_http_response
  """
  Test the function defined in :func:`commons.http_requests.post`.
  """
  fake_body = {'fake': 'body'}
  fake_url = 'http://fake-url'
  expected_response = {'status_code': 200, 'content': 'fake-content'}
  request_post_mock = mocker.patch('requests.post', return_value=expected_response)
  mock_handler = mocker.patch('commons.http_requests.handle_http_response')
  post(fake_url, json=fake_body)
  request_post_mock.assert_called_once_with(fake_url, json=fake_body)
  mock_handler.assert_called_once_with(expected_response, True)


def test_put_success(mocker):  # All the possible errors are tested in test_handle_http_response
  """
  Test the function defined in :func:`commons.http_requests.put`.
  """
  fake_body = {'fake': 'body'}
  fake_url = 'http://fake-url'
  expected_response = {'status_code': 200, 'content': 'fake-content'}
  request_put_mock = mocker.patch('requests.put', return_value=expected_response)
  mock_handler = mocker.patch('commons.http_requests.handle_http_response')
  put(fake_url, json=fake_body)
  request_put_mock.assert_called_once_with(fake_url, json=fake_body)
  mock_handler.assert_called_once_with(expected_response)


def test_delete_success(mocker):  # All the possible errors are tested in test_handle_http_response
  """
  Test the function defined in :func:`commons.http_requests.delete`.
  """
  fake_body = {'fake': 'body'}
  fake_url = 'http://fake-url'
  expected_response = {'status_code': 200, 'content': 'fake-content'}
  request_delete_mock = mocker.patch('requests.delete', return_value=expected_response)
  mock_handler = mocker.patch('commons.http_requests.handle_http_response')
  delete(fake_url, params=fake_body)
  request_delete_mock.assert_called_once_with(fake_url, params=fake_body)
  mock_handler.assert_called_once_with(expected_response, True)
