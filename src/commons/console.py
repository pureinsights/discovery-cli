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
from yaspin.spinners import Spinners

from commons.constants import ERROR_FORMAT, EXCEPTION_FORMAT, WARNING_FORMAT
from commons.custom_classes import DataInconsistency, PdpException

Spinner: Union[Yaspin, None] = None  # An instance of the Yaspin spinner
buffer: str = ''  # This is a buffer to store messages that want to be printed after the spinner stops
# This two list are useful to know if an error or warning has been shown to the user
printed_warnings = []
printed_errors = []
printed_exceptions = []
# These are variables to control the error and warning printing
is_warnings_suppressed = False
is_errors_suppressed = False


def suppress_warnings(suppress: bool):
  """
  Avoids to any warning be printed in console.
  """
  global is_warnings_suppressed
  is_warnings_suppressed = suppress


def suppress_errors(suppress: bool):
  """
  Avoids to any error or exception be printed in console.
  """
  global is_errors_suppressed
  is_errors_suppressed = suppress


def get_number_errors_exceptions():
  """
  Returns a sum of the number of errors and the number of exceptions that happened.
  """
  global printed_errors, printed_exceptions
  return len(printed_errors) + len(printed_exceptions)


def create_spinner(*args, **kwargs):
  """
  Creates and starts a new spinner.

  :param *args args: The positional arguments passed to Yaspin.yaspin.
  :param **kwargs kwargs: The keyword arguments passed to Yaspin.yaspin.
  """
  global Spinner
  if Spinner is not None:
    stop_spinner()
  if len(args) <= 0:
    Spinner = yaspin(Spinners.dots12, **kwargs)
  else:
    Spinner = yaspin(*args, **kwargs)
  Spinner.start()


def stop_spinner(prefix: str = '', suffix: str = ''):
  """
  Stops and clean the instance of the spinner.
  """
  global Spinner, buffer
  if Spinner is None:
    return
  Spinner.stop()
  Spinner = None
  if buffer != '':
    print_console(f'{prefix}{buffer}{suffix}', nl=False)
    buffer = ''


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
  Will stop the spinner and sets a text to the success status.

  :param str message: The message that will be showed.
  :key str prefix: A string that will be added in front fo the message.
  :key str suffix: A string that will be added at the end of the message.
  :key str icon: A string to show at the front of the message. Default is '✔ '.
  :param **kwargs kwargs: Keyword arguments passed to print_console.
  """
  icon = kwargs.get('icon', '✔ ')
  prefix = kwargs.get('prefix', '')
  if Spinner is None:
    print_console(message, prefix=f'{prefix}{icon}')
    return
  suffix = kwargs.get('suffix', '')
  Spinner.text = message
  if icon is not None and icon != '':
    Spinner.ok(icon)
  stop_spinner(prefix, suffix)


def spinner_fail(message: str, **kwargs):
  """
  Will stop the spinner and sets a text to the failed status.

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
  stop_spinner(prefix, suffix)


def print_console(message: any, *args, **kwargs):
  """
  Prints a message to console. Avoids problems with the spinner of Yaspin.

  :param any message: The message that will be printed.
  :param *args args: Positional arguments passed to click.secho.
  :param **kwargs kwargs: Keyword arguments passed to click.secho.
  """
  global buffer
  prefix = kwargs.get('prefix', '')
  suffix = kwargs.get('suffix', '')
  nl = kwargs.get('nl', True)
  new_line = ''
  if nl:
    new_line = '\n'
  if Spinner is not None:
    # Instead of printing the message we save it in the buffer,
    # then when the spinner stops it will print all the buffer.
    buffer += f'{prefix}{message}{suffix}{new_line}'
    return
  kwargs.pop('prefix', None)
  kwargs.pop('suffix', None)
  kwargs.pop('nl', None)
  click.secho(f'{prefix}{message}{suffix}', *args, nl=nl, **kwargs)


def print_warning(message: str, *args, **kwargs):
  """
  Prints a message with a specific format and style for warnings.

  :param str message: The message that will be printed.
  :key str prefix: A string that will be added in front fo the message.
  :key str suffix: A string that will be added at the end of the message.
  """
  global printed_warnings
  printed_warnings += [message]
  if is_warnings_suppressed:
    return
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
  global printed_errors
  printed_errors += [message]
  if is_errors_suppressed:
    return
  prefix = kwargs.get('prefix', '')
  suffix = kwargs.get('suffix', '')
  error_message = f'{prefix}{ERROR_FORMAT.format(message=message)}{suffix}'
  exception = kwargs.get('exception', PdpException(message=message, handled=not raise_exception))
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
  global printed_exceptions
  printed_exceptions += [exception]
  if is_errors_suppressed:
    return
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
      print_aux(f'ConnectionError. Can not connect to {exception.request.url}.', prefix=prefix, suffix=suffix)
    case DataInconsistency():
      print_aux(exception.message, not exception.handled)
    case PdpException():
      print_aux(exception.message, not exception.handled)
    case FileNotFoundError():
      print_aux(f'{exception.strerror}: {exception.filename}')
    case _:
      print_aux(prefix + EXCEPTION_FORMAT.format(exception=type(exception).__name__, error='') + suffix,
                raise_exception)


def verbose(**kwargs):
  """
  Will execute any of "verbose_func" or "not_verbose_func" based on "verbose" flag. And return
  whatever the function returns. Helpful to manage more complex behaviors to a verbose command, rather than
  just print some text in console.
  :key bool verbose: The flag tha defines which function will be called.
  :key Callable verbose_func: The function that will be called if the verbose flag is True. If is not callable
                              then the value it will be returned instead.
  :key Callable not_verbose_func: The function that will be called if the verbose flag is False. If is not callable
                              then the value it will be returned instead.
  :rtype: any
  :return: Returns whatever the function called based on the verbose flag returns, if is not callable
           the argument will be return instead.
  """
  verbose = kwargs.get('verbose', False)
  verbose_func = kwargs.get('verbose_func', lambda: None)
  not_verbose_func = kwargs.get('not_verbose_func', lambda: None)
  if verbose:
    return verbose_func() if callable(verbose_func) else verbose_func
  return not_verbose_func() if callable(not_verbose_func) else not_verbose_func
