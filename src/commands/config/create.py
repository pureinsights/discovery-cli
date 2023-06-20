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
import json
import os.path

import click

from commons.console import create_spinner, print_console, print_warning, spinner_change_text, spinner_fail, spinner_ok, \
  suppress_errors, suppress_warnings
from commons.custom_classes import PdpEntity
from commons.file_system import has_pdp_project_structure, read_entities, write_entities
from commons.handlers import handle_and_continue, handle_and_exit
from commons.pdp_products import are_same_pdp_entity, create_or_update_entity, get_all_entities_names_ids, \
  identify_entity, \
  json_to_pdp_entities, replace_ids, replace_names_by_ids
from commons.raisers import raise_for_pdp_data_inconsistencies


def interactive_input(placeholder: str):
  """
  Opens the default text editor with a placeholder.
  :param str placeholder: A text that will be showed within the file in the text editor.
  :rtype: list[dict]
  :return: Try to parse the text entered by the user in JSON format to PDP entities.
  """
  result: str | None = click.edit(placeholder)
  if result is not None:
    return json_to_pdp_entities(result)
  return json_to_pdp_entities(placeholder)


def input_stage(file_path: str | None, interactive: bool) -> tuple[str, list[dict]]:
  """
  This stage is responsible for reading the configuration of the entity to be created.
  :param str file_path: The path to the configuration of the entity.
  :param bool interactive: This will launch an editor to allow the user edit the configuration before create the entity.
  :rtype: tuple[str, list[dict]]
  :return: A tuple with the first element representing from where the entities were read, and the second element the
           entities read.
  """
  if file_path is None:
    if not interactive:
      return '"given by user"', []
    placeholder = [
      {
        "name": "",
        "description": "",
        "type": ""
      }
    ]

    return '"given by user"', interactive_input(json.dumps(placeholder, indent=2))

  _, entities = handle_and_exit(read_entities,
                                {'message': 'Could not read the entity.', 'show_exception': True},
                                file_path)
  if interactive:
    return file_path, interactive_input(json.dumps(entities, indent=2))

  return file_path, entities


def parsing_stage(project_path: str, entity_type: PdpEntity, entities: list[dict], file_path: str) -> list[dict]:
  """
  This stage is responsible to prepare the entities to the deployment (replace names by ids).
  :param str project_path: The path to the root of the project.
  :param PdpEntity entity_type: The entity type of the entities that will be parsed.
  :param list[dict] entities: The entities that will be parsed.
  :param str file_path: The file from the entities were read.
  :rtype: list[dict]
  :return: The entities with all template names replaced by the respective ids.
  """
  if not has_pdp_project_structure(project_path):
    return entities

  ids_names = {}
  for entity in entities:
    name = entity.get('name', None)
    _id = entity.get('id', None)
    if name is not None:
      ids_names[name] = entity
    if _id is not None:
      ids_names[_id] = entity
  raise_for_pdp_data_inconsistencies(project_path, ids_names)

  parsed_entities = []
  ids_names = get_all_entities_names_ids(project_path, parsed_entities)

  entity_type = copy.deepcopy(entity_type)  # Creates a copy to avoid modify the original entity_type
  entity_type.associated_file_name = file_path
  replace_names_by_ids(entity_type, entities, ids_names)

  return entities


def writing_stage(project_path: str, entity_type: PdpEntity, entities: list[dict]) -> list[dict]:
  """
  This stage is the responsible to write the entities created.
  :param str project_path: The path to the root of the project.
  :param PdpEntity entity_type: The entity type of the entities to be written.
  :param list[dict] entities: The list of entities that will be written.
  :rtype: list[dict]
  :return: The list of entities written.
  """
  if not has_pdp_project_structure(project_path):
    return []

  file_path = os.path.join(project_path, entity_type.product.title(), entity_type.associated_file_name)

  entities_read = read_entities(file_path)
  new_entities = []

  # Delete duplicated items
  for entity in entities:
    is_duplicated = False
    for entity_read in entities_read:
      if are_same_pdp_entity(entity, entity_read):
        if entity.get('id', None) is not None and entity_read.get('id', None) == entity.get('id', None):
          entities_read.remove(entity_read)
          break
        else:
          print_warning(
            f"Entity {identify_entity(entity)} already exists on {entity_type.associated_file_name}."
            " So it won't be added."
          )
        is_duplicated = True
        break
    if not is_duplicated:
      new_entities += [entity]

  write_entities(file_path, entities_read + new_entities)
  product_path = os.path.join(project_path, entity_type.product.title())
  replace_ids(product_path, suppress_warnings=True)
  return new_entities


