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
import json

from commands.config.get import get_entities_by_ids
from commons.constants import URL_GET_ALL
from commons.custom_classes import PdpEntity
from commons.http_requests import get
from commons.pdp_products import delete_pdp_entity, get_entity_type_by_name


def delete_all_entities(config: dict, entity_types: list[PdpEntity], cascade: bool) -> dict:
  """
  Try to delete and return a dict with those deleted entities.
  :param dict config: The configuration containing the product's url.
  :param list[PdpEntity] entity_types: A list of entity types from which the command will try to delete the entities.
  :param bool cascade: Will pass a boolean to the Pdp endpoint api.
  :rtype: dict
  :return: A dict containing as keys the name of the entity type and as vale a list with the deleted entities.
  """
  entities = {}
  for entity_type in entity_types:
    product = entity_type.product
    entities[product] = entities.get(product, {})
    entities_found = get(URL_GET_ALL.format(config[product], entity=entity_type.type))
    if entities_found is None:
      continue

    entities_found = json.loads(entities_found).get('content', [])
    entities[product][entity_type.type] = []
    for entity in entities_found:

      if not delete_pdp_entity(config, entity_type, entity.get('id', False), cascade):
        continue
      entities[product][entity_type.type] += [entity]

  return entities


def delete_entities_by_ids(config, entity_ids, entity_types, cascade):
  entities_to_delete = get_entities_by_ids(config, entity_ids, entity_types, {}, False)
  deleted_entities = {}
  for product in entities_to_delete.keys():
    for entity_type_name in entities_to_delete[product]:
      entity_type = get_entity_type_by_name(entity_type_name)
      deleted_entities[entity_type.product] = deleted_entities.get(entity_type.product, {})
      deleted_entities[entity_type.product][entity_type.type] = []
      for entity in entities_to_delete[product][entity_type_name]:
        if not delete_pdp_entity(config, entity_type, entity.get('id', False), cascade):
          continue

        deleted_entities[entity_type.product][entity_type.type] += [entity]

  return deleted_entities


def run(config: dict, entity_types: [PdpEntity], entity_ids: list[str], cascade: bool):
  deleted_entities = {}

  if len(entity_ids) > 0:
    deleted_entities = delete_entities_by_ids(config, entity_ids, entity_types, cascade)
  else:
    deleted_entities = delete_all_entities(config, entity_types, cascade)

  print(deleted_entities)
