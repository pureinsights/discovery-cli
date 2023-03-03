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
import pytest

from commons.constants import WARNING_SEVERITY
from commons.custom_classes import DataInconsistency
from commons.pdp_products import associate_values_from_entities, identify_entity


def test_identify_entity_is_not_dict():
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
  response = identify_entity(entity)
  assert response == f'{property} {entity[property]}'


def test_identify_entity_return_default():
  entity = { }
  response = identify_entity(entity, default='fake-default')
  assert response == 'fake-default'


@pytest.mark.parametrize('entity1, entity2', [
  ({ 'property': 'property' }, { }),
  ({ }, { 'property': 'property' })
])
def test_associate_values_from_entities_raises_DataInconsistency(entity1, entity2):
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
  response = associate_values_from_entities(entity1, 'property', entity2, 'property')
  assert (entity1['property'], entity2['property']) == response