def deployment_stage(config: dict, entity_type: PdpEntity, entities: list[dict], ignore_ids: bool, verbose: bool):
  """
  This stage is the responsible to deploy the entities.
  :param dict config: The configuration containing the url of the product's APIs.
  :param PdpEntity entity_type: The entity type of the entities that will be deployed.
  :param list[dict] entities: The list of entities to deploy.
  :param bool ignore_ids: If True will always create new instances of the entities,
         if not will update the entities with ids.
  :param bool verbose: Determines the level of information showed.
  :rtype: list[dict]
  :return: A list of entities that were deployed correctly
  """
  deployed_entities = []
  action = 'deployed'  # This is used to let know the user if the entity is being created or updated.
  for entity in entities:
    if verbose:
      create_spinner()
      spinner_change_text(f"Deploying entity {identify_entity(entity)}...")

    if ignore_ids:
      entity.pop('id', None)

    if entity.get('id', None) is not None:
      action = 'updated'

    product_url = config.get(entity_type.product, f'Wrong product name {entity_type.product}')
    _, _id = handle_and_continue(
      create_or_update_entity, {'show_exception': verbose}, product_url, entity_type.type,
      entity, verbose=False
    )

    if _id is None:
      if verbose:
        spinner_fail(
          f"Entity {click.style(identify_entity(entity), fg='blue')} could {click.style('not', fg='red')} be {action}."
        )
      continue
    if verbose:
      spinner_ok(
        f"Entity {click.style(identify_entity(entity), fg='blue')} {action} with id {click.style(_id, fg='green')}."
      )
    entity['id'] = _id
    deployed_entities += [entity]

  return deployed_entities


def printing_stage(entity_type: PdpEntity, entities: list[dict], _json: bool, pretty: bool):
  """
  This stage is the responsible to print the information related with the creation of the entities.
  :param PdpEntity entity_type: The entity type of the entities.
  :param list[dict] entities: The list of entities to print.
  :param bool _json: If True, the stage will only print an array with JSON objects. Warnings and errors are suppressed.
  :param bool pretty: If True, the result will be showed in  a human-readable JSON format.
  """
  if pretty:
    return print_console(json.dumps(entities, indent=2))

  if _json:
    return print_console(entities)

  for entity in entities:
    entity_str = click.style(identify_entity(entity), fg='blue')
    print_console(f'Entity {entity_str} added to {click.style(entity_type.associated_file_name, fg="green")}.')


def run_deployment(config: dict, entity_type: PdpEntity, entities: list[dict], ignore_ids: bool, _json: bool):
  """
  Runs the deployment stage and returns the entities which should be written.
  :param dict config: The configuration of the project, containing the url of the products APIs.
  :param PdpEntity entity_type: The entity type of the entities that will be added to the project.
  :param list[dict] entities: A list with the entities to deploy.
  :param bool _json: If True, the result will be showed as a JSON.
  :param bool ignore_ids: If True, will try to create new instances of the entities.
  """
  entities_copy = [{**entity} for entity in entities]
  # Deploys the entities list
  entities_deployed = deployment_stage(config, entity_type, entities_copy, ignore_ids, not _json)

  entities_to_write = []
  for entity_deployed in entities_deployed:
    for entity in entities:
      # If entity contains all the attributes with the same value as entity_deployed, then is the same entity
      if are_same_pdp_entity(entity, entity_deployed):
        entity['id'] = entity_deployed.get('id', None)
        entities_to_write += [entity]
        entities.remove(entity)
        break
  return entities_to_write


def run(config: dict, project_path: str, entity_type: PdpEntity, file: str, has_to_deploy: bool,
        _json: bool, pretty: bool, ignore_ids: bool, interactive: bool):
  """
  Add to a project and deploy entities from different sources. Templates, files and interactive mode.
  :param dict config: The configuration of the project, containing the url of the products APIs.
  :param str project_path: The path to the root of the PDP project.
  :param PdpEntity entity_type: The entity type of the entities that will be added to the project.
  :param str file: The path to the location of the entities' configuration.
  :param bool has_to_deploy: If True, the entities will be deployed to the respective product API.
  :param bool _json: If True, the result will be showed as a JSON.
  :param bool pretty: If True, the result will be showed in  a human-readable JSON format.
  :param bool ignore_ids: If True, will try to create new instances of the entities.
  :param bool interactive: Will open a text editor with a placeholder of the entity configuration.
  """
  suppress_errors(_json)
  suppress_warnings(_json)

  # Read the entities to create
  _file, entities = input_stage(file, interactive)

  # Removes the id property if ignore-ids is activated
  if ignore_ids:
    for entity in entities:
      entity.pop('id', None)

  if has_to_deploy and has_pdp_project_structure(project_path):
    # Replace all the {{ fromName }} template, with the respective ids.
    entities = parsing_stage(project_path, entity_type, entities, file)

  # If the flag is true, then deploys the entities to the products
  if has_to_deploy:
    entities = run_deployment(config, entity_type, entities, ignore_ids, _json)

  if not has_pdp_project_structure(project_path) and not has_to_deploy:
    print_warning('The command create must be called within a PDP project to add the entity to the files.')

  # Writes the entities to the pdp project
  entities = writing_stage(project_path, entity_type, entities)

  # Show the entities deployed or entities read
  printing_stage(entity_type, entities, _json, pretty)
