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
#
#  Permission to use, copy, modify or distribute this software and its
#  documentation for any purpose is subject to a licensing agreement with
#  Pureinsights Technology Ltd.
#
#  All information contained within this file is the property of
#  Pureinsights Technology Ltd. The distribution or reproduction of this
#  file or any information contained within is strictly forbidden unless
#  prior written permission has been granted by Pureinsights Technology Ltd.

from commands.config.delete import delete_all_entities, delete_entities_by_ids
from commons.constants import ENDPOINT, INGESTION_PROCESSOR_ENTITY, PIPELINE, SEED
from pdp import load_config


def test_delete_all_entities(mocker):
  """
  Test the function defined in :func:`commands.config.delete.delete_all_entities`.
  """
  mocker.patch("commands.config.delete.get", side_effect=[True, None, True, True, True])
  mocker.patch("commands.config.delete.json.loads", side_effect=[
    {'content': [{'id': 'fake1'}, {'id': 'fake2'}]},
    {'content': [{'id': 'fake3'}, {'id': 'fake4'}]},
    {'content': [{'id': 'fake5'}]}
  ])
  mocker.patch("commands.config.delete.delete_pdp_entity", side_effect=[True, False, True, True, True])
  entity_types = [SEED, PIPELINE, ENDPOINT, INGESTION_PROCESSOR_ENTITY]
  config = load_config('config.ini', 'DEFAULT')
  deleted_entities = delete_all_entities(config, entity_types, False, False)
  expected_response = {
    'ingestion': {
      'seed': [{'id': 'fake1'}],
      'processor': [{'id': 'fake5'}]
    },
    'discovery': {
      'endpoint': [{'id': 'fake3'}, {'id': 'fake4'}]
    }
  }
  assert deleted_entities == expected_response


def test_delete_entities_by_ids(mocker):
  """
  Test the function defined in :func:`commands.config.delete.delete_all_entities`.
  """
  entities = [(SEED, {'id': 'fake1'}), (PIPELINE, {'id': 'fake2'}), (ENDPOINT, {'id': 'fake3'}),
              (ENDPOINT, {'id': 'fake4'}), (None, None), (ENDPOINT, {'id': 'fake5'})]
  was_deleted_list = [True, True, False, True, True]
  mocker.patch("commands.config.delete.get_entity_by_id", side_effect=entities)
  mocker.patch("commands.config.delete.delete_pdp_entity", side_effect=was_deleted_list)
  entity_types = [SEED]
  entity_ids = ['fake1', 'fake2', 'fake3', 'fake4', 'fake6', 'fake5']
  deleted_entities = delete_entities_by_ids({}, entity_ids, entity_types, False, False)
  assert deleted_entities == {
    'ingestion': {
      'seed': [{'id': 'fake1'}],
      'pipeline': [{'id': 'fake2'}]
    },
    'discovery': {
      'endpoint': [{'id': 'fake4'}, {'id': 'fake5'}]
    }
  }
  assert entity_ids == ['fake3', 'fake6']


def test_delete_entities_by_ids_get_failed(mocker, mock_custom_exception):
  """
  Test the function defined in :func:`commands.config.delete.delete_all_entities`.
  """
  mocker.patch("commands.config.delete.get_entity_by_id", side_effect=lambda *args: mock_custom_exception(Exception))
  mocker.patch("commands.config.delete.delete_pdp_entity")
  entity_types = [SEED]
  entity_ids = ['fake1']
  deleted_entities = delete_entities_by_ids({}, entity_ids, entity_types, False, False)
  assert deleted_entities == {}
  assert entity_ids == ['fake1']
