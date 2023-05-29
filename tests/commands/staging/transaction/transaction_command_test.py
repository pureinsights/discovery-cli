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
