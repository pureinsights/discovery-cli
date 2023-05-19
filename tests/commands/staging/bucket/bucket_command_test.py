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


def test_delete(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.bucket.command.delete`.
  """
  mocker.patch("commands.staging.bucket.delete.create_spinner")
  mocker.patch("commands.staging.bucket.delete.delete", side_effect=[None, '{"acknowledged":true}'])
  response = cli.invoke(pdp,
                        ["staging", "bucket", "delete", "--bucket", "fake-bucket1", "--bucket", "fake-bucket2"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_delete.snapshot')


def test_status(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.bucket.command.status`.
  """
  mocker.patch("commands.staging.bucket.status.get", return_value=b'{"content": ['
                                                                  b'{"bucket": "fake1"},'
                                                                  b'{"bucket": "fake2"},'
                                                                  b'{"bucket": "fake3"}'
                                                                  b']}')
  response = cli.invoke(pdp,
                        ["staging", "bucket", "status"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_status.snapshot')


def test_status_json(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.bucket.command.status`,
  when the user wants the result in JSON format.
  """
  mocker.patch("commands.staging.bucket.status.get", return_value=b'{"content": ['
                                                                  b'{"bucket": "fake1"},'
                                                                  b'{"bucket": "fake2"},'
                                                                  b'{"bucket": "fake3"}'
                                                                  b']}')
  response = cli.invoke(pdp,
                        ["staging", "bucket", "status", "--json"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_status_json.snapshot')


def test_status_specific_bucket(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.bucket.command.status`,
  when the user wants the status for a specific bucket.
  """
  mocker.patch("commands.staging.bucket.status.get", return_value=b'{"bucket": "fake1"}')
  response = cli.invoke(pdp,
                        ["staging", "bucket", "status", "--bucket", "fake1", "--asc", "bucket", "--desc", "bucket"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_status_specific_bucket.snapshot')


def test_status_failed(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.bucket.command.status`,
  when the get method fails.
  """
  mocker.patch("commands.staging.bucket.status.get", return_value=None)
  response = cli.invoke(pdp,
                        ["staging", "bucket", "status", "--bucket", "fake1", "--asc", "bucket", "--desc", "bucket"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_status_specific_bucket.snapshot')
