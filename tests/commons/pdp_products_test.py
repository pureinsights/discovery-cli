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
from unittest.mock import mock_open

import pytest

from commons.constants import CORE, CREDENTIAL, DISCOVERY, DISCOVERY_PROCESSOR_ENTITY, ENDPOINT, INGESTION, \
  INGESTION_PROCESSOR, \
  INGESTION_PROCESSOR_ENTITY, PIPELINE, \
  SCHEDULER, SEED, URL_CREATE, URL_GET_BY_ID, URL_UPDATE
from commons.custom_classes import DataInconsistency, PdpException
from commons.pdp_products import are_same_pdp_entity, associate_values_from_entities, clear_from_name_format, \
  create_or_update_entity, \
  delete_entity_from_pdp_project, delete_pdp_entity, delete_references_from_entity, export_entities, \
  get_all_entities_names_ids, get_entity_by_id_from_local_entities, get_entity_type_by_name, identify_entity, \
  json_to_pdp_entities, order_products_to_deploy, \
  replace_ids, \
  replace_ids_for_names, \
  replace_names_by_ids, replace_value, search_value_from_entity


def test_identify_entity_is_not_dict():
  """
  Test the function defined in :func:`commons.pdp_products.identify_entity`,
  when the entity is not a dict.
  """
  entity = 'fake-entity'
  response = identify_entity(entity)
  assert response == entity


@pytest.mark.parametrize('entity, property', [
  ({'type': 'fake-type'}, 'type'),
  ({'description': 'fake-description'}, 'description'),
  ({'id': 'fake-id'}, 'id'),
  ({'name': 'fake-name'}, 'name')
])
def test_identify_entity_return_successfully(entity, property):
  """
  Test the function defined in :func:`commons.pdp_products.identify_entity`.
  """
  response = identify_entity(entity)
  assert response == f'{property} "{entity[property]}"'


def test_identify_entity_return_default():
  """
  Test the function defined in :func:`commons.pdp_products.identify_entity`,
  when the entity does not have the property.
  """
  entity = {}
  response = identify_entity(entity, default='fake-default')
  assert response == 'fake-default'


def test_associate_values_from_entities_raises_DataInconsistency(mocker):
  """
  Test the function defined in :func:`commons.pdp_products.associate_values_from_entities`,
  when the entities do not have the properties.
  """
  entity1 = {'property': 'property'}
  entity2 = {}
  mocker.patch('commons.pdp_products.print_exception')
  mock_exception = mocker.patch('commons.pdp_products.DataInconsistency')
  associate_values_from_entities(entity1, 'property', entity2, 'property')
  mock_exception.assert_called_with(message=f'Entity with {entity2} does not have field property.',
                                    severity='warning', handled=True)
  associate_values_from_entities(entity2, 'property', entity1, 'property')
  mock_exception.assert_called_with(message=f'Entity with {entity2} does not have field property.',
                                    severity='warning', handled=True)


@pytest.mark.parametrize('entity1, entity2', [
  ({'property': 1}, {'property': 2}),
  ({'property': 2}, {'property': 1}),
  ({'property': 1}, {'property': 1}),
  ({'property': 2}, {'property': 2})
])
def test_associate_values_from_entities_successful(entity1, entity2):
  """
  Test the function defined in :func:`commons.pdp_products.associate_values_from_entities`.
  """
  response = associate_values_from_entities(entity1, 'property', entity2, 'property')
  assert (entity1['property'], entity2['property']) == response


@pytest.mark.parametrize('dir_name, entity_str, entities, call_count, ids', [
  ('./Ingestion', '[{"id":"fake-id","name":"fake-name"}]', [{}], 4, {}),
  ('./Ingestion', '[{"id":"fake-id","name":"fake-name"}]', {}, 4, None),
  ('Ingestion', '[{"id":"fake-id","name":"fake-name"}]', [{}], 4, {}),
  ('fake-path', '[{"id":"fake-id","name":"fake-name"}]', [{}], 0, {}),
  ('./Discovery', '[{"id":"fake-id","name":"fake-name"}]', [{}], 2, None),
  ('Core', '[{"id":"fake-id","name":"fake-name"}]', {}, 1, None),
])
def test_replace_ids(mocker, dir_name, entity_str, entities, call_count, ids):
  """
  Test the function defined in :func:`commons.pdp_products.replace_ids`.
  """
  mocker.patch('builtins.open', mock_open(read_data=entity_str))
  mocker.patch('commons.pdp_products.json.load', return_value=entities)
  mock_dump = mocker.patch('commons.pdp_products.json.dump')
  mock_replace = mocker.patch('commons.pdp_products.replace_ids_for_names')
  replace_ids(dir_name, ids)
  assert mock_dump.call_count == call_count
  assert mock_replace.call_count == call_count


