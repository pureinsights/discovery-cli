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
from unittest.mock import call

from commons.constants import STAGING_API_URL, URL_DELETE_TRANSACTION, URL_PURGE_TRANSACTION
from pdp import pdp
from pdp_test import cli


def test_get_transaction(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.transaction.command.get`.
  """
  mocker.patch('commands.staging.transaction.get.get', return_value=b'{'
                                                                    b'"transactionId":"fakeid",'
                                                                    b'"content":['
                                                                    b'{"transaction":"fake"},'
                                                                    b'{"transaction":"fake"},'
                                                                    b'{"transaction":"fake"}'
                                                                    b']'
                                                                    b'}')
  response = cli.invoke(pdp, ["staging", "transaction", "get", "--bucket", "fake-bucket"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_get_transaction.snapshot')


def test_get_transaction_json(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.transaction.command.get`.
  """
  mocker.patch('commands.staging.transaction.get.get', return_value=b'{'
                                                                    b'"transactionId":"fakeid",'
                                                                    b'"content":['
                                                                    b'{"transaction":"fake"},'
                                                                    b'{"transaction":"fake"},'
                                                                    b'{"transaction":"fake"}'
                                                                    b']'
                                                                    b'}')
  response = cli.invoke(pdp, ["staging", "transaction", "get", "--bucket", "fake-bucket", '-j'])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_get_transaction_json.snapshot')


def test_delete_no_all_or_id_flags(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.transaction.command.delete`,
  when the user don't provide the flag --all or --id.
  """
  delete_mock = mocker.patch("commands.staging.transaction.delete.delete", return_value='{"acknowledged": false}')
  mocker.patch("commands.staging.transaction.delete.create_spinner")
  response = cli.invoke(pdp, ["staging", "transaction", "delete", "--bucket", "fake-bucket"])
  assert delete_mock.call_count == 0
  assert response.exit_code == 1
  snapshot.assert_match(response.exception.message, 'test_delete_no_all_or_id_flags.snapshot')


def test_delete_all_transactions(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.transaction.command.delete`.
  """
  delete_mock = mocker.patch("commands.staging.transaction.delete.delete", return_value='{"acknowledged": true}')
  mocker.patch("commands.staging.transaction.delete.create_spinner")
  response = cli.invoke(pdp, ["staging", "transaction", "delete", "--bucket", "fake-bucket", '--all'])
  delete_mock.assert_called_once_with(URL_PURGE_TRANSACTION.format(STAGING_API_URL, bucket="fake-bucket"))
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_delete_all_transactions.snapshot')


def test_delete_all_transactions_failed(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.transaction.command.delete`,
  when try to delete all the transactions and fails.
  """
  delete_mock = mocker.patch("commands.staging.transaction.delete.delete", return_value='{"acknowledged": false}')
  mocker.patch("commands.staging.transaction.delete.create_spinner")
  response = cli.invoke(pdp, ["staging", "transaction", "delete", "--bucket", "fake-bucket", '--all'])
  delete_mock.assert_called_once_with(URL_PURGE_TRANSACTION.format(STAGING_API_URL, bucket="fake-bucket"))
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_delete_all_transactions_failed.snapshot')


def test_delete_specific_transactions(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.transaction.command.delete`,
  when the user provides one or more --id flags.
  """
  delete_mock = mocker.patch("commands.staging.transaction.delete.delete",
                             side_effect=['{"acknowledged": true}', '{"acknowledged": false}',
                                          '{"acknowledged": true}'])
  mocker.patch("commands.staging.transaction.delete.create_spinner")
  response = cli.invoke(pdp, ["staging", "transaction", "delete", "--bucket", "fake-bucket", '--id', 'fake-id1', '--id',
                              'fake-id2', '--id', 'fake-id3'])
  assert delete_mock.call_args_list == [
    call(URL_DELETE_TRANSACTION.format(STAGING_API_URL, bucket="fake-bucket", transaction='fake-id1'),
         params={'transactionId': 'fake-id1'}),
    call(URL_DELETE_TRANSACTION.format(STAGING_API_URL, bucket="fake-bucket", transaction='fake-id2'),
         params={'transactionId': 'fake-id2'}),
    call(URL_DELETE_TRANSACTION.format(STAGING_API_URL, bucket="fake-bucket", transaction='fake-id3'),
         params={'transactionId': 'fake-id3'})
  ]
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_delete_specific_transactions.snapshot')
