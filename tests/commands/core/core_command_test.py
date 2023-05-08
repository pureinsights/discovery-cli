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


def test_core(snapshot):
  """
  Test the command defined in :func:`src.commands.core.command.core`.
  """
  response = cli.invoke(pdp, ["core"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_core.snapshot')


def test_search(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.core.command.search`.
  """
  mocker.patch('commands.core.search.create_spinner')
  mocker.patch('commands.core.search.post')
  mocker.patch('commands.core.search.json.loads', return_value={"content": [
    {'id': '6376af03-1af2-41a2-aef6-62aefc73a870', 'name': 'fake-name1', 'description': None,
     'active': True},
    {'id': '6376af03-1af2-41a2-aef6-62aefc73a871', 'name': 'fake-name2', 'description': 'Entity 2',
     'active': False}
  ]})
  response = cli.invoke(pdp, ["core", "search", "--asc", "id", "--desc", "name", "-q", "fake-entity"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_search.snapshot')


def test_search_json(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.core.command.search`,
  when --json flag is active.
  """
  mocker.patch('commands.core.search.create_spinner')
  mocker.patch('commands.core.search.post')
  mocker.patch('commands.core.search.json.loads', return_value={"content": [
    {'id': '6376af03-1af2-41a2-aef6-62aefc73a870', 'name': 'fake-name1', 'description': None,
     'active': True},
    {'id': '6376af03-1af2-41a2-aef6-62aefc73a871', 'name': 'fake-name2', 'description': 'Entity 2',
     'active': False}
  ]})
  response = cli.invoke(pdp, ["core", "search", "--json"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_search_json.snapshot')


def test_search_not_found(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.core.command.search`,
  when no entities were found.
  """
  mocker.patch('commands.core.search.create_spinner')
  mocker.patch('commands.core.search.post')
  mocker.patch('commands.core.search.json.loads', return_value={"content": []})
  response = cli.invoke(pdp, ["core", "search", "--json"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_search_not_found.snapshot')


def test_search_no_response(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.core.command.search`,
  when no entities were found.
  """
  mocker.patch('commands.core.search.create_spinner')
  mocker.patch('commands.core.search.post', return_value=None)
  response = cli.invoke(pdp, ["core", "search"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_search_no_response.snapshot')


def test_log_level(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.core.command.log_level`.
  """
  post_mock = mocker.patch("commands.core.log_level.post", return_value=b'{ "acknowledged": true }')
  response = cli.invoke(pdp, ["core", "log-level", "--component", "fake-component", "--level", "warn"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_log_level.snapshot')


def test_log_level_failed(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.core.command.log_level`,
  when the response is None or acknowledged false.
  """
  post_mock = mocker.patch("commands.core.log_level.post", return_value=b'{ "acknowledged": false }')
  response = cli.invoke(pdp, ["core", "log-level", "--component", "fake-component", "--level", "warn"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_log_level_failed.snapshot')