def test_replace_ids_failed_replace_id(mocker):
  """
  Test the function defined in :func:`commons.pdp_products.replace_ids`.
  """
  mocker.patch('builtins.open', mock_open(read_data=''))
  mocker.patch('commons.pdp_products.json.load', return_value={})
  mocker.patch('commons.pdp_products.handle_and_continue', return_value=(False, {}))
  mock_dump = mocker.patch('commons.pdp_products.json.dump')
  mock_replace = mocker.patch('commons.pdp_products.replace_ids_for_names')
  replace_ids('./Ingestion', {})
  assert mock_dump.call_count == 4
  assert mock_replace.call_count == 0


@pytest.mark.parametrize('entity_type, entities, expected_ids', [
  (INGESTION_PROCESSOR_ENTITY, [{'id': 'fake-id4', 'name': 'fake-name4'}], {'fake-id4': 'fake-name4'}),
  (PIPELINE, [{
    'id': 'fake-id4',
    'name': 'fake-name4',
    'steps': [{'processorId': 'fake-id1'}, {'processorId': 'fake-id2'}, {'processorId': 'fake-id4'}]
  }], {'fake-id4': 'fake-name4'}),
  (SEED, [{}], {}),
  (SCHEDULER, [{'id': 'fake-id4', 'name': 'fake-name4'}], {'fake-id4': 'fake-name4'}),
  (CREDENTIAL, [{'id': 'fake-id4', 'name': 'fake-name4'}], {'fake-id4': 'fake-name4'}),
  (ENDPOINT, [{'id': 'fake-id4', 'name': 'fake-name4'}], {'fake-id4': 'fake-name4'}),
  (DISCOVERY_PROCESSOR_ENTITY, [{'id': 'fake-id4', 'name': 'fake-name4'}], {'fake-id4': 'fake-name4'})
])
def test_replace_ids_for_names(entity_type, entities, expected_ids):
  """
  Test the function defined in :func:`commons.pdp_products.replace_ids_for_names`.
  """
  ids = {
    'fake-id1': 'fake-name1',
    'fake-id2': 'fake-name2',
    'fake-id3': 'fake-name3'
  }
  expected_ids = {**expected_ids, **ids}
  response = replace_ids_for_names(entity_type, entities, ids)
  assert response == expected_ids


def test_replace_ids_for_names_failed_associate_values(mocker):
  """
  Test the function defined in :func:`commons.pdp_products.replace_ids_for_names`.
  """
  mocker.patch('commons.pdp_products.handle_and_continue', return_value=(False, None))
  ids = {
    'fake-id1': 'fake-name1',
    'fake-id2': 'fake-name2',
    'fake-id3': 'fake-name3'
  }
  expected_ids = ids
  response = replace_ids_for_names(INGESTION_PROCESSOR, [{'id': 'fake-id4', 'name': 'fake-name4'}], ids)
  assert response == expected_ids


@pytest.mark.parametrize('entity_type, entity, expected_entity', [
  (PIPELINE, {
    'id': 'fake-id7',
    'name': 'fake-name7',
    'steps': {'processorId': 'fake-id1'}
  }, {
     'id': 'fake-id7',
     'name': 'fake-name7',
     'steps': {'processorId': '{{ fromName(\'fake-name1\') }}'}
   }),
  (PIPELINE, {
    'id': 'fake-id7',
    'name': 'fake-name7',
    'steps': {'processorId': None}
  }, {
     'id': 'fake-id7',
     'name': 'fake-name7',
     'steps': {'processorId': None}
   }),
  (PIPELINE, {
    'id': 'fake-id7',
    'name': 'fake-name7',
    'fake-child': {'fake-child': {'processorId': 'fake-id1'}}
  }, {
     'id': 'fake-id7',
     'name': 'fake-name7',
     'fake-child': {'fake-child': {'processorId': '{{ fromName(\'fake-name1\') }}'}}
   }),
  (PIPELINE, {
    'id': 'fake-id7',
    'name': 'fake-name7',
    'fake-child': {'fake-children': [{'processorId': 'fake-id1'}]}
  }, {
     'id': 'fake-id7',
     'name': 'fake-name7',
     'fake-child': {'fake-children': [{'processorId': '{{ fromName(\'fake-name1\') }}'}]}
   }),
  (PIPELINE, {
    'id': 'fake-id7',
    'name': 'fake-name7',
    'fake-child': {'fake-children': [{'processorId': 'fake-id12'}]},
    'processors': ['fake-id3']
  }, {
     'id': 'fake-id7',
     'name': 'fake-name7',
     'fake-child': {'fake-children': [{'processorId': 'fake-id12'}]},
     'processors': ['{{ fromName(\'fake-name3\') }}']
   }),
  (PIPELINE, 'fake-entity', 'fake-entity')
])
def test_replace_value_successful(entity_type, entity, expected_entity):
  """
  Test the function defined in :func:`commons.pdp_products.replace_value`.
  """
  values = {
    'fake-id1': 'fake-name1',
    'fake-id2': 'fake-name2',
    'fake-id3': 'fake-name3',
    'fake-id4': 'fake-name4',
    'fake-id5': 'fake-name5',
    'fake-id6': 'fake-name6',
    'fake-id12': None
  }
  replace_value(entity_type, entity, values)
  assert entity == expected_entity


