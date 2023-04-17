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
import errno
import os

import click
import pytest
import requests
from yaspin.spinners import Spinners

import commons.console
from commons.console import create_spinner, get_number_errors_exceptions, print_console, print_error, \
  print_exception, print_warning, spinner_change_text, spinner_fail, spinner_ok, stop_spinner, \
  suppress_errors, suppress_warnings, verbose
from commons.constants import EXCEPTION_FORMAT, WARNING_FORMAT, WARNING_SEVERITY
from commons.custom_classes import DataInconsistency, PdpException


def test_create_spinner_successful(mocker):
  """
  Test the function defined in :func:`commons.console.create_spinner`.
  """
  mock_spinner_start = mocker.patch('commons.console.Yaspin.start')
  create_spinner()
  mock_spinner_start.assert_called_once()
  assert commons.console.Spinner.text is not None
  commons.console.Spinner = None


def test_create_spinner_successful_with_arguments_and_a_spinner_created(mocker):
  """
  Test the function defined in :func:`commons.console.create_spinner`.
  """
  mock_spinner_start = mocker.patch('commons.console.Yaspin.start')
  mock_spinner_stop = mocker.patch('commons.console.Yaspin.stop')
  create_spinner()
  create_spinner(Spinners.dots12)
  mock_spinner_stop.assert_called_once()
  assert mock_spinner_start.call_count == 2
  assert commons.console.Spinner.text is not None
  commons.console.Spinner = None


def test_stop_spinner_successful(mocker):
  """
  Test the function defined in :func:`commons.console.stop_spinner`.
  """
  mocker.patch('commons.console.Yaspin.start')
  mock_spinner_stop = mocker.patch('commons.console.Yaspin.stop')
  create_spinner()
  stop_spinner()
  assert mock_spinner_stop.call_count == 1
  assert commons.console.Spinner is None


def test_stop_spinner_spinner_is_None(mocker):
  """
  Test the function defined in :func:`commons.console.stop_spinner`,
  when Spinner is None.
  """
  mock_spinner_start = mocker.patch('commons.console.Yaspin.stop')
  stop_spinner()
  assert mock_spinner_start.call_count == 0


def test_spinner_change_text_successful(mocker):
  """
  Test the function defined in :func:`commons.console.spinner_change_text`.
  """
  mocker.patch('commons.console.Yaspin.start')
  create_spinner()
  spinner_change_text('fake-text')
  assert commons.console.Spinner.text == 'fake-text'
  commons.console.Spinner = None


def test_spinner_change_text_spinner_is_none(mocker):
  """
  Test the function defined in :func:`commons.console.spinner_change_text`,
  when Spinner is None.
  """
  mocker.patch('commons.console.Yaspin.start')
  spinner_change_text('fake-text')
  assert commons.console.Spinner is None


def test_spinner_ok_successful(mocker):
  """
  Test the function defined in :func:`commons.console.spinner_ok`.
  """
  mocker.patch('commons.console.Yaspin.start')
  mock_ok = mocker.patch('commons.console.Yaspin.ok')
  mock_stop = mocker.patch('commons.console.Yaspin.stop')
  message = 'ok-fake'
  icon = 'fake-icon'
  commons.console.buffer = 'some-fake-text'
  create_spinner()
  spinner_ok(message, icon=icon)
  mock_ok.assert_called_once_with(icon)
  mock_stop.assert_called_once()
  assert commons.console.buffer == ''


def test_spinner_ok_successful(mocker):
  """
  Test the function defined in :func:`commons.console.spinner_ok`,
  without icon.
  """
  mocker.patch('commons.console.Yaspin.start')
  mock_ok = mocker.patch('commons.console.Yaspin.ok')
  mock_stop = mocker.patch('commons.console.Yaspin.stop')
  message = 'ok-fake'
  icon = ''
  commons.console.buffer = 'some-fake-text'
  create_spinner()
  spinner_ok(message, icon=icon)
  assert mock_ok.call_count == 0
  mock_stop.assert_called_once()
  assert commons.console.buffer == ''


def test_spinner_ok_spinner_is_none(mocker):
  """
  Test the function defined in :func:`commons.console.spinner_ok`,
  when Spinner is None.
  """
  mock_print_console = mocker.patch('commons.console.print_console')
  message = 'ok-fake'
  icon = ''
  spinner_ok(message, icon=icon)
  mock_print_console.assert_called_once_with(message, prefix=icon)


