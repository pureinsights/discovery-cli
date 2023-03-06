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

from commons.constants import CREDENTIAL, ENDPOINT, ENDPOINT_PROCESSOR, PIPELINE, PROCESSOR, SCHEDULER, SEED, \
  WARNING_SEVERITY
from commons.custom_classes import DataInconsistency, PdpException
from commons.pdp_products import associate_values_from_entities, export_all_entities, identify_entity, replace_ids, \
  replace_ids_for_names, \
  replace_value


def test_identify_entity_is_not_dict():
  """
  Test the function defined in :func:`commons.pdp_products.identify_entity`,
  when the entity is not a dict.
  """
  entity = 'fake-entity'
  response = identify_entity(entity)
  assert response == entity


@pytest.mark.parametrize('entity, property', [
  ({ 'type': 'fake-type' }, 'type'),
  ({ 'description': 'fake-description' }, 'description'),
  ({ 'id': 'fake-id' }, 'id'),
  ({ 'name': 'fake-name' }, 'name')
])
def test_identify_entity_return_successfully(entity, property):
  """
  Test the function defined in :func:`commons.pdp_products.identify_entity`.
  """
  response = identify_entity(entity)
  assert response == f'{property} {entity[property]}'


def test_identify_entity_return_default():
  """
  Test the function defined in :func:`commons.pdp_products.identify_entity`,
  when the entity does not have the property.
  """
  entity = { }
  response = identify_entity(entity, default='fake-default')
  assert response == 'fake-default'


@pytest.mark.parametrize('entity1, entity2', [
  ({ 'property': 'property' }, { }),
  ({ }, { 'property': 'property' })
])
def test_associate_values_from_entities_raises_DataInconsistency(entity1, entity2):
  """
  Test the function defined in :func:`commons.pdp_products.associate_values_from_entities`,
  when the entities do not have the properties.
  """
  with pytest.raises(DataInconsistency) as exception:
    associate_values_from_entities(entity1, 'property', entity2, 'property')
  assert exception.value.message == f'Entity with {entity1} does not have field property.' or \
         exception.value.message == f'Entity with {entity2} does not have field property.'
  assert exception.value.severity == WARNING_SEVERITY


@pytest.mark.parametrize('entity1, entity2', [
  ({ 'property': 1 }, { 'property': 2 }),
  ({ 'property': 2 }, { 'property': 1 }),
  ({ 'property': 1 }, { 'property': 1 }),
  ({ 'property': 2 }, { 'property': 2 })
])
def test_associate_values_from_entities_successful(entity1, entity2):
  """
  Test the function defined in :func:`commons.pdp_products.associate_values_from_entities`.
  """
  response = associate_values_from_entities(entity1, 'property', entity2, 'property')
  assert (entity1['property'], entity2['property']) == response