def test_replace_value_successful_with_custom_to_field():
  """
  Test the function defined in :func:`commons.pdp_products.replace_value`.
  """
  values = {
    'fake-id1': 'fake-name1',
    'fake-id2': 'fake-name2',
    'fake-id3': 'fake-name3',
    'fake-id4': 'fake-name4',
    'fake-id5': 'fake-name5',
    'fake-id6': 'fake-name6',
    'fake-id12': None
  }
  entity = {
    'id': 'fake-id7',
    'name': 'fake-name7',
    'fake-child': {'fake-children': [{'processorId': 'fake-id6'}]},
    'processors': ['fake-id3']
  }
  expected_entity = {
    'id': 'fake-id7',
    'name': 'fake-name7',
    'fake-child': {'fake-children': [{'processorId': "{{ fromName('fake-name6') }}"}]},
    'processors': ['fake-id3']
  }
  replace_value(SEED, entity, values, to_fields='processorId')
  assert entity == expected_entity


@pytest.mark.parametrize('entity_type, entity, expected_message', [
  (PIPELINE, {
    'id': 'fake-id7',
    'name': 'fake-name7',
    'steps': {'processorId': 'fake-id8'}
  }, 'Value "fake-id8" does not exist while attempting to replace in field '
     '"processorId". Entity has no name and no id in file pipelines.json. That '
     'could means that the name "fake-id8" do not exists or the entity "fake-id8" '
     'do not have an Id.'
   ),
  (SEED, {
    'pipelineId': 'fake-id8'
  }, 'Value "fake-id8" does not exist while attempting to replace in field '
     '"pipelineId". Entity has no name and no id in file seeds.json. That could '
     'means that the name "fake-id8" do not exists or the entity "fake-id8" do not '
     'have an Id.'
   ),
  (SCHEDULER, {
    'id': 'fake-id7',
    'name': 'fake-name7',
    'seedId': 'fake-id8'
  }, 'Value "fake-id8" does not exist while attempting to replace in field '
     '"seedId". Entity name "fake-name7" in file cron_jobs.json. That could means '
     'that the name "fake-id8" do not exists or the entity "fake-id8" do not have '
     'an Id.'
   )
])
def test_replace_value_failed(mocker, entity_type, entity, expected_message):
  """
  Test the function defined in :func:`commons.pdp_products.replace_value`,
  when the value to replace is not present on the values.
  """
  mock_warning = mocker.patch('commons.pdp_products.print_warning')
  values = {
    'fake-id1': 'fake-name1',
    'fake-id2': 'fake-name2',
    'fake-id3': 'fake-name3',
    'fake-id4': 'fake-name4',
    'fake-id5': 'fake-name5',
    'fake-id6': 'fake-name6'
  }
  replace_value(entity_type, entity, values, to_fields=['processorId', 'pipelineId', 'seedId'])
  mock_warning.assert_called_with(expected_message)


def test_export_entities_successful_without_extract(mocker):
  """
  Test the function defined in :func:`commons.pdp_products.export_entities`,
  without extracting the files.
  """
  content = b'fake-content'
  mocker.patch('commons.pdp_products.get', return_value=content)
  m = mock_open()
  mocker.patch('commons.pdp_products.open', m)
  api = 'http://fake-url'
  path = 'fake-path'
  ids = {'id': 'name'}
  response = export_entities(api, path, False, ids=ids)
  m.assert_called_once_with(os.path.join(path, 'export.zip'), 'wb')
  m().write.assert_called_once_with(content)
  assert ids == response


def test_export_entity_by_id_successful_without_extract(mocker):
  """
  Test the function defined in :func:`commons.pdp_products.export_entities`,
  without extracting the files and just exporting one entity.
  """
  content = b'fake-content'
  get_mock = mocker.patch('commons.pdp_products.get', return_value=content)
  m = mock_open()
  mocker.patch('commons.pdp_products.open', m)
  api = 'http://fake-url'
  path = 'fake-path'
  ids = {'id': 'name'}
  entity_type = SEED
  entity_id = 'fakeid'
  response = export_entities(api, path, False, ids=ids, entity_type=entity_type, entity_id=entity_id)
  m.assert_called_once_with(os.path.join(path, 'export.zip'), 'wb')
  m().write.assert_called_once_with(content)
  assert ids == response
  get_mock.assert_called_once_with(
    f'{api}/export/{entity_type.type}',
    params={f'{entity_type.type}Id': entity_id, 'includeDependencies': False}
  )


