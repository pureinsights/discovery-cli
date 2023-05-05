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


def test_get(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.execution.command.get`.
  """
  mocker.patch(
    "commands.execution.get.get",
    return_value=b'{"content": ['
                 b'{"id":"execution-id1", "pipelineId":"pipeline-id", "jobId":"job-id", "steps" : [{},{},{}]},'
                 b'{"id":"execution-id2", "pipelineId":"pipeline-id", "jobId":"job-id", "steps" : [{},{},{}]},'
                 b'{"id":"execution-id3", "pipelineId":"pipeline-id", "jobId":"job-id", "steps" : [{},{},{}]}'
                 b']}'
  )
  response = cli.invoke(pdp, ["seed-exec", "get", "--seed", "fake-id"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_get.snapshot')


def test_get_verbose(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.execution.command.get`,
  when the verbose flag is True.
  """
  mocker.patch(
    "commands.execution.get.get",
    return_value=b'{"content": ['
                 b'{"id":"execution-id1", "pipelineId":"pipeline-id", "jobId":"job-id", "steps" : [{},{},{}]},'
                 b'{"id":"execution-id2", "pipelineId":"pipeline-id", "jobId":"job-id", "steps" : [{}]},'
                 b'{"id":"execution-id3", "pipelineId":"pipeline-id", "jobId":"job-id", "steps" : [{},{}]}'
                 b']}'
  )
  response = cli.invoke(pdp, ["seed-exec", "get", "--seed", "fake-id", "-v"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_get_verbose.snapshot')


def test_get_json(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.execution.command.get`,
  when the json flag is True.
  """
  mocker.patch(
    "commands.execution.get.get",
    return_value=b'{"content": ['
                 b'{"id":"execution-id1", "pipelineId":"pipeline-id", "jobId":"job-id", "steps" : [{},{},{}]},'
                 b'{"id":"execution-id2", "pipelineId":"pipeline-id", "jobId":"job-id", "steps" : [{}]},'
                 b'{"id":"execution-id3", "pipelineId":"pipeline-id", "jobId":"job-id", "steps" : [{},{}]}'
                 b']}'
  )
  response = cli.invoke(pdp, ["seed-exec", "get", "--seed", "fake-id", "-j"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_get_json.snapshot')


def test_get_by_ids(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.execution.command.get`,
  when execution ids were provided.
  """
  mocker.patch(
    "commands.execution.get.get",
    side_effect=[
      b'{"id":"execution-id1", "pipelineId":"pipeline-id", "jobId":"job-id", "steps" : [{},{},{}]}',
      b'{"id":"execution-id2", "pipelineId":"pipeline-id", "jobId":"job-id", "steps" : [{}]}',
      b'{"id":"execution-id3", "pipelineId":"pipeline-id", "jobId":"job-id", "steps" : [{},{}]}'
    ]
  )
  response = cli.invoke(pdp, ["seed-exec", "get", "--seed", "fake-id", "--execution", "execution-id1", "--execution",
                              "execution-id2", "--execution", "execution-id3"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_get_by_ids.snapshot')


def test_get_emtpy_executions_by_ids(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.execution.command.get`,
  when the seed doesn't have executions.
  """
  mocker.patch(
    "commands.execution.get.get",
    side_effect=[]
  )
  response = cli.invoke(pdp, ["seed-exec", "get", "--seed", "fake-id", "--execution", "execution-id1", "--execution",
                              "execution-id2", "--execution", "execution-id3"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_get_emtpy_executions_by_ids.snapshot')


def test_get_emtpy_executions(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.execution.command.get`,
  when the seed doesn't have executions.
  """
  mocker.patch(
    "commands.execution.get.get",
    return_value=None
  )
  response = cli.invoke(pdp, ["seed-exec", "get", "--seed", "fake-id", '--asc', 'pipelineId', '--desc', 'status'])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_get_emtpy_executions.snapshot')
