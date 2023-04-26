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

import click

from commands.config.get import get_entity_by_id
from commons.console import print_console
from commons.constants import URL_GET_ALL
from commons.custom_classes import PdpEntity
from commons.handlers import handle_and_continue
from commons.http_requests import get
from commons.pdp_products import delete_pdp_entity, identify_entity


def delete_all_entities(config: dict, entity_types: list[PdpEntity], cascade: bool, local: bool) -> dict:
  """
  Try to delete and return a dict with those deleted entities.
  :param dict config: The configuration containing the product's url.
  :param list[PdpEntity] entity_types: A list of entity types from which the command will try to delete the entities.
  :param bool cascade: Will pass a boolean to the Pdp endpoint api.
  :param bool local: Will delete the configuration of the entity from the PDP project.
  :rtype: dict
  :return: A dict containing as keys the name of the entity type and as vale a list with the deleted entities.
  """
  entities = {}
  for entity_type in entity_types:
    product = entity_type.product
    entities[product] = entities.get(product, {})
    _, entities_found = handle_and_continue(get, {'show_exception': True},
                                            URL_GET_ALL.format(config[product], entity=entity_type.type))
    if entities_found is None:
      continue

    entities_found = json.loads(entities_found).get('content', [])
    entities[product][entity_type.type] = []
    for entity in entities_found:

      _, deleted = handle_and_continue(delete_pdp_entity, {'show_exception': True}, config, entity_type,
                                       entity, cascade, local)
      if not deleted:
        continue

      entities[product][entity_type.type] += [entity]

  return entities


def delete_entities_by_ids(config: dict, entity_ids: list[str], entity_types: list[PdpEntity], cascade: bool,
                           local: bool):
  """
  Try to delete a list of entities of the given types.
  :param dict config: The configuration containing the product's url.
  :param list[str] entity_ids: The list of ids of each entity to delete.
  :param list[PdpEntity] entity_types: A list of entity types from which the command will try to delete the entities.
  :param bool cascade: If True, it will delete all the entities referenced in the entity
                       with an id contained in the given list of ids.
  :param bool local: If True it deletes the configuration entity from the PDP project files.
  :rtype: dict
  :return: A dictionary containing the entities deleted, categorized by product and entity type.
  """
  deleted_entities = {}
  for entity_id in [*entity_ids]:
    _, res = handle_and_continue(get_entity_by_id, {'show_exception': True}, config, entity_id,
                                 entity_types)

    if res is None:
      continue

    entity_type, entity = res
    if entity is None:
      continue

    _, deleted = handle_and_continue(delete_pdp_entity, {'show_exception': True}, config, entity_type,
                                     entity, cascade, local)
    if not deleted:
      continue

    entity_ids.remove(entity.get('id', 'None'))
    deleted_entities[entity_type.product] = deleted_entities.get(entity_type.product, {})
    deleted_entities[entity_type.product][entity_type.type] = deleted_entities[entity_type.product].get(
      entity_type.type, []
    )
    deleted_entities[entity_type.product][entity_type.type] += [entity]

  return deleted_entities


def delete_value_from_list(value: any, _list: list[any]):
  try:
    _list.remove(value)
  except ValueError:
    pass


def print_stage(deleted_entities: dict, entities_not_deleted: list[str]):
  """
  Prints information about which entities were deleted and which wasn't.
  :param dict deleted_entities: A dictionary containing the deleted entities categorized by product and entity type.
  :param list[str] entities_not_deleted: A list of ids of those not deleted entities.
  """
  for product in deleted_entities.keys():
    if len(deleted_entities[product].keys()) <= 0:
      continue
    print_console(f"{click.style(product.title(), fg='green')}:")

    for entity_type in deleted_entities[product].keys():
      if len(deleted_entities[product][entity_type]) <= 0:
        continue
      print_console(f"{click.style(entity_type.title(), fg='cyan')}:", prefix='  ')

      for entity in deleted_entities[product][entity_type]:
        _id = entity.get('id', None)
        delete_value_from_list(_id, entities_not_deleted)
        entity_styled = click.style(entity.get('id', identify_entity(entity)), fg='green')
        print_console(f"Entity {entity_styled} deleted successfully.", prefix='    ✔ ')

  for entity_id in entities_not_deleted:
    print_console(f"Entity {click.style(entity_id, fg='red')} could not be deleted.")


def run(config: dict, entity_types: [PdpEntity], entity_ids: list[str], cascade: bool, local: bool):
  """
  Deletes entities from the PDP products and from the pdp project.
  :param dict config: The configuration containing the product's url.
  :param list[str] entity_ids: The list of ids of each entity to delete.
  :param list[PdpEntity] entity_types: A list of entity types from which the command will try to delete the entities.
  :param bool cascade: If True, it will delete all the entities referenced in the entity
                       with an id contained in the given list of ids.
  :param bool local: If True it deletes the configuration entity from the PDP project files.
  """
  deleted_entities = {}
  entities_no_deleted = [*entity_ids]
  if len(entity_ids) > 0:
    deleted_entities = delete_entities_by_ids(config, entity_ids, entity_types, cascade, local)
    if entities_no_deleted == entity_ids:
      print_console(f"{click.style('No entities', fg='red')} were deleted.")
      return
  else:
    deleted_entities = delete_all_entities(config, entity_types, cascade, local)

  print_stage(deleted_entities, entities_no_deleted)