def test_export_entities_successful_without_extract_verbose(mocker):
  """
  Test the function defined in :func:`commons.pdp_products.export_entities`,
  without extracting the files.
  """
  content = b'fake-content'
  mocker.patch('commons.pdp_products.get', return_value=content)
  m = mock_open()
  mocker.patch('commons.pdp_products.open', m)
  mocker.patch('commons.pdp_products.create_spinner')
  mock_change_text = mocker.patch('commons.pdp_products.spinner_change_text')
  mock_spinner_ok = mocker.patch('commons.pdp_products.spinner_ok')
  api = 'http://fake-url'
  path = 'fake-path'
  ids = {'id': 'name'}
  response = export_entities(api, path, False, ids=ids, verbose=True)
  m.assert_called_once_with(os.path.join(path, 'export.zip'), 'wb')
  m().write.assert_called_once_with(content)
  assert mock_change_text.call_count == 2
  assert mock_spinner_ok.call_count == 1
  assert ids == response


def test_export_entities_successful_with_extraction(mocker, mock_path_exists):
  """
  Test the function defined in :func:`commons.pdp_products.export_entities`,
  extracting the files.
  """
  mock_path_exists(True)
  content = b'fake-content'
  mocker.patch('commons.pdp_products.get', return_value=content)
  m = mock_open()
  mocker.patch('commons.pdp_products.open', m)
  mocker.patch('commons.pdp_products.zipfile')
  mocker.patch('commons.pdp_products.os.remove')
  mocker.patch('commons.pdp_products.handle_and_continue', return_value=(True, {}))
  api = 'http://fake-url'
  path = 'fake-path'
  ids = {'id': 'name'}
  response = export_entities(api, path, True, ids=ids)
  m.assert_called_once_with(os.path.join(path, 'export.zip'), 'wb')
  m().write.assert_called_once_with(content)
  assert ids == response


def test_export_entities_failed_with_extraction(mocker, mock_path_exists):
  """
  Test the function defined in :func:`commons.pdp_products.export_entities`,
  when fail while trying to extract the files.
  """
  mock_path_exists(False)
  content = b'fake-content'
  mocker.patch('commons.pdp_products.get', return_value=content)
  m = mock_open()
  mocker.patch('commons.pdp_products.open', m)
  mocker.patch('commons.pdp_products.zipfile')
  mocker.patch('commons.pdp_products.os.remove')
  mocker.patch('commons.pdp_products.handle_and_continue', return_value=(False, None))
  api = 'http://fake-url'
  path = 'fake-path'
  ids = {'id': 'name'}
  response = export_entities(api, path, True, ids=ids)
  m.assert_called_once_with(os.path.join(path, 'export.zip'), 'wb')
  m().write.assert_called_once_with(content)
  assert ids == response


def test_export_entities_successful_with_extraction_verbose(mocker, mock_path_exists):
  """
  Test the function defined in :func:`commons.pdp_products.export_entities`,
  extracting the files, with verbose flag.
  """
  mock_path_exists(False)
  content = b'fake-content'
  mocker.patch('commons.pdp_products.get', return_value=content)
  m = mock_open()
  mocker.patch('commons.pdp_products.open', m)
  mocker.patch('commons.pdp_products.zipfile')
  mocker.patch('commons.pdp_products.create_spinner')
  mock_change_text = mocker.patch('commons.pdp_products.spinner_change_text')
  mock_spinner_ok = mocker.patch('commons.pdp_products.spinner_ok')
  mocker.patch('commons.pdp_products.os.remove')
  api = 'http://fake-url'
  path = 'fake-path'
  ids = {'id': 'name'}
  response = export_entities(api, path, True, ids=ids, verbose=True)
  m.assert_called_once_with(os.path.join(path, 'export.zip'), 'wb')
  m().write.assert_called_once_with(content)
  assert mock_change_text.call_count == 3
  assert mock_spinner_ok.call_count == 1
  assert ids == response


@pytest.mark.parametrize('entity_type,', [(SEED)])
def test_replace_names_by_ids(entity_type):
  """
  Test the function defined in :func:`commons.pdp_products.replace_names_by_ids`.
  """
  entities = [
    {'id': 'id01', 'name': 'name01'},
    {'id': 'id02', 'name': 'name02', 'pipelineId': "{{ fromName('name01') }}"},
    {'name': 'name03', 'pipelineId': "{{fromName('name02')}}"},
    {'id': None, 'pipelineId': "{{fromName('fake-name')}}"}
  ]
  expected_entities = [
    {'id': 'id01', 'name': 'name01'},
    {'id': 'id02', 'name': 'name02', 'pipelineId': "id01"},
    {'name': 'name03', 'pipelineId': "id02"},
    {'id': None, 'pipelineId': "{{fromName('fake-name')}}"}
  ]
  names = replace_names_by_ids(entity_type, entities, {})
  assert names == {'name01': 'id01', 'id01': 'id01', 'name02': 'id02', 'id02': 'id02'}
  assert entities == expected_entities