def test_spinner_fail_successful(mocker):
  """
  Test the function defined in :func:`commons.console.spinner_fail`.
  """
  mocker.patch('commons.console.Yaspin.start')
  mock_fail = mocker.patch('commons.console.Yaspin.fail')
  mock_stop = mocker.patch('commons.console.Yaspin.stop')
  message = 'ok-fake'
  icon = 'fake-icon'
  commons.console.buffer = 'some-fake-text'
  create_spinner()
  spinner_fail(message, icon=icon)
  mock_fail.assert_called_once_with(icon)
  mock_stop.assert_called_once()
  assert commons.console.buffer == ''


def test_spinner_fail_spinner_is_none(mocker):
  """
  Test the function defined in :func:`commons.console.spinner_fail`,
  when Spinner is None.
  """
  mock_print_console = mocker.patch('commons.console.print_console')
  message = 'ok-fake'
  icon = 'fake-icon'
  spinner_fail(message, icon=icon)
  mock_print_console.assert_called_once_with(message, prefix=icon)


def test_print_console_successful(mocker):
  """
  Test the function defined in :func:`commons.console.print_console`.
  """
  mock_secho = mocker.patch('commons.console.click.secho')
  message = 'fake-message'
  print_console(message)
  mock_secho.assert_called_once_with(message, nl=True)


def test_print_console_spinner_is_not_none(mocker):
  """
  Test the function defined in :func:`commons.console.print_console`,
  when Spinner is not None.
  """
  mocker.patch('commons.console.Yaspin.start')
  mock_secho = mocker.patch('commons.console.click.secho')
  message = 'fake-message'
  create_spinner()
  print_console(message)
  assert mock_secho.call_count == 0
  assert commons.console.buffer == f'{message}\n'


def test_print_console_no_print_if_message_is_none(mocker):
  """
  Test the function defined in :func:`commons.console.print_console`,
  when the message is None.
  """
  secho_mock = mocker.patch("commons.console.click.secho")
  print_console(None)
  assert secho_mock.call_count == 0


def test_print_warning_successful(mocker):
  """
  Test the function defined in :func:`commons.console.print_warning`.
  """
  mock_print_console = mocker.patch('commons.console.print_console')
  message = 'fake-warning'
  styled_warning = click.style(WARNING_FORMAT.format(message=message), fg='yellow')
  print_warning(message)
  mock_print_console.assert_called_once_with(styled_warning)


def test_print_warning_no_print_if_message_is_none(mocker):
  """
  Test the function defined in :func:`commons.console.print_warning`,
  when the message is None.
  """
  secho_mock = mocker.patch("commons.console.click.secho")
  print_warning(None)
  assert secho_mock.call_count == 0


@pytest.mark.parametrize('exception', [
  DataInconsistency(message='fake-message', handled=True),
  DataInconsistency(message='fake-message', handled=False),
  PdpException(message='fake-message', handled=True),
  PdpException(message='fake-message', handled=False)
])
def test_print_exception_successful_errors(mocker, exception):
  """
  Test the function defined in :func:`commons.console.print_exception`,
  with error severity.
  """
  mock_error = mocker.patch('commons.console.print_error')
  print_exception(exception)
  mock_error.assert_called_once_with(exception.message, not exception.handled)


@pytest.mark.parametrize('exception', [
  DataInconsistency(message='fake-message', handled=True, severity=WARNING_SEVERITY),
  DataInconsistency(message='fake-message', handled=False, severity=WARNING_SEVERITY)
])
def test_print_exception_successful_warnings(mocker, exception):
  """
  Test the function defined in :func:`commons.console.print_exception`,
  with warning severity.
  """
  mock_warning = mocker.patch('commons.console.print_warning')
  print_exception(exception)
  mock_warning.assert_called_once_with(exception.message, not exception.handled)


def test_print_exception_connection_error(mocker):
  """
  Test the function defined in :func:`commons.console.print_exception`,
  with a connection error.
  """
  mock_error = mocker.patch('commons.console.print_error')
  fake_url = 'http://fake-url'
  custom_exception = requests.exceptions.ConnectionError()
  mocker.patch.object(custom_exception, 'request')
  custom_exception.request.url = fake_url
  print_exception(custom_exception)
  mock_error.assert_called_once_with(f'ConnectionError. Can not connect to {custom_exception.request.url}.',
                                     prefix='',
                                     suffix='')


