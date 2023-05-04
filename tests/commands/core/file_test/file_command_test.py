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
import os

from pdp import pdp
from pdp_test import cli


def test_file(snapshot):
  """
  Test the command defined in :func:`src.commands.core.file.command.file`.
  """
  response = cli.invoke(pdp, ["core", "file"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_file.snapshot')


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


def test_download_file(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.core.command.download`.
  """
  mocker.patch("commands.core.file.download.get", return_value=b'{"acknowledged": true }')
  mock_write = mocker.patch("commands.core.file.download.write_binary_file")
  response = cli.invoke(pdp, ["core", "file", "download", "--name", "seeds", "--path", "./fake-path/seed.json"])
  mock_write.assert_called_once_with("./fake-path/seed.json", b'{"acknowledged": true }')
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_download_file.snapshot')


def test_download_file_without_rename(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.core.command.download`,
  when the given path is a folder.
  """
  mocker.patch("commands.core.file.download.get", return_value=b'{"acknowledged": true }')
  mocker.patch("commands.core.file.download.os.path.isdir", return_value=True)
  mock_write = mocker.patch("commands.core.file.download.write_binary_file")
  response = cli.invoke(pdp, ["core", "file", "download", "--name", "seeds", "--path", "./fake-path/"])
  mock_write.assert_called_once_with("./fake-path/seeds", b'{"acknowledged": true }')
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_download_file_without_rename.snapshot')


def test_download_file_without_path(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.core.command.download`,
  when no path was entered.
  """
  mocker.patch("commands.core.file.download.get", return_value=b'{"acknowledged": true }')
  mocker.patch("commands.core.file.download.os.path.isdir", return_value=True)
  mock_write = mocker.patch("commands.core.file.download.write_binary_file")
  response = cli.invoke(pdp, ["core", "file", "download", "--name", "seeds"])
  expected_path = os.path.join(".", "seeds")
  mock_write.assert_called_once_with(expected_path, b'{"acknowledged": true }')
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_download_file_without_path.snapshot')


def test_download_file_within_pdp_project(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.core.command.download`,
  when the given path is a PDP project.
  """
  mocker.patch("commands.core.file.download.get", return_value=b'{"acknowledged": true }')
  mocker.patch("commands.core.file.download.os.path.isdir", return_value=True)
  mocker.patch("commands.core.file.download.has_pdp_project_structure", return_value=True)
  mock_write = mocker.patch("commands.core.file.download.write_binary_file")
  response = cli.invoke(pdp, ["core", "file", "download", "--name", "seeds"])
  expected_path = os.path.join(".", "Core", "files", "seeds")
  mock_write.assert_called_once_with(expected_path, b'{"acknowledged": true }')
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_download_file_within_pdp_project.snapshot')


def test_delete_files(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.core.command.delete`,
  when the given path is a PDP project.
  """
  mocker.patch("commands.core.file.delete.create_spinner")
  mocker.patch("commands.core.file.delete.has_pdp_project_structure", side_effect=[True, True, False])
  mocker.patch("commands.core.file.delete.handle_and_continue",
               side_effect=[(True, b'{"acknowledged": true }'), (True, b'{"acknowledged": false }'), (False, None)])
  ok_mock = mocker.patch("commands.core.file.delete.spinner_ok")
  fail_mock = mocker.patch("commands.core.file.delete.spinner_fail")
  response = cli.invoke(pdp, ["core", "file", "delete", "--name", "fake-name1", "--name", "fake-name2", "--name",
                              "fake-name3"])
  assert response.exit_code == 0
  assert ok_mock.call_count == 1
  assert fail_mock.call_count == 2


def test_delete_files_locally(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.core.command.delete`.
  """
  mocker.patch("commands.core.file.delete.create_spinner")
  mocker.patch("commands.core.file.delete.os.remove")
  mocker.patch("commands.core.file.delete.has_pdp_project_structure", side_effect=[True, True, False])
  mocker.patch("commands.core.file.delete.handle_and_continue",
               side_effect=[(True, b'{"acknowledged": true }'), (True, b'{"acknowledged": false }'), (False, None)])
  ok_mock = mocker.patch("commands.core.file.delete.spinner_ok")
  fail_mock = mocker.patch("commands.core.file.delete.spinner_fail")
  response = cli.invoke(pdp, ["core", "file", "delete", "--name", "fake-name1", "--name", "fake-name2", "--name",
                              "fake-name3", "--local"])
  assert response.exit_code == 0
  assert ok_mock.call_count == 1
  assert fail_mock.call_count == 2


def test_list(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.core.command.ls`.
  """
  mocker.patch("commands.core.file.list.get", return_value=b'["fake-file", "fake-file2"]')
  response = cli.invoke(pdp, ["core", "file", "ls"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_list.snapshot')


def test_list_empty(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.core.command.ls`,
  when the list of files is empty.
  """
  mocker.patch("commands.core.file.list.get", return_value=b'[]')
  response = cli.invoke(pdp, ["core", "file", "ls"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_list_empty.snapshot')


def test_list_json(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.core.command.ls`,
  when the --json flag is True.
  """
  mocker.patch("commands.core.file.list.get", return_value=b'["fake-file", "fake-file2"]')
  response = cli.invoke(pdp, ["core", "file", "ls", "--json"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_list_json.snapshot')