def test_clear_from_name_format():
  """
  Test the function defined in :func:`commons.pdp_products.clear_from_name_format`.
  """
  fake_text = "{{ fromName('fake_text') }}"
  result = clear_from_name_format(fake_text)
  assert result == 'fake_text'


def test_clear_from_name_format_value_not_str():
  """
  Test the function defined in :func:`commons.pdp_products.clear_from_name_format`,
  when the value is not a str.
  """
  result = clear_from_name_format(1)
  assert result == 1


def test_order_products_to_deploy():
  """
  Test the function defined in :func:`commons.pdp_products.order_products_to_deploy`.
  """
  # The only order that matters is that core is before ingestion
  products = [DISCOVERY, INGESTION, CORE]
  assert order_products_to_deploy(products) == [DISCOVERY, CORE, INGESTION]


def test_order_products_to_deploy_without_core():
  """
  Test the function defined in :func:`commons.pdp_products.order_products_to_deploy`,
  when core is not one of the products.
  """
  # The only order that matters is that core is before ingestion
  products = [INGESTION, DISCOVERY]
  assert order_products_to_deploy(products) == [INGESTION, DISCOVERY]


def test_order_products_to_deploy_without_products():
  """
  Test the function defined in :func:`commons.pdp_products.order_products_to_deploy`,
  when no products were provided.
  """
  # The only order that matters is that core is before ingestion
  products = None
  assert order_products_to_deploy(products) == []


def test_create_or_update_entity(mocker):
  """
  Test the function defined in :func:`commons.pdp_products.create_or_update_entity`,
  when it is creating a new entity.
  """
  mocker.patch("commons.pdp_products.verbose")
  fakeid = 'fakeid01'
  post_mock = mocker.patch("commons.pdp_products.post", return_value=f'{{"id": "{fakeid}"}}')
  fake_url = 'http://fake-url'
  fake_type = 'fake-type'
  fake_entity = {'noid': fakeid}
  id = create_or_update_entity(fake_url, fake_type, fake_entity)
  post_mock.assert_called_once_with(URL_CREATE.format(fake_url, entity=fake_type), json=fake_entity)
  assert id == fakeid


def test_create_or_update_entity_creating_entity_with_new_id(mocker):
  """
  Test the function defined in :func:`commons.pdp_products.create_or_update_entity`,
  when it is creating a new entity but with a given id.
  """
  mocker.patch("commons.pdp_products.verbose")
  fakeid = 'fakeid01'
  get_mock = mocker.patch("commons.pdp_products.get", return_value=None)
  post_mock = mocker.patch("commons.pdp_products.post", return_value=f'{{"id": "{fakeid}"}}')
  fake_url = 'http://fake-url'
  fake_type = 'fake-type'
  fake_entity = {'id': fakeid}
  id = create_or_update_entity(fake_url, fake_type, fake_entity)
  get_mock.assert_called_once_with(URL_GET_BY_ID.format(fake_url, entity=fake_type, id=fakeid))
  post_mock.assert_called_once_with(URL_CREATE.format(fake_url, entity=fake_type), json=fake_entity)
  assert id == fakeid


def test_create_or_update_entity_updating_entity(mocker):
  """
  Test the function defined in :func:`commons.pdp_products.create_or_update_entity`,
  when it is updating an entity.
  """
  mocker.patch("commons.pdp_products.verbose")
  fakeid = 'fakeid01'
  get_mock = mocker.patch("commons.pdp_products.get", return_value=fakeid)
  put_mock = mocker.patch("commons.pdp_products.put", return_value=f'{{"id": "{fakeid}"}}')
  fake_url = 'http://fake-url'
  fake_type = 'fake-type'
  fake_entity = {'id': fakeid}
  id = create_or_update_entity(fake_url, fake_type, fake_entity)
  get_mock.assert_called_once_with(URL_GET_BY_ID.format(fake_url, entity=fake_type, id=fakeid))
  put_mock.assert_called_once_with(URL_UPDATE.format(fake_url, entity=fake_type, id=fakeid), json=fake_entity)
  assert id == fakeid


# TODO: Delete this if is not necessary
def test_create_or_update_entity_updating_entity_verbose(mocker):
  """
  Test the function defined in :func:`commons.pdp_products.create_or_update_entity`,
  when it is updating an entity.
  """
  mocker.patch("commons.pdp_products.verbose")
  fakeid = 'fakeid01'
  get_mock = mocker.patch("commons.pdp_products.get", return_value=fakeid)
  put_mock = mocker.patch("commons.pdp_products.put", return_value=f'{{"id": "{fakeid}"}}')
  fake_url = 'http://fake-url'
  fake_type = 'fake-type'
  fake_entity = {'id': fakeid}
  id = create_or_update_entity(fake_url, fake_type, fake_entity, verbose=True)
  get_mock.assert_called_once_with(URL_GET_BY_ID.format(fake_url, entity=fake_type, id=fakeid))
  put_mock.assert_called_once_with(URL_UPDATE.format(fake_url, entity=fake_type, id=fakeid), json=fake_entity)
  assert id == fakeid


