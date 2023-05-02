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


def test_upload_file(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.core.command.upload`.
  """
  mocker.patch("commands.core.file.upload.put", return_value=b'{"acknowledged": true }')
  mocker.patch("commands.core.file.upload.read_binary_file", return_value="")
  mocker.patch("commons.file_system.raise_file_not_found_error")
  response = cli.invoke(pdp, ["core", "file", "upload", "--name", "fake-name", "--path", "fake-path"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_upload_file.snapshot')


def test_upload_file_without_name(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.core.command.upload`,
  if the user don't provide a --name.
  """
  mocker.patch("commands.core.file.upload.put", return_value=b'{"acknowledged": true }')
  mocker.patch("commands.core.file.upload.read_binary_file", return_value="")
  mocker.patch("commons.file_system.raise_file_not_found_error")
  response = cli.invoke(pdp, ["core", "file", "upload", "--path", "fake-path"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_upload_file_without_name.snapshot')


def test_upload_no_uploaded(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.core.command.upload`,
  if the user don't provide a --name.
  """
  mocker.patch("commands.core.file.upload.put", return_value=b'{"acknowledged": false }')
  mocker.patch("commands.core.file.upload.read_binary_file", return_value="")
  mocker.patch("commons.file_system.raise_file_not_found_error")
  response = cli.invoke(pdp, ["core", "file", "upload", "--path", "fake-path"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_upload_no_uploaded.snapshot')


def test_upload_failed(mocker, snapshot, mock_custom_exception):
  """
  Test the command defined in :func:`src.commands.core.command.upload`,
  if the user don't provide a --name.
  """
  mocker.patch("commands.core.file.upload.put", side_effect=lambda *args, **kwargs: mock_custom_exception(Exception))
  mocker.patch("commands.core.file.upload.read_binary_file", return_value="")
  mocker.patch("commons.file_system.raise_file_not_found_error")
  response = cli.invoke(pdp, ["core", "file", "upload", "--path", "fake-path"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_upload_failed.snapshot')
