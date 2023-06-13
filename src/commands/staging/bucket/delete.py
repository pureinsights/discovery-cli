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

from commons.console import create_spinner, spinner_change_text, spinner_fail, spinner_ok
from commons.constants import STAGING, URL_GENERIC_BUCKET
from commons.handlers import handle_and_continue
from commons.http_requests import delete


def run(config: dict, buckets: list[str]):
  """
  Delete the given bucket.
  :param dict config: The configuration containing the pdp products' url.
  :param list[str] buckets: The list of names of the buckets to delete.
  """
  for bucket in buckets:
    create_spinner()
    spinner_change_text(f'Deleting bucket {bucket}...')
    bucket_str = click.style(bucket, fg='cyan')
    _, acknowledged = handle_and_continue(
      delete, {'show_exception': True},
      URL_GENERIC_BUCKET.format(config[STAGING], bucket=bucket),
      status_404_as_error=False
    )
    if acknowledged is None or not json.loads(acknowledged).get('acknowledged', False):
      spinner_fail(f"Couldn't delete the bucket {bucket_str}.")
      continue

    spinner_ok(f"Bucket {bucket} was deleted successfully.")
