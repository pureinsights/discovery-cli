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

import click

from commons.console import create_spinner, print_console, print_warning, spinner_change_text, \
  spinner_fail, spinner_ok
from commons.constants import STAGING, URL_GENERIC_BUCKET, URL_GENERIC_ITEM
from commons.custom_classes import PdpException
from commons.handlers import handle_and_continue, handle_and_exit
from commons.http_requests import delete, get


def run(config: dict, bucket: str, item_ids: list[str], filter: bool):
  """
  Delete one or all the items of a bucket, without delete the bucket.
  :param dict config: The configuration containing the pdp products' url.
  :param str bucket: The name of the bucket to delete the items.
  :param list[str] item_ids: A list of ids of the items to delete. If empty then all the items will be deleted.
  :param bool filter: A flag to open a text editor to capture the filter criteria.
  """
  if filter:
    filter_criteria: str | None = click.edit()
    if filter_criteria is None:
      print_warning('If you wrote some criteria, please be sure you save before close the text editor.')
      raise PdpException(message="The filter criteria can't be empty.")

    res = delete(f'{URL_GENERIC_BUCKET.format(config[STAGING], bucket=bucket)}/filter',
                 json=json.loads(filter_criteria))
    if res is None:
      return print_console("Couldn't delete any item.")

    res = json.loads(res)
    if len(res) <= 0:
      return print_console("No items were deleted.")

    for item in res:
      print_console(
        f"Item {click.style(item.get('contentId'), fg='green')} deleted successfully. "
        f"Transaction id {click.style(item.get('transactionId', None), fg='cyan')}"
      )
    return

  if len(item_ids) <= 0:
    _, res = handle_and_exit(get, {'show_exception': True, 'message': "Couldn't delete the items."},
                             URL_GENERIC_BUCKET.format(config[STAGING], bucket=bucket),
                             params={'contentType': 'METADATA'})
    res = json.loads(res)
    if res.get('empty', False):
      return print_console(f"The bucket {click.style(bucket, fg='cyan')} is empty.")

    item_ids = [content['contentId'] for content in res.get('content', [])]

  for item in item_ids:
    item_str = click.style(item, fg='green')
    create_spinner()
    spinner_change_text(f"Deleting item {item_str}...")
    handler_config = {'show_exception': True}
    _, res = handle_and_continue(delete, handler_config,
                                 URL_GENERIC_ITEM.format(config[STAGING], bucket=bucket, content_id=item))
    if res is None:
      spinner_fail(f"Couldn't delete the item {item_str}.")
      continue

    res = json.loads(res)
    spinner_ok(
      f"Item {item_str} deleted successfully. Transaction id {click.style(res.get('transactionId', None), fg='cyan')}."
    )