def test_create_or_update_entity_error_occurred(mocker, snapshot):
  """
  Test the function defined in :func:`commons.pdp_products.create_or_update_entity`,
  when an error occurred.
  """
  mocker.patch("commons.console.create_spinner")
  mocker.patch("commons.pdp_products.verbose")
  fakeid = 'fakeid01'
  print_error_mock = mocker.patch("commons.pdp_products.print_error")
  get_mock = mocker.patch("commons.pdp_products.get", side_effect=PdpException)
  fake_url = 'http://fake-url'
  fake_type = 'fake-type'
  fake_entity = {'id': fakeid}
  id = create_or_update_entity(fake_url, fake_type, fake_entity, verbose=True)
  get_mock.assert_called_once_with(URL_GET_BY_ID.format(fake_url, entity=fake_type, id=fakeid))
  snapshot.assert_match(str(print_error_mock.call_args_list), 'test_create_or_update_entity_error_occurred.snapshot')


def test_get_entity_type_by_name_ingestion_processor():
  """
  Test the function defined in :func:`commons.pdp_products.get_entity_type_by_name`,
  when the entity name is ingestionProcessor.
  """
  entity_type = get_entity_type_by_name('ingestionProcessor')
  assert entity_type == INGESTION_PROCESSOR_ENTITY


def test_get_entity_type_by_name_discovery_processor():
  """
  Test the function defined in :func:`commons.pdp_products.get_entity_type_by_name`,
  when the entity name is discoveryProcessor.
  """
  entity_type = get_entity_type_by_name('discoveryProcessor')
  assert entity_type == DISCOVERY_PROCESSOR_ENTITY


def test_get_entity_type_by_name_unrecognized_entity_name():
  """
  Test the function defined in :func:`commons.pdp_products.get_entity_type_by_name`,
  when the entity name is not recognized.
  """
  entity_type = get_entity_type_by_name('fake_entity')
  assert entity_type is None


def test_get_all_entities_names_ids_bad_structure(mocker, test_project_path):
  """
  Test the function defined in :func:`commons.pdp_products.get_all_entities_names_ids`,
  when the given path does not have a pdp project structure.
  """
  mocker.patch("commons.pdp_products.has_pdp_project_structure", return_value=False)
  result = get_all_entities_names_ids(test_project_path(), [])
  assert result == {}


def test_get_all_entities_names_ids_can_not_read(mocker, test_project_path):
  """
  Test the function defined in :func:`commons.pdp_products.get_all_entities_names_ids`,
  when the function can not read the entities.
  """
  mocker.patch("commons.pdp_products.handle_and_continue", return_value=(False, None))
  result = get_all_entities_names_ids(test_project_path(), [])
  assert result == {}


def test_are_same_pdp_entity_not_same_attributes():
  """
  Test the function defined in :func:`commons.pdp_products.are_same_pdp_entity`,
  when the entities have different keys.
  """
  assert not are_same_pdp_entity({'a': 1}, {'a': 1, 'b': 1})


def test_are_same_pdp_entity_not_same_values():
  """
  Test the function defined in :func:`commons.pdp_products.are_same_pdp_entity`,
  when the entities have the same keys but have different values.
  """
  assert not are_same_pdp_entity({'a': 1, 'b': 2}, {'a': 1, 'b': 1})


def test_are_same_pdp_entity_same_values():
  """
  Test the function defined in :func:`commons.pdp_products.are_same_pdp_entity`,
  when the entities have the same keys and have the same values.
  """
  assert are_same_pdp_entity({'a': 1, 'b': 2, 'id': 2}, {'a': 1, 'b': 2})


def test_are_same_pdp_entity_same_ids():
  """
  Test the function defined in :func:`commons.pdp_products.are_same_pdp_entity`,
  when the entities have the same keys and have the same values.
  """
  assert are_same_pdp_entity({'b': 2, 'id': 2}, {'a': 1, 'b': 2, 'id': 2})


def test_are_same_pdp_entity_equal_entities():
  """
  Test the function defined in :func:`commons.pdp_products.are_same_pdp_entity`,
  when the entities are equal.
  """
  assert are_same_pdp_entity({'a': 1, 'b': 2, 'id': 2}, {'a': 1, 'b': 2, 'id': 2})


def test_json_to_pdp_entities():
  """
  Test the function defined in :func:`commons.pdp_products.json_to_pdp_entities`.
  """
  entities_json = '[ { "id":"fake"} ]'
  entities = json_to_pdp_entities(entities_json)
  assert entities == [{"id": "fake"}]


