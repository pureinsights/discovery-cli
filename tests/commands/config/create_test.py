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
from commands.config.create import deployment_stage, input_stage, interactive_input, parsing_stage, run as run_create, \
  writing_stage
from commons.constants import PIPELINE


def test_interactive_input_result_none(mocker):
  """
  Test the function defined in :func:`src.commands.config.create.interactive_input_result`,
  when the result of the user is None.
  """
  mocker.patch("commands.config.create.click.edit", return_value=None)
  json_to_entities_mock = mocker.patch("commands.config.create.json_to_pdp_entities")
  placeholder = "fake placeholder"
  interactive_input(placeholder)
  json_to_entities_mock.assert_called_once_with(placeholder)


def test_input_stage_not_file_not_interactive():
  """
  Test the function defined in :func:`src.commands.config.create.input_stage`,
  when the result of the user is None.
  """
  assert input_stage(None, False) == ('"given by user"', [])


def test_input_stage_file_and_interactive(mocker):
  """
  Test the function defined in :func:`src.commands.config.create.input_stage`,
  when the result of the user is None.
  """
  mocker.patch("commands.config.create.read_entities", return_value="[]")
  mocker.patch("commands.config.create.interactive_input", return_value=['fake'])
  assert input_stage("None", True) == ('None', ['fake'])


def test_parsing_stage_no_pdp_project_structure(mocker):
  """
  Test the function defined in :func:`src.commands.config.create.parsing_stage`,
  when the project doesn't have a pdp project structure.
  """
  mocker.patch("commands.config.create.has_pdp_project_structure", return_value=False)
  assert parsing_stage('fake path', PIPELINE, [{'fake': True}], 'fake path') == [{'fake': True}]


def test_parsing_stage_no_name_and_id(mocker):
  """
  Test the function defined in :func:`src.commands.config.create.parsing_stage`,
  when the entity doesn't have id and name.
  """
  mocker.patch("commands.config.create.has_pdp_project_structure", return_value=True)
  mocker.patch("commands.config.create.raise_for_pdp_data_inconsistencies")
  mocker.patch("commands.config.create.get_all_entities_names_ids", return_value={})
  assert parsing_stage('fake path', PIPELINE, [{'fake': True}, {'fake2': False}], 'fake path') == [{'fake': True},
                                                                                                   {'fake2': False}]


def test_writing_stage_no_pdp_project_structure(mocker):
  """
  Test the function defined in :func:`src.commands.config.create.writing_stage`,
  when the project doesn't have a pdp project structure.
  """
  mocker.patch("commands.config.create.has_pdp_project_structure", return_value=False)
  assert writing_stage('fake path', PIPELINE, [{'fake': 'fake-value'}]) == []


def test_writing_stage_no_entities_read(mocker):
  """
  Test the function defined in :func:`src.commands.config.create.writing_stage`,
  when no entities were read.
  """
  mocker.patch("commands.config.create.replace_ids")
  mocker.patch("commands.config.create.write_entities")
  mocker.patch("commands.config.create.read_entities", return_value=[])
  mocker.patch("commands.config.create.has_pdp_project_structure", return_value=True)
  assert writing_stage('fake path', PIPELINE, [{'fake': 'fake-value', 'id': 2}]) == [{'fake': 'fake-value', 'id': 2}]


def test_deployment_stage_updating_entity(mocker):
  """
  Test the function defined in :func:`src.commands.config.create.deployment_stage`,
  when the entities contain an id and needs to be updated.
  """
  mocker.patch("commands.config.create.create_spinner")
  mocker.patch("commands.config.create.handle_and_continue", return_value=(False, None))
  deployed_entities = deployment_stage({}, PIPELINE, [{'fake': 'fake-value', 'id': 2}], False, True)
  assert deployed_entities == []


def test_deployment_stage_updating_entity_json_flag(mocker):
  """
  Test the function defined in :func:`src.commands.config.create.deployment_stage`,
  when the entities contain an id and needs to be updated.
  """
  mocker.patch("commands.config.create.create_spinner")
  mocker.patch("commands.config.create.handle_and_continue", return_value=(False, None))
  deployed_entities = deployment_stage({}, PIPELINE, [{'fake': 'fake-value', 'id': 2}], False, False)
  assert deployed_entities == []


def test_deployment_stage_creating_entity_json_flag(mocker):
  """
  Test the function defined in :func:`src.commands.config.create.deployment_stage`,
  when the json flag is True.
  """
  mocker.patch("commands.config.create.create_spinner")
  mocker.patch("commands.config.create.handle_and_continue", return_value=(True, 'fakeid'))
  deployed_entities = deployment_stage({}, PIPELINE, [{'fake': 'fake-value', 'id': 2}], False, False)
  assert deployed_entities == [{'fake': 'fake-value', 'id': 'fakeid'}]


def test_run_no_pdp_project_structure(mocker):
  """
  Test the function defined in :func:`src.commands.config.create.run`,
  when the path passed don't have a valid pdp project structure.
  """
  mocker.patch("commands.config.create.are_same_pdp_entity", return_value=False)
  mocker.patch("commands.config.create.has_pdp_project_structure", return_value=False)
  mocker.patch("commands.config.create.input_stage",
               return_value=('fake-path', [{'fake': 'fake-value', 'id': 'fakeid'}]))
  mocker.patch("commands.config.create.deployment_stage",
               return_value=[{'fake': 'fake-value1', 'id': 'fakeid'}])
  parsing_stage_mock = mocker.patch("commands.config.create.parsing_stage")
  writing_stage_mock = mocker.patch("commands.config.create.writing_stage")
  printing_stage_mock = mocker.patch("commands.config.create.printing_stage")
  run_create({}, 'fake-path', PIPELINE, 'fake-path', True, False, False, False, False)
  assert parsing_stage_mock.call_count == 0
  writing_stage_mock.assert_called_once_with('fake-path', PIPELINE, [])
  printing_stage_mock.assert_called_once()