def test_print_exception_file_not_found_error(mocker):
  """
  Test the function defined in :func:`commons.console.file_not_found_error`,
  with a File Not Found Error.
  """
  mock_error = mocker.patch('commons.console.print_error')
  fake_path = 'fake-path'
  custom_exception = FileNotFoundError(errno.ENOENT, os.strerror(errno.ENOENT), fake_path)
  print_exception(custom_exception)
  mock_error.assert_called_once_with(f'{custom_exception.strerror}: {custom_exception.filename}')


def test_print_error_no_print_if_message_is_none(mocker):
  """
  Test the function defined in :func:`commons.console.print_error`,
  when the message is None.
  """
  secho_mock = mocker.patch("commons.console.click.secho")
  print_error(None)
  assert secho_mock.call_count == 0


@pytest.mark.parametrize('params', [
  (Exception('Fake Exception'), False),
  (Exception('Fake Exception'), True)
])
def test_print_exception_general_exception_error(mocker, params):
  """
  Test the function defined in :func:`commons.console.print_exception`,
  with any unhandled exception.
  """
  mock_error = mocker.patch('commons.console.print_error')
  custom_exception, raise_exception = params
  print_exception(custom_exception, raise_exception=raise_exception)
  mock_error.assert_called_once_with(EXCEPTION_FORMAT.format(exception=type(custom_exception).__name__, error=''),
                                     raise_exception)


def test_suppress_warnings(mocker):
  """
  Test the function defined in :func:`commons.console.suppress_warnings`.
  """
  print_console_mock = mocker.patch("commons.console.print_console")
  suppress_warnings(True)
  print_warning("fake-warning")
  assert commons.console.is_warnings_suppressed
  assert print_console_mock.call_count == 0


def test_suppress_errors(mocker):
  """
  Test the function defined in :func:`commons.console.suppress_errors`.
  """
  print_console_mock = mocker.patch("commons.console.print_console")
  suppress_errors(True)
  print_error("")
  print_exception(Exception())
  assert commons.console.is_errors_suppressed
  assert print_console_mock.call_count == 0


def test_get_number_errors_exceptions():
  """
  Test the function defined in :func:`commons.console.get_number_errors_exceptions`.
  """
  commons.console.printed_errors = []
  commons.console.printed_exceptions = []
  suppress_errors(True)
  print_error("fake error")
  print_exception(Exception())
  assert get_number_errors_exceptions() == 2


def test_verbose():
  """
  Test the function defined in :func:`commons.console.verbose`.
  """
  result = verbose(
    verbose=True,
    verbose_func=lambda: 'verbose',
    not_verbose_func=lambda: 'no verbose'
  )
  assert result == 'verbose'


def test_verbose_when_not_verbose():
  """
  Test the function defined in :func:`commons.console.verbose`,
  when verbose is False.
  """
  result = verbose(
    verbose=False,
    verbose_func=lambda: 'verbose',
    not_verbose_func=lambda: 'no verbose'
  )
  assert result == 'no verbose'


def test_verbose_when_not_verbose_and_is_not_a_function():
  """
  Test the function defined in :func:`commons.console.verbose`,
  when verbose is False and the not_verbose_func is not a function.
  """
  result = verbose(
    verbose=False,
    verbose_func='verbose',
    not_verbose_func='no verbose'
  )
  assert result == 'no verbose'


def test_verbose_is_not_a_function():
  """
  Test the function defined in :func:`commons.console.verbose`,
  when the verbose_func is not a function.
  """
  result = verbose(
    verbose=True,
    verbose_func='verbose',
    not_verbose_func='no verbose'
  )
  assert result == 'verbose'


def test_verbose_function_not_provided_verbose():
  """
  Test the function defined in :func:`commons.console.verbose`,
  when the verbose_func were not provided.
  """
  result = verbose(
    verbose=True,
    not_verbose_func='no verbose'
  )
  assert result is None


def test_verbose_function_not_provided_not_verbose():
  """
  Test the function defined in :func:`commons.console.verbose`,
  when the verbose_func were not provided.
  """
  result = verbose(
    verbose=False,
    verbose_func='verbose',
  )
  assert result is None
