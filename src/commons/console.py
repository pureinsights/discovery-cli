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
from typing import Union

import click
import requests as req
from yaspin import yaspin
from yaspin.core import Yaspin

from commons.constants import ERROR_FORMAT, EXCEPTION_FORMAT, WARNING_FORMAT
from commons.custom_classes import DataInconsistency, PdpException

Spinner: Union[Yaspin, None] = None  # An instance of the Yaspin spinner
buffer: str = ''  # This is a buffer to store messages that want to be printed after the spinner stops


def create_spinner(*args, **kwargs):
  """
  Creates and starts a new spinner.

  :param *args args: The positional arguments passed to Yaspin.yaspin.
  :param **kwargs kwargs: The keyword arguments passed to Yaspin.yaspin.
  """
  global Spinner
  Spinner = yaspin(*args, **kwargs)
  Spinner.start()


def stop_spinner():
  """
  It stops and clean the instance of the spinner.
  """
  global Spinner
  if Spinner is None:
    return
  Spinner.stop()
  Spinner = None


def spinner_change_text(message: str):
  """
  Changes the current text of the spinner.

  :param str message: The text that will be showed.
  """
  if Spinner is None:
    return
  Spinner.text = message


def spinner_ok(message: str, **kwargs):
  """
  It will stop the spinner and sets a text to the success status.

  :param str message: The message that will be showed.
  :key str prefix: A string that will be added in front fo the message.
  :key str suffix: A string that will be added at the end of the message.
  :key str icon: A string to show at the front of the message. Default is '✔ '.
  :param **kwargs kwargs: Keyword arguments passed to print_console.
  """
  icon = kwargs.get('icon', '✔ ')
  prefix = kwargs.get('prefix', '')
  suffix = kwargs.get('suffix', '')
  if Spinner is None:
    print_console(message, prefix=f'{prefix}{icon}')
    return
  global buffer
  icon = kwargs.get('icon', '✔ ')
  prefix = kwargs.get('prefix', '')
  suffix = kwargs.get('suffix', '')
  Spinner.text = message
  Spinner.ok(icon)
  stop_spinner()
  if buffer != '':
    print_console(f'{prefix}{buffer}{suffix}', nl=False)
    buffer = ''


def spinner_fail(message: str, **kwargs):
  """
  It will stop the spinner and sets a text to the failed status.

  :param str message: The message that will be showed.
  :key str prefix: A string that will be added in front fo the message.
  :key str suffix: A string that will be added at the end of the message.
  :key str icon: A string to show at the front of the message. Default is '❌ '.
  :param **kwargs kwargs: Keyword arguments passed to print_console.
  """
  prefix = kwargs.get('prefix', '')
  suffix = kwargs.get('suffix', '')
  icon = kwargs.get('icon', '❌ ')
  if Spinner is None:
    print_console(message, prefix=f'{prefix}{icon}')
    return
  global buffer
  Spinner.text = message
  Spinner.fail(icon)
  stop_spinner()
  if buffer != '':
    print_console(f'{prefix}{buffer}{suffix}', nl=False)
    buffer = ''


def print_console(message: str, *args, **kwargs):
  """
  Prints a message to console. It avoids problems with the spinner of Yaspin.

  :param str message: The message that will be printed.
  :param *args args: Positional arguments passed to click.secho.
  :param **kwargs kwargs: Keyword arguments passed to click.secho.
  """
  global buffer
  if Spinner is not None:
    # Instead of printing the message we save it in the buffer,
    # then when the spinner stops it will print all the buffer.
    buffer += f'{message}\n'
    return
  click.secho(message, *args, **kwargs)


def print_warning(message: str, *args, **kwargs):
  """
  Prints a message with a specific format and style for warnings.

  :param str message: The message that will be printed.
  :key str prefix: A string that will be added in front fo the message.
  :key str suffix: A string that will be added at the end of the message.
  """
  prefix = kwargs.get('prefix', '')
  suffix = kwargs.get('suffix', '')
  warning_message = f'{prefix}{WARNING_FORMAT.format(message=message)}{suffix}'
  styled_message = click.style(warning_message, fg='yellow')
  print_console(styled_message)


def print_error(message: str, raise_exception: bool = False, **kwargs):
  """
  Prints a message with a specific format and style for errors.

  :param str message: The message that will be printed.
  :param bool raise_exception: If True an PdpException will be raised.
  :key str prefix: A string that will be added in front fo the message.
  :key str suffix: A string that will be added at the end of the message.
  :key Exception exception: An exception that will be raised if raise_exceptions is True.
  :raises PdpException: When raise_exception is True.
  """
  prefix = kwargs.get('prefix', '')
  suffix = kwargs.get('suffix', '')
  error_message = f'{prefix}{ERROR_FORMAT.format(message=message)}{suffix}'
  exception = kwargs.get('exception', PdpException(message=error_message, handled=not raise_exception))
  styled_message = click.style(error_message, fg='red')
  print_console(styled_message, err=True)
  if raise_exception:
    raise exception


def print_exception(exception, **kwargs):
  """
  Prints a message with a specific format and style for errors.

  :param Exception exception: An exception to decide which message to show.
  :key str prefix: A string that will be added in front fo the message.
  :key str suffix: A string that will be added at the end of the message.
  """
  import commons.constants as constants
  prefix = kwargs.get('prefix', '')
  suffix = kwargs.get('suffix', '')
  raise_exception = kwargs.get('raise_exception', False)
  severity = constants.ERROR_SEVERITY
  if hasattr(exception, 'severity'):
    severity = exception.severity
  print_aux = print_error
  match severity:
    case constants.WARNING_SEVERITY:
      print_aux = print_warning
  match exception:
    case req.exceptions.ConnectionError():
      print_aux(f'ConnectionError. Can not connect with {exception.request.url}.', prefix=prefix, suffix=suffix)
    case DataInconsistency():
      print_aux(exception.message, not exception.handled)
    case PdpException():
      print_aux(exception.message, not exception.handled)
    case _:
      print_aux(prefix + EXCEPTION_FORMAT.format(exception=type(exception).__name__, error='') + suffix,
                raise_exception)
