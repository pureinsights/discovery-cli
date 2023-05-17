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


def test_add_item(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.item.command.add`.
  """
  mocker.patch("commands.staging.item.add.put", return_value=b'{'
                                                             b'"transactionId": "645e9901c8e408abf7e1a194",'
                                                             b'"timestamp": "2023-05-12T19:52:33.245Z",'
                                                             b'"action": "ADD",'
                                                             b'"bucket": "test",'
                                                             b'"contentId": "test3",'
                                                             b'"checksum": "23b5c58d597754037351ebdc5497882b"'
                                                             b'}')
  mocker.patch("commands.staging.item.add.read_binary_file", return_value=b"{}")
  response = cli.invoke(pdp, ["staging", "item", "add", "--bucket", "fake-bucket", "--item-id", "fake-id", "--file",
                              "fake-path"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_add_item.snapshot')


def test_add_item_verbose(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.item.command.add`.
  """
  mocker.patch("commands.staging.item.add.put", return_value=b'{'
                                                             b'"transactionId": "645e9901c8e408abf7e1a194",'
                                                             b'"timestamp": "2023-05-12T19:52:33.245Z",'
                                                             b'"action": "ADD",'
                                                             b'"bucket": "test",'
                                                             b'"contentId": "test3",'
                                                             b'"checksum": "23b5c58d597754037351ebdc5497882b"'
                                                             b'}')
  mocker.patch("commands.staging.item.add.read_binary_file", return_value=b"{}")
  response = cli.invoke(pdp, ["staging", "item", "add", "--bucket", "fake-bucket", "--item-id", "fake-id", "--file",
                              "fake-path", "-v"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_add_item_verbose.snapshot')


def test_add_item_json(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.item.command.add`.
  """
  mocker.patch("commands.staging.item.add.put", return_value=b'{'
                                                             b'"transactionId": "645e9901c8e408abf7e1a194",'
                                                             b'"timestamp": "2023-05-12T19:52:33.245Z",'
                                                             b'"action": "ADD",'
                                                             b'"bucket": "test",'
                                                             b'"contentId": "test3",'
                                                             b'"checksum": "23b5c58d597754037351ebdc5497882b"'
                                                             b'}')
  mocker.patch("commands.staging.item.add.read_binary_file", return_value=b"{}")
  response = cli.invoke(pdp, ["staging", "item", "add", "--bucket", "fake-bucket", "--item-id", "fake-id", "--file",
                              "fake-path", "-j"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_add_item_json.snapshot')


def test_add_item_from_file_and_interactive(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.item.command.add`,
  when the --file and --interactive flags were provided.
  """
  mocker.patch("commands.staging.item.add.put", return_value=b'{'
                                                             b'"transactionId": "645e9901c8e408abf7e1a194",'
                                                             b'"timestamp": "2023-05-12T19:52:33.245Z",'
                                                             b'"action": "ADD",'
                                                             b'"bucket": "test",'
                                                             b'"contentId": "test3",'
                                                             b'"checksum": "23b5c58d597754037351ebdc5497882b"'
                                                             b'}')
  mocker.patch("commands.staging.item.add.read_binary_file", return_value=b'{\n\n}')
  mock_edit = mocker.patch("commands.staging.item.add.click.edit", return_value='{\n"fake":"property"\n}')
  response = cli.invoke(pdp, ["staging", "item", "add", "--bucket", "fake-bucket", "--item-id", "fake-id", "--file",
                              "fake-path", "--interactive"])
  assert response.exit_code == 0
  mock_edit.assert_called_once_with("{\n\n}")
  snapshot.assert_match(response.output, 'test_add_item_from_file_and_interactive.snapshot')


def test_add_item_from_file_and_interactive_no_content(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.item.command.add`,
  when the --file and --interactive flags were provided.
  """
  mocker.patch("commands.staging.item.add.put", return_value=b'{'
                                                             b'"transactionId": "645e9901c8e408abf7e1a194",'
                                                             b'"timestamp": "2023-05-12T19:52:33.245Z",'
                                                             b'"action": "ADD",'
                                                             b'"bucket": "test",'
                                                             b'"contentId": "test3",'
                                                             b'"checksum": "23b5c58d597754037351ebdc5497882b"'
                                                             b'}')
  mocker.patch("commands.staging.item.add.read_binary_file", return_value=b'{\n\n}')
  mock_edit = mocker.patch("commands.staging.item.add.click.edit", return_value=None)
  response = cli.invoke(pdp, ["staging", "item", "add", "--bucket", "fake-bucket", "--item-id", "fake-id", "--file",
                              "fake-path", "--interactive"])
  assert response.exit_code == 1
  mock_edit.assert_called_once_with("{\n\n}")
  snapshot.assert_match(response.output, 'test_add_item_from_file_and_interactive.snapshot')


def test_add_item_interactive(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.item.command.add`,
  when the --interactive flags were provided.
  """
  mocker.patch("commands.staging.item.add.put", return_value=b'{'
                                                             b'"transactionId": "645e9901c8e408abf7e1a194",'
                                                             b'"timestamp": "2023-05-12T19:52:33.245Z",'
                                                             b'"action": "ADD",'
                                                             b'"bucket": "test",'
                                                             b'"contentId": "test3",'
                                                             b'"checksum": "23b5c58d597754037351ebdc5497882b"'
                                                             b'}')
  mock_edit = mocker.patch("commands.staging.item.add.click.edit", return_value='{\n"fake":"property"\n}')
  response = cli.invoke(pdp,
                        ["staging", "item", "add", "--bucket", "fake-bucket", "--item-id", "fake-id", "--interactive"])
  assert response.exit_code == 0
  mock_edit.assert_called_once_with("{\n\n}")
  snapshot.assert_match(response.output, 'test_add_item_interactive.snapshot')


def test_add_item_no_id(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.item.command.add`,
  when the --item-id was not provided.
  """
  mocker.patch("commands.staging.item.add.put", return_value=b'{'
                                                             b'"transactionId": "645e9901c8e408abf7e1a194",'
                                                             b'"timestamp": "2023-05-12T19:52:33.245Z",'
                                                             b'"action": "ADD",'
                                                             b'"bucket": "test",'
                                                             b'"contentId": "test3",'
                                                             b'"checksum": "23b5c58d597754037351ebdc5497882b"'
                                                             b'}')
  mock_edit = mocker.patch("commands.staging.item.add.click.edit", return_value='{\n"fake":"property"\n}')
  mocker.patch("commands.staging.item.command.uuid.uuid4", return_value="autogenerate-fake")
  response = cli.invoke(pdp,
                        ["staging", "item", "add", "--bucket", "fake-bucket", "--interactive"])
  assert response.exit_code == 0
  mock_edit.assert_called_once_with("{\n\n}")
  snapshot.assert_match(response.output, 'test_add_item_interactive.snapshot')


def test_add_item_interactive_no_content(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.item.command.add`,
  when the --interactive flags were provided.
  """
  mocker.patch("commands.staging.item.add.put", return_value=b'{'
                                                             b'"transactionId": "645e9901c8e408abf7e1a194",'
                                                             b'"timestamp": "2023-05-12T19:52:33.245Z",'
                                                             b'"action": "ADD",'
                                                             b'"bucket": "test",'
                                                             b'"contentId": "test3",'
                                                             b'"checksum": "23b5c58d597754037351ebdc5497882b"'
                                                             b'}')
  mock_edit = mocker.patch("commands.staging.item.add.click.edit", return_value=None)
  response = cli.invoke(pdp,
                        ["staging", "item", "add", "--bucket", "fake-bucket", "--item-id", "fake-id", "--interactive"])
  assert response.exit_code == 1
  mock_edit.assert_called_once_with("{\n\n}")
  snapshot.assert_match(response.output, 'test_add_item_interactive_no_content.snapshot')


def test_add_item_no_content(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.item.command.add`,
  when the --interactive or --file flags were not provided.
  """
  mocker.patch("commands.staging.item.add.put", return_value=b'{'
                                                             b'"transactionId": "645e9901c8e408abf7e1a194",'
                                                             b'"timestamp": "2023-05-12T19:52:33.245Z",'
                                                             b'"action": "ADD",'
                                                             b'"bucket": "test",'
                                                             b'"contentId": "test3",'
                                                             b'"checksum": "23b5c58d597754037351ebdc5497882b"'
                                                             b'}')
  response = cli.invoke(pdp,
                        ["staging", "item", "add", "--bucket", "fake-bucket", "--item-id", "fake-id"])
  assert response.exit_code == 1
  snapshot.assert_match(response.output, 'test_add_item_interactive.snapshot')


def test_add_item_could_not_add(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.item.command.add`,
  when the --interactive or --file flags were not provided.
  """
  mocker.patch("commands.staging.item.add.read_binary_file", return_value=b'{\n\n}')
  mocker.patch("commands.staging.item.add.put", return_value=None)
  response = cli.invoke(pdp,
                        ["staging", "item", "add", "--bucket", "fake-bucket", "--item-id", "fake-id", "--file", "fake"])
  assert response.exit_code == 1
  snapshot.assert_match(response.output, 'test_add_item_interactive.snapshot')


def test_get_item(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.item.command.get`.
  """
  mocker.patch("commands.staging.item.get.get", return_value=b'{"fake":"content"}')
  response = cli.invoke(pdp,
                        ["staging", "item", "get", "--bucket", "fake-bucket", "--item-id", "fake-id", "--item-id",
                         "fake-id2", "--content-type", "metadata"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_get_item.snapshot')


def test_get_item_json(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.item.command.get`,
  when the --json flag was provided.
  """
  mocker.patch("commands.staging.item.get.get", return_value=b'{"fake":"content"}')
  response = cli.invoke(pdp,
                        ["staging", "item", "get", "--bucket", "fake-bucket", "--item-id", "fake-id", "--content-type",
                         "both", "-j"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_get_item_json.snapshot')


def test_get_item_failed(mocker, snapshot, mock_custom_exception):
  """
  Test the command defined in :func:`src.commands.staging.item.command.get`,
  when something went wrong.
  """
  mocker.patch("commands.staging.item.get.get", side_effect=lambda *args, **kwargs: mock_custom_exception(Exception))
  response = cli.invoke(pdp,
                        ["staging", "item", "get", "--bucket", "fake-bucket", "--item-id", "fake-id", "--content-type",
                         "both"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_get_item_failed.snapshot')


def test_delete_item_all(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.item.command.delete`.
  """
  mocker.patch("commands.staging.item.delete.create_spinner")
  mocker.patch("commands.staging.item.delete.get", return_value='{'
                                                                '"token":"fake-token",'
                                                                '"content": ['
                                                                '{"contentId":"fake-id1"},'
                                                                '{"contentId":"fake-id2"},'
                                                                '{"contentId":"fake-id3"}'
                                                                '] '
                                                                '}')
  mocker.patch("commands.staging.item.delete.delete", side_effect=[
    '{"transactionId": "fake-transaction1"}',
    None,
    '{"transactionId": "fake-transaction3"}'
  ])
  response = cli.invoke(pdp,
                        ["staging", "item", "delete", "--bucket", "fake-bucket", "--all"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_delete_item_all.snapshot')


def test_delete_items(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.item.command.delete`,
  when one or more --item-id flags were provided.
  """
  mocker.patch("commands.staging.item.delete.create_spinner")
  mock_get = mocker.patch("commands.staging.item.delete.get")
  mocker.patch("commands.staging.item.delete.delete", side_effect=[
    '{"transactionId": "fake-transaction1"}',
    None
  ])
  response = cli.invoke(pdp,
                        ["staging", "item", "delete", "--bucket", "fake-bucket", "-i", "fake1", "-i", "fake2"])
  assert mock_get.call_count == 0
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_delete_items.snapshot')


def test_delete_item_not_allowed(snapshot):
  """
  Test the command defined in :func:`src.commands.staging.item.command.delete`,
  when no --item-id and --all flags were provided.
  """
  response = cli.invoke(pdp,
                        ["staging", "item", "delete", "--bucket", "fake-bucket"])
  assert response.exit_code == 1
  snapshot.assert_match(response.exception.message, 'test_delete_item_not_allowed.snapshot')


def test_delete_item_filter(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.item.command.delete`,
  when the user wants to delete by filter.
  """
  mocker.patch("commands.staging.item.delete.click.edit", return_value="{}")
  mocker.patch(
    "commands.staging.item.delete.delete",
    return_value='['
                 '{"transactionId": "fake-transaction1", "contentId": "fake-content1"},'
                 '{"transactionId": "fake-transaction2", "contentId": "fake-content2"}'
                 ']'
  )
  response = cli.invoke(pdp,
                        ["staging", "item", "delete", "--bucket", "fake-bucket", "--filter"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_delete_item_filter.snapshot')


def test_delete_item_filter_no_deleted(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.item.command.delete`,
  when the filter doesn't match with any item.
  """
  mocker.patch("commands.staging.item.delete.click.edit", return_value="{}")
  mocker.patch(
    "commands.staging.item.delete.delete",
    return_value='[]'
  )
  response = cli.invoke(pdp,
                        ["staging", "item", "delete", "--bucket", "fake-bucket", "--filter"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_delete_item_filter_no_deleted.snapshot')


def test_delete_item_filter_failed(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.item.command.delete`,
  when the call to delete filter fails.
  """
  mocker.patch("commands.staging.item.delete.click.edit", return_value="{}")
  mocker.patch(
    "commands.staging.item.delete.delete",
    return_value=None
  )
  response = cli.invoke(pdp,
                        ["staging", "item", "delete", "--bucket", "fake-bucket", "--filter"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_delete_item_filter_failed.snapshot')


def test_delete_item_filter_no_criteria(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.staging.item.command.delete`,
  when the user don't provide the filter criteria.
  """
  mocker.patch("commands.staging.item.delete.click.edit", return_value=None)
  response = cli.invoke(pdp,
                        ["staging", "item", "delete", "--bucket", "fake-bucket", "--filter"])
  assert response.exit_code == 1
  snapshot.assert_match(response.exception.message, 'test_delete_item_filter_no_criteria.snapshot')
