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
from commons.constants import STAGING, URL_DELETE_TRANSACTION, URL_PURGE_TRANSACTION
from commons.handlers import handle_and_continue
from commons.http_requests import delete


def run(config: dict, bucket: str, transaction_ids: list[str], purge: bool):
  """
  Delete one or more transactions for a given bucket.
  :param dict config: The configuration containing the pdp products' url.
  :param str bucket: The name of the bucket to delete the transactions.
  :param str transaction_ids: The list of transactions to delete.
  :param bool purge: This is a boolean flag. It will call the purge endpoint instead of delete.
  """
  if len(transaction_ids) <= 0:
    create_spinner()
    spinner_change_text(f"Deleting all the transactions for the bucket {click.style(bucket, fg='cyan')}")
    res = delete(URL_PURGE_TRANSACTION.format(config[STAGING], bucket=bucket))
    if res is None or not json.loads(res).get('acknowledged', False):
      return spinner_fail("Could not delete all the transactions.")

    return spinner_ok("All the transactions has been deleted.")

  url = URL_PURGE_TRANSACTION if purge else URL_DELETE_TRANSACTION
  for transaction_id in transaction_ids:
    create_spinner()
    spinner_change_text(f"Deleting the transaction {click.style(transaction_id, fg='green')} for "
                        f"{click.style(bucket, fg='cyan')} bucket.")
    _, res = handle_and_continue(delete, {'show_exception': True},
                                 url.format(config[STAGING], bucket=bucket, transaction=transaction_id),
                                 params={'transactionId': transaction_id})
    if res is None or not json.loads(res).get('acknowledged', False):
      spinner_fail(f"Could not delete the transaction {click.style(transaction_id, fg='green')}.")
      continue

    spinner_ok(f"Transaction {click.style(transaction_id, fg='green')} has been deleted.")
