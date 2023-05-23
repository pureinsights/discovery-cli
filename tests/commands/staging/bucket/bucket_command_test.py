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


def test_get_bucket(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.bucket.command.get`.
  """
  mocker.patch('commands.staging.bucket.get.get', return_value=b'{'
                                                               b'"transactionId": "645e9901c8e408abf7e1a194",'
                                                               b'"timestamp": "2023-05-12T19:52:33.245Z",'
                                                               b'"action": "ADD",'
                                                               b'"bucket": "test",'
                                                               b'"contentId": "test3",'
                                                               b'"checksum": "23b5c58d597754037351ebdc5497882b"'
                                                               b'}')
  response = cli.invoke(pdp, ["staging", "bucket", "get", "--bucket", "fake"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_get_bucket.snapshot')


def test_get_bucket_filter(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.bucket.command.get`,
  when the user wants to filter the data.
  """
  mock_post = mocker.patch(
    'commands.staging.bucket.get.post',
    return_value=b'{'
                 b'"transactionId": "645e9901c8e408abf7e1a194",'
                 b'"timestamp": "2023-05-12T19:52:33.245Z",'
                 b'"action": "ADD",'
                 b'"bucket": "test",'
                 b'"contentId": "test3",'
                 b'"checksum": "23b5c58d597754037351ebdc5497882b"'
                 b'}'
  )
  response = cli.invoke(pdp, ["staging", "bucket", "get", "--bucket", "fake", "--page", "2", "--token", "fake-token",
                              "--content-type",
                              "CONTENT"])
  assert response.exit_code == 0
  mock_post.assert_called_once_with(
    'http://localhost:8081/content/fake/filter',
    query_params={'token': 'fake-token', 'contentType': 'CONTENT', 'size': None}
  )
  snapshot.assert_match(response.output, 'test_get_bucket_filter.snapshot')


def test_get_bucket_pagination(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.bucket.command.get`,
  when the user wants to use pagination on the data.
  """
  mock_post = mocker.patch(
    'commands.staging.bucket.get.post',
    return_value=b'{'
                 b'"transactionId": "645e9901c8e408abf7e1a194",'
                 b'"timestamp": "2023-05-12T19:52:33.245Z",'
                 b'"action": "ADD",'
                 b'"bucket": "test",'
                 b'"contentId": "test3",'
                 b'"checksum": "23b5c58d597754037351ebdc5497882b"'
                 b'}'
  )
  response = cli.invoke(pdp, ["staging", "bucket", "get", "--bucket", "fake", "--page", "2", "--asc",
                              "CONTENT", "--desc", "fake-property"])
  assert response.exit_code == 0
  mock_post.assert_called_once_with(
    'http://localhost:8081/content/fake/query',
    query_params={'page': '2', 'sort': ['CONTENT,asc', 'fake-property,desc'], 'size': None}
  )
  snapshot.assert_match(response.output, 'test_get_bucket_filter.snapshot')


def test_get_bucket_json(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.bucket.command.get`.
  """
  mocker.patch('commands.staging.bucket.get.get', return_value=b'{'
                                                               b'"transactionId": "645e9901c8e408abf7e1a194",'
                                                               b'"timestamp": "2023-05-12T19:52:33.245Z",'
                                                               b'"action": "ADD",'
                                                               b'"bucket": "test",'
                                                               b'"contentId": "test3",'
                                                               b'"checksum": "23b5c58d597754037351ebdc5497882b"'
                                                               b'}')
  response = cli.invoke(pdp, ["staging", "bucket", "get", "--bucket", "fake", "--json"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_get_bucket_json.snapshot')
