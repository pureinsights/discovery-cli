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
import requests

from commons.handlers import handle_and_continue, handle_and_exit, handle_exceptions, handle_http_response


def mock_custom_exception(exception):
  if exception is not None:
    raise exception


def test_handle_exceptions_exception_handled(mocker):
  """
  Should show a specific message for each exception handled.
  """
  exception = Exception()
  mock_print = mocker.patch('commons.handlers.print_exception')
  mock_stop_spinner = mocker.patch('commons.handlers.stop_spinner')
  handle_exceptions(mock_custom_exception, exception)
  mock_print.assert_called_once_with(exception)
  mock_stop_spinner.assert_called_once()


def test_handle_exceptions_no_exception_happen(mocker):
  """
  Should show a specific message for each exception handled.
  """
  exception = None
  mock_print = mocker.patch('commons.handlers.print_exception')
  mock_stop_spinner = mocker.patch('commons.handlers.stop_spinner')
  handle_exceptions(mock_custom_exception, exception)
  assert mock_print.call_count == 0
  mock_stop_spinner.assert_called_once()


def test_handle_http_response_status_distinct_2xx(mocker):
  response = requests.Response()
  response.status_code = 404
  with pytest.raises(requests.exceptions.HTTPError) as exception:
    handle_http_response(response)
  assert exception is not None


def test_handle_http_response_no_content(mocker):
  response = requests.Response()
  response.status_code = 204
  handled_response = handle_http_response(response)
  assert handled_response is None


def test_handle_http_response_success(mocker):
  response = mocker.Mock(content=b'Hello world', status_code=200)
  handled_response = handle_http_response(response)
  assert handled_response == response.content


def test_handle_and_exit_successful():
  response = handle_and_exit(mock_custom_exception, { }, None)
  assert response == (True, None)


def test_handle_and_exit_fail():
  with pytest.raises(Exception) as exception:
    handle_and_exit(mock_custom_exception, { }, Exception())
  assert exception is not None


def test_handle_and_exit_show_exception_and_error(mocker):
  mock_print_exception = mocker.patch('commons.handlers.print_exception')
  mock_print_error = mocker.patch('commons.handlers.print_error')
  message = 'fake-error'
  custom_exception = Exception
  with pytest.raises(Exception) as exception:
    handle_and_exit(mock_custom_exception, {
      'message': message,
      'show_exception': True
    }, custom_exception)
  assert mock_print_exception.call_count == 1
  mock_print_error.assert_called_once_with(message, True, prefix='', suffix='')


def test_handle_and_continue_show_error_and_exception(mocker):
  mock_print_exception = mocker.patch('commons.handlers.print_exception')
  mock_print_error = mocker.patch('commons.handlers.print_error')
  message = 'fake-message'
  handle_and_continue(mock_custom_exception, {
    'message': message,
    'show_exception': True
  }, Exception)
  assert mock_print_exception.call_count == 1
  mock_print_error.assert_called_once_with(message, False, prefix='', suffix='')


def test_handle_and_exit_not_show_nothing():
  response = handle_and_continue(mock_custom_exception, { }, Exception)
  assert response == (False, None)
