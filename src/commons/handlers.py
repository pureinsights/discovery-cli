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
import json

import requests as req
import requests.exceptions

from commons.console import print_error, print_exception, print_warning, stop_spinner
from commons.constants import EXCEPTION_FORMAT
from commons.custom_classes import PdpException


def handle_exceptions(func: callable, *args, **kwargs):
  """
  Tries to execute the function 'func', if some exception happens will handle it and use the
  print_exception function to print a specific message based on the exception type.

  :param callable func: The function th execute.
  :param tuple[Any, ...] args: The positional arguments of the function 'func'.
  :param dict[str, Any] kwargs: The key-value arguments of the function 'func'.
  """
  try:
    func(*args, **kwargs)
  except Exception as exception:
    if not hasattr(exception, 'handled'):
      print_exception(exception)
    elif not exception.handled:
      exception.handled = True
      print_exception(exception)

  finally:
    stop_spinner()


def handle_http_response(res: req.Response, status_404_as_error: bool = True) -> any:
  """
  Handle the responses for any http call. Raise an exception if the status code is not 2xx.

  :param requests.Response res: The response to be handled.
  :param bool status_404_as_error: If true status 404 will raise an error, if not it will return None
  :return: Returns the content of the response.
  :rtype: Any
  :raises HTTPError: When the status is not a 2xx response.
  """
  try:
    if res.status_code == 404 and not status_404_as_error:
      return None
    res.raise_for_status()  # raises an exception when the status is not a 2xx response
    if res.status_code != 204:
      return res.content
    return None
  except req.exceptions.HTTPError as exception:
    content = exception.response.content
    if content is None:
      raise exception
    response = json.loads(content.decode('utf-8'))
    errors = '\n\t'.join(response.get('errors', []))
    method = exception.request.method
    url = exception.request.url
    content = {
      'status': res.status_code,
      'errors': errors
    }
    raise PdpException(message=f"Could not '{method}' to {url} due to:\n\t{errors}", content=content)


def handle_and_exit(func: callable, handle_and_exit_params: dict, *args, **kwargs) -> tuple[bool, any]:
  """
  Tries to execute the given function, if an exception happens will be handled and print the given message.
  If raise_exception is True will raise the same Exception that it handle.
  It's helpful to print a specific message for a potential exception.

  :param callable func: The function that will be executed.
  :param params handle_and_exit_params: A dict containing the params for the handler function.
  :param *args args: The positional arguments for the 'func' function.
  :param **kwargs kwargs: The key-value arguments for the 'func' function.
  :type params: str message: The message to print if an exception happens.
                str prefix: A str added front of the message.
                str suffix: A str added at the end of the message.
                bool show_exception: Uses the print_exception to
                bool raise_exception: Raises the same exception that was handled.
  :rtype: None
  :raises Exception: Can raise the same exception handled in order to be handled by an upper handler as well.
  """
  error_message = handle_and_exit_params.get('message', None)
  prefix = handle_and_exit_params.get('prefix', '')
  suffix = handle_and_exit_params.get('suffix', '')
  show_exception = handle_and_exit_params.get('show_exception', False)
  try:
    return True, func(*args, **kwargs)

  except Exception as error:
    error = handle_and_exit_params.get('exception', error)
    if hasattr(error, 'handled'):
      error.handled = True
    if show_exception:
      print_exception(error, prefix=prefix, suffix=suffix)

    if error_message is not None:
      print_error(error_message, True, prefix=prefix, suffix=suffix)

    if handle_and_exit_params.get('exception', None) is not None:
      raise error

    raise PdpException(message=EXCEPTION_FORMAT.format(exception=type(error).__name__, error=''),
                       handled=show_exception or error_message is not None)


def handle_and_continue(func: callable, handle_and_continue_params: dict, *args, **kwargs):
  """
  Tries to execute the given function, if an exception happens will be handled and print the given message.
  If an exception occurs will be handled and then continue the execution.


  :param callable func: The function that will be executed.
  :param params handle_and_continue_params: A dict containing the params for the handler function.
  :param *args args: The positional arguments for the 'func' function.
  :param **kwargs kwargs: The key-value arguments for the 'func' function.
  :type params: str message: The message to print if an exception happens.
                str prefix: A str added front of the message.
                str suffix: A str added at the end of the message.
                bool show_exception: Uses the print_exception to
  :rtype: tuple[bool, Any]
  :return: Returns a tuple containing in the first position a boolean that represents if there was an exception or
           if the function was executed successfully. The second one is the return of the function or None if an
           exception happened.
  :raises Exception: Can raise the same exception handled in order to be handled by an upper handler as well.
  """
  error_message = handle_and_continue_params.get('message', None)
  prefix = handle_and_continue_params.get('prefix', '')
  suffix = handle_and_continue_params.get('suffix', '')
  show_exception = handle_and_continue_params.get('show_exception', False)
  warning = handle_and_continue_params.get('warning', False)
  try:
    return True, func(*args, **kwargs)

  except Exception as error:

    if error_message is not None:
      if warning:
        print_warning(error_message, prefix=prefix, suffix=suffix)
      else:
        print_error(error_message, False, prefix=prefix, suffix=suffix)

    if show_exception:
      if hasattr(error, 'handled'):
        error.handled = True
      print_exception(error, prefix=prefix, suffix=suffix)

    return False, None
