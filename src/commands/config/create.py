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
import copy

from commons.custom_classes import PdpEntity
from commons.file_system import has_pdp_project_structure, read_entities
from commons.handlers import handle_and_continue, handle_and_exit
from commons.pdp_products import create_or_update_entity, get_all_entities_names_ids, replace_names_by_ids


def input_stage(file_path: str):
  if file_path is None:
    return '"given by user"', []

  _, entities = handle_and_exit(read_entities,
                                {'message': 'Could not read the entity.', 'show_exception': True},
                                file_path)
  return file_path, entities


def parsing_stage(project_path: str, entity_type: PdpEntity, entities: list[dict], file_path: str):
  if not has_pdp_project_structure(project_path): return entities

  parsed_entities = []
  ids_names = get_all_entities_names_ids(project_path, parsed_entities)

  entity_type = copy.deepcopy(entity_type)  # Creates a copy to avoid modify the original entity_type
  entity_type.associated_file_name = file_path
  replace_names_by_ids(entity_type, entities, ids_names)

  return parsed_entities


def writing_stage(project_path: str, entity_type: PdpEntity, entities: list[dict]):
  if not has_pdp_project_structure(project_path): return


def deployment_stage(config: dict, entity_type: PdpEntity, entities: list[dict]):
  deployed_entities = []
  for entity in entities:
    product_url = config.get(entity_type.product, f'Wrong product name {entity_type.product}')
    _, _id = handle_and_continue(
      create_or_update_entity, {'show_exception': True}, product_url, entity_type.type,
      entity, verbose=False
    )

    if _id is None:
      continue
    deployed_entities += [entity]

  return deployed_entities


def run(config: dict, project_path: str, entity_type: PdpEntity, file: str, has_to_deploy: bool, json: bool):
  file, entities = input_stage(file)
  parsing_stage(project_path, entity_type, entities, file)
  if has_to_deploy:
    resulting_entities = deployment_stage(config, entity_type, entities)
  print(entities)