@pytest.mark.parametrize('dir_name, entity_str, entities, call_count, ids', [
  ('./Ingestion', '[{"id":"fake-id","name":"fake-name"}]', [{ }], 4, { }),
  ('./Ingestion', '[{"id":"fake-id","name":"fake-name"}]', { }, 4, None),
  ('Ingestion', '[{"id":"fake-id","name":"fake-name"}]', [{ }], 4, { }),
  ('fake-path', '[{"id":"fake-id","name":"fake-name"}]', [{ }], 0, { }),
  ('./Discovery', '[{"id":"fake-id","name":"fake-name"}]', [{ }], 2, None),
  ('Core', '[{"id":"fake-id","name":"fake-name"}]', { }, 1, None),
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


@pytest.mark.parametrize('entity_type, entities, expected_ids', [
  (PROCESSOR, [{ 'id': 'fake-id4', 'name': 'fake-name4' }], { 'fake-id4': 'fake-name4' }),
  (PIPELINE, [{
    'id': 'fake-id4',
    'name': 'fake-name4',
    'steps': [{ 'processorId': 'fake-id1' }, { 'processorId': 'fake-id2' }, { 'processorId': 'fake-id4' }]
  }], { 'fake-id4': 'fake-name4' }),
  (SEED, [{ }], { }),
  (SCHEDULER, [{ 'id': 'fake-id4', 'name': 'fake-name4' }], { 'fake-id4': 'fake-name4' }),
  (CREDENTIAL, [{ 'id': 'fake-id4', 'name': 'fake-name4' }], { 'fake-id4': 'fake-name4' }),
  (ENDPOINT, [{ 'id': 'fake-id4', 'name': 'fake-name4' }], { 'fake-id4': 'fake-name4' }),
  (ENDPOINT_PROCESSOR, [{ 'id': 'fake-id4', 'name': 'fake-name4' }], { 'fake-id4': 'fake-name4' })
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
  expected_ids = { **expected_ids, **ids }
  response = replace_ids_for_names(entity_type, entities, ids)
  assert response == expected_ids


@pytest.mark.parametrize('entity_type, entity, expected_entity', [
  (PIPELINE, {
    'id': 'fake-id7',
    'name': 'fake-name7',
    'steps': { 'processorId': 'fake-id1' }
  }, {
     'id': 'fake-id7',
     'name': 'fake-name7',
     'steps': { 'processorId': '{{ fromName(\'fake-name1\') }}' }
   }),
  (PIPELINE, {
    'id': 'fake-id7',
    'name': 'fake-name7',
    'steps': { 'processorId': None }
  }, {
     'id': 'fake-id7',
     'name': 'fake-name7',
     'steps': { 'processorId': None }
   }),
  (PIPELINE, {
    'id': 'fake-id7',
    'name': 'fake-name7',
    'fake-child': { 'fake-child': { 'processorId': 'fake-id1' } }
  }, {
     'id': 'fake-id7',
     'name': 'fake-name7',
     'fake-child': { 'fake-child': { 'processorId': '{{ fromName(\'fake-name1\') }}' } }
   }),
  (PIPELINE, {
    'id': 'fake-id7',
    'name': 'fake-name7',
    'fake-child': { 'fake-children': [{ 'processorId': 'fake-id1' }] }
  }, {
     'id': 'fake-id7',
     'name': 'fake-name7',
     'fake-child': { 'fake-children': [{ 'processorId': '{{ fromName(\'fake-name1\') }}' }] }
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
    'fake-id6': 'fake-name6'
  }
  replace_value(entity_type, entity, values)
  assert entity == expected_entity


@pytest.mark.parametrize('entity_type, entity, expected_message', [
  (PIPELINE, {
    'id': 'fake-id7',
    'name': 'fake-name7',
    'steps': { 'processorId': 'fake-id8' }
  }, f'[ERROR]: Id "fake-id8" does not exist while attempting to replace in processorId.' \
     f' Child of entity "name fake-name7" in file pipelines.json.'
   ),
  (SEED, {
    'pipelineId': 'fake-id8'
  }, f'[ERROR]: Id "fake-id8" does not exist while attempting to replace in pipelineId.' \
     f' Entity has no name and no id in file seeds.json.'
   ),
  (SCHEDULER, {
    'id': 'fake-id7',
    'name': 'fake-name7',
    'seedId': 'fake-id8'
  }, f'[ERROR]: Id "fake-id8" does not exist while attempting to replace in seedId.' \
     f' Entity "name fake-name7" in file cron_jobs.json.'
   )
])
def test_replace_value_failed(entity_type, entity, expected_message):
  """
  Test the function defined in :func:`commons.pdp_products.replace_value`,
  when the value to replace is not present on the values.
  """
  values = {
    'fake-id1': 'fake-name1',
    'fake-id2': 'fake-name2',
    'fake-id3': 'fake-name3',
    'fake-id4': 'fake-name4',
    'fake-id5': 'fake-name5',
    'fake-id6': 'fake-name6'
  }
  with pytest.raises(PdpException) as exception:
    replace_value(entity_type, entity, values, to_fields=['processorId', 'pipelineId', 'seedId'])
  assert exception.value.message == expected_message


def test_export_all_entities_successful_without_extract(mocker):
  """
  Test the function defined in :func:`commons.pdp_products.export_all_entities`,
  without extracting the files.
  """
  content = b'fake-content'
  mocker.patch('commons.pdp_products.get', return_value=content)
  m = mock_open()
  mocker.patch('commons.pdp_products.open', m)
  api = 'http://fake-url'
  path = 'fake-path'
  ids = { 'id': 'name' }
  response = export_all_entities(api, path, False, ids=ids)
  m.assert_called_once_with(os.path.join(path, 'export.zip'), 'wb')
  m().write.assert_called_once_with(content)
  assert ids == response


def test_export_all_entities_successful_with_extraction(mocker, mock_path_exists):
  """
  Test the function defined in :func:`commons.pdp_products.export_all_entities`,
  extracting the files.
  """
  mock_path_exists(True)
  content = b'fake-content'
  mocker.patch('commons.pdp_products.get', return_value=content)
  m = mock_open()
  mocker.patch('commons.pdp_products.open', m)
  mocker.patch('commons.pdp_products.zipfile')
  mocker.patch('commons.pdp_products.os.remove')
  mocker.patch('commons.pdp_products.handle_and_continue', return_value=(True, { }))
  api = 'http://fake-url'
  path = 'fake-path'
  ids = { 'id': 'name' }
  response = export_all_entities(api, path, True, ids=ids)
  m.assert_called_once_with(os.path.join(path, 'export.zip'), 'wb')
  m().write.assert_called_once_with(content)
  assert ids == response


def test_export_all_entities_failed_with_extraction(mocker, mock_path_exists):
  """
  Test the function defined in :func:`commons.pdp_products.export_all_entities`,
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
  ids = { 'id': 'name' }
  response = export_all_entities(api, path, True, ids=ids)
  m.assert_called_once_with(os.path.join(path, 'export.zip'), 'wb')
  m().write.assert_called_once_with(content)
  assert ids == response