def test_json_to_pdp_entities_not_list():
  """
  Test the function defined in :func:`commons.pdp_products.json_to_pdp_entities`,
  when the JSON just contain one entity.
  """
  entities_json = '{ "id":"fake"}'
  entities = json_to_pdp_entities(entities_json)
  assert entities == [{"id": "fake"}]


def test_json_to_pdp_entities_bad_json():
  """
  Test the function defined in :func:`commons.pdp_products.json_to_pdp_entities`,
  when the JSON doesn't have a valid format.
  """
  entities_json = '{ "id":"fake"'
  with pytest.raises(PdpException) as error:
    entities = json_to_pdp_entities(entities_json)
  assert error.value.message == 'JSONDecodeError: Could not parse the JSON text. ' \
                                'Please check the file has a valid JSON format.'


def test_delete_pdp_entity_without_id():
  """
  Test the function defined in :func:`commons.pdp_products.delete_pdp_entity`,
  when the given entity doesn't have an id.
  """
  assert not delete_pdp_entity({}, SEED, {}, False, False)


def test_delete_pdp_entity_no_local(mocker):
  """
  Test the function defined in :func:`commons.pdp_products.delete_pdp_entity`,
  when the given entity doesn't have an id.
  """
  mock_delete_from_project = mocker.patch("commons.pdp_products.delete_entity_from_pdp_project")
  mocker.patch("commons.pdp_products.delete")
  mocker.patch("commons.pdp_products.json.loads", return_value={'acknowledged': True})
  assert delete_pdp_entity({'ingestion': ''}, SEED, {'id': 'fake1'}, False, False)
  assert mock_delete_from_project.call_count == 0


def test_delete_pdp_entity_local(mocker):
  """
  Test the function defined in :func:`commons.pdp_products.delete_pdp_entity`,
  when the given entity doesn't have an id.
  """
  mock_delete_from_project = mocker.patch("commons.pdp_products.delete_entity_from_pdp_project")
  mocker.patch("commons.pdp_products.delete")
  mocker.patch("commons.pdp_products.json.loads", return_value={'acknowledged': True})
  assert delete_pdp_entity({'ingestion': ''}, SEED, {'id': 'fake1'}, False, True)
  assert mock_delete_from_project.call_count == 1


def test_delete_entity_from_pdp_project_no_pdp_project_structure(mocker):
  """
  Test the function defined in :func:`commons.pdp_products.delete_entity_from_pdp_project`,
  when the project path doesn't have a pdp project structure.
  """
  write_mock = mocker.patch("commons.pdp_products.write_entities")
  mocker.patch("commons.pdp_products.has_pdp_project_structure", return_value=False)
  delete_entity_from_pdp_project({'project_path': ''}, SEED, {}, False)
  assert write_mock.call_count == 0


def test_delete_entity_from_pdp_project(mocker):
  """
  Test the function defined in :func:`commons.pdp_products.delete_entity_from_pdp_project`.
  """
  entity = {
    "seed": {
      "records": 1000,
      "recordSize": 8096
    },
    "name": "Cascade",
    "type": "random-generator-connector",
    "pipelineId": "{{ fromName('Cascade Pipeline') }}",
    "properties": {
      "index": "random_generated_docs"
    },
    "batchSize": "100",
    "id": "fakeidCascadeSeed"
  }
  mocker.patch("commons.pdp_products.has_pdp_project_structure", return_value=True)
  mocker.patch("commons.pdp_products.read_entities", return_value=[{'id': 'fake'}, entity])
  mocker.patch("commons.pdp_products.delete_references_from_entity")
  write_mock = mocker.patch("commons.pdp_products.write_entities")
  delete_entity_from_pdp_project({'project_path': 'fake-path'}, SEED, entity, False)
  file_path = os.path.join('fake-path', SEED.product.title(), SEED.associated_file_name)
  write_mock.assert_called_once_with(file_path, [{'id': 'fake'}])


def test_delete_entity_from_pdp_project_no_entities_read(mocker):
  """
  Test the function defined in :func:`commons.pdp_products.delete_pdp_entity`,
  when the pdp project doesn't have entities.
  """
  entity = {
    "seed": {
      "records": 1000,
      "recordSize": 8096
    },
    "name": "Cascade",
    "type": "random-generator-connector",
    "pipelineId": "{{ fromName('Cascade Pipeline') }}",
    "properties": {
      "index": "random_generated_docs"
    },
    "batchSize": "100",
    "id": "fakeidCascadeSeed"
  }
  mocker.patch("commons.pdp_products.has_pdp_project_structure", return_value=True)
  mocker.patch("commons.pdp_products.read_entities", return_value=[])
  mock_delete_references = mocker.patch("commons.pdp_products.delete_references_from_entity")
  write_mock = mocker.patch("commons.pdp_products.write_entities")
  delete_entity_from_pdp_project({'project_path': 'fake-path'}, SEED, entity, True)
  file_path = os.path.join('fake-path', SEED.product.title(), SEED.associated_file_name)
  write_mock.assert_called_once_with(file_path, [])
  assert mock_delete_references.call_count == 1


