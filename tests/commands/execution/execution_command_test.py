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


def test_start(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.execution.command.start`.
  """
  mocker.patch("commands.execution.start.post", return_value=b'{"id": "fake-execution-id"}')
  response = cli.invoke(pdp, ["seed-exec", "start", "--seed", "fake-id", "--scan-type", "incremental"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_start.snapshot')


def test_reset(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.execution.command.reset`.
  """
  mocker.patch("commands.execution.reset.post", return_value=b'{"acknowledged": true}')
  response = cli.invoke(pdp, ["seed-exec", "reset", "--seed", "fake-id"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_reset.snapshot')


def test_reset_failed(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.execution.command.reset`,
  when could not reset the seed.
  """
  mocker.patch("commands.execution.reset.post", return_value=b'{"acknowledged": false}')
  response = cli.invoke(pdp, ["seed-exec", "reset", "--seed", "fake-id"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_reset_failed.snapshot')


def test_control(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.execution.command.control`.
  """
  mocker.patch("commands.execution.control.put", return_value=b'{"acknowledged": true}')
  response = cli.invoke(pdp, ["seed-exec", "control", "--seed", "fake-id", "--action", "HALT"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_control.snapshot')


def test_control_failed(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.execution.command.control`,
  when the api returns an acknowledged false.
  """
  mocker.patch("commands.execution.control.put", return_value=b'{"acknowledged": false}')
  response = cli.invoke(pdp, ["seed-exec", "control", "--seed", "fake-id", "--action", "RESUME"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_control_failed.snapshot')