def test_delete_references_from_entity(mocker, test_project_path):
  """
  Test the function defined in :func:`commons.pdp_products.delete_references_from_entity`,
  when the pdp project doesn't have entities.
  """
  mocker.patch("commons.pdp_products.has_pdp_project_structure", return_value=True)
  mocker.patch("commands.config.get.get_entity_by_id", side_effect=[(None, None)])
  mock_delete = mocker.patch("commons.pdp_products.delete_entity_from_pdp_project")
  entity = {
    "seed": {
      "records": 1000,
      "recordSize": 8096
    },
    "name": "Cascade",
    "type": "random-generator-connector",
    "pipelineId": "fakeidCascadePipeline",
    "properties": {
      "index": "random_generated_docs"
    },
    "batchSize": "100",
    "id": "fakeidCascadeSeed"
  }
  config = {'project_path': test_project_path()}
  delete_references_from_entity(config, SEED, entity, True)
  reference_entity = {
    "name": "Cascade Pipeline",
    "active": True,
    "steps": [
      {
        "processorId": "fakeidIP1",
        "action": "hydrate"
      },
      {
        "processorId": "fakeidCascadeProcessor",
        "action": "hydrate"
      }
    ],
    "id": "fakeidCascadePipeline"
  }
  mock_delete.assert_called_once_with(config, PIPELINE, reference_entity, True)


def test_delete_references_from_entity_no_references_field(mocker, test_project_path, mock_custom_exception):
  """
  Test the function defined in :func:`commons.pdp_products.delete_references_from_entity`,
  when the entity doesn't have the reference entity field.
  """
  mocker.patch("commons.pdp_products.has_pdp_project_structure", return_value=True)
  mocker.patch("commons.pdp_products.search_value_from_entity",
               side_effect=lambda *args: mock_custom_exception(Exception))
  mock_delete = mocker.patch("commons.pdp_products.delete_entity_from_pdp_project")
  entity = {
    "seed": {
      "records": 1000,
      "recordSize": 8096
    },
    "name": "Cascade",
    "type": "random-generator-connector",
    "pipelineId": "fakeidCascadePipeline",
    "properties": {
      "index": "random_generated_docs"
    },
    "batchSize": "100",
    "id": "fakeidCascadeSeed"
  }
  config = {'project_path': test_project_path()}
  delete_references_from_entity(config, SEED, entity, True)
  assert mock_delete.call_count == 0


def test_delete_references_from_entity_entities_no_deleted(mocker, test_project_path, mock_custom_exception):
  """
  Test the function defined in :func:`commons.pdp_products.delete_references_from_entity`,
  when some entities were not deleted from the PDP API.
  """
  mocker.patch("commons.pdp_products.has_pdp_project_structure", return_value=True)
  mocker.patch(
    "commons.pdp_products.search_value_from_entity",
    return_value=[None, 'fakeidCascadePipeline', 'fakeid']
  )
  mocker.patch("commands.config.get.get_entity_by_id", side_effect=[({}, {}), (None, None)])
  mock_delete = mocker.patch("commons.pdp_products.delete_entity_from_pdp_project")
  entity = {
    "seed": {
      "records": 1000,
      "recordSize": 8096
    },
    "name": "Cascade",
    "type": "random-generator-connector",
    "pipelineId": "fakeidCascadePipeline",
    "properties": {
      "index": "random_generated_docs"
    },
    "batchSize": "100",
    "id": "fakeidCascadeSeed"
  }
  config = {'project_path': test_project_path()}
  delete_references_from_entity(config, SEED, entity, True)
  assert mock_delete.call_count == 0


def test_get_entity_by_id_from_local_entities_no_project_structure(mocker):
  """
  Test the function defined in :func:`commons.pdp_products.get_entity_by_id_from_local_entities`,
  when the project path doesn't have a PDP project structure
  """
  mocker.patch("commons.pdp_products.has_pdp_project_structure", return_value=False)
  mock_handle_and_exit = mocker.patch("commons.pdp_products.handle_and_exit")
  get_entity_by_id_from_local_entities("", SEED, 'fakeid')
  assert mock_handle_and_exit.call_count == 0


def test_search_value_from_entity_not_a_dict_or_list():
  """
  Test the function defined in :func:`commons.pdp_products.get_entity_by_id_from_local_entities`,
  when the project path doesn't have a PDP project structure
  """
  with pytest.raises(DataInconsistency) as error:
    search_value_from_entity('fake-property', 76)
  assert error.value.message == "The given vale is type int."
