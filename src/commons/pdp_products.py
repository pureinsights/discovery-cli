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
import os
import re
import zipfile

import click

from commons.console import print_console, print_error, print_exception, print_warning, verbose
from commons.constants import CORE, DISCOVERY_PROCESSOR, ENTITIES, FROM_NAME_FORMAT, INGESTION, INGESTION_PROCESSOR, \
  PRODUCTS, STAGING, \
  URL_CREATE, \
  URL_DELETE, URL_EXPORT_ALL, \
  URL_GET_BY_ID, URL_UPDATE, WARNING_SEVERITY
from commons.custom_classes import DataInconsistency, PdpEntity, PdpException
from commons.file_system import has_pdp_project_structure, read_entities, write_entities
from commons.handlers import handle_and_continue, handle_and_exit
from commons.http_requests import delete, get, post, put
from commons.raisers import raise_file_not_found_error
from commons.utils import flat_list


def identify_entity(entity: dict, fields=['name', 'id', 'description', 'type'], **kwargs):
  """
  Returns a string which makes easier for the user identify the entity.

  :param dict entity: The entity to identify.
  :param list[str] fields: The fields to try to return.
  :key
  :rtype: str
  :return: The name of the field and the value. If none of the fields are in the entity itself will be returned.
  """
  format_str = kwargs.get('format', '{field} "{ref}"')
  if type(entity) is not dict:
    return entity
  default_value = kwargs.get('default', entity)
  for field in fields:
    ref = entity.get(field, None)
    if ref is not None:
      return format_str.format(field=field, ref=ref)

  return default_value


def associate_values_from_entities(from_entity: dict, from_field: str, to_entity: dict, to_field: str,
                                   entity_type: PdpEntity = None, suppress_warnings: bool = False):
  """
  Associate two values from to entities. Helpful to associate the
  id of an entity with the name of the same entity.

  :param dict from_entity: The entity to get the from_field.
  :param str from_field: The name of the field to be used as key.
  :param dict to_entity: The entity to get the to_field.
  :param str to_field: The name of the field to be used as value.
  :rtype: tuple[any,any]
  :return: A tuple representing a key-value.
  :raises DataInconsistency: Where an entity doesn't have a field. It's raised as WARNING.
  """
  value_from = from_entity.get(from_field, None)
  value_to = to_entity.get(to_field, None)

  message = 'Entity with {entity} does not have field {field}.'

  if entity_type is not None:
    message += f' Entity on file {entity_type.associated_file_name}'

  if value_from is None and not suppress_warnings:
    print_exception(
      DataInconsistency(message=message.format(entity=identify_entity(from_entity), field=from_field),
                        severity=WARNING_SEVERITY, handled=True)
    )

  if value_to is None and not suppress_warnings:
    print_exception(
      DataInconsistency(message=message.format(entity=identify_entity(to_entity), field=to_field),
                        severity=WARNING_SEVERITY, handled=True)
    )

  return value_from, value_to


def replace_ids(path: str, ids=None, **kwargs):
  """
  Reads all the files of all entities and reformat the file to be readable for humans.
  Also, it replaces the ids referenced in other entities by {{ fromName('{name}') }}.

  :param str path: A path to the entity files.
  :param str ids: A dict with ids as keys and names as values.
  :rtype: dict
  :return: The same ids dict but updated.
  """
  suppress_warnings = kwargs.get('suppress_warnings', False)
  if ids is None:
    ids = {}
  *_, dir_name = os.path.split(path)
  for entity_type in ENTITIES:
    if dir_name == entity_type.product.title():
      file_path = os.path.join(path, entity_type.associated_file_name)
      with open(file_path, 'r+') as file:
        entities = json.load(file)
        if type(entities) is not list:
          entities = [entities]
        success, new_ids = handle_and_continue(replace_ids_for_names, {'show_exception': True},
                                               entity_type, entities, ids, suppress_warnings=suppress_warnings)
        if success:
          ids = new_ids
        file.seek(0)
        json.dump(entities, file, indent=2)
        file.truncate()
  return ids


def replace_names_by_ids(entity_type: PdpEntity, entities: list[dict], names: dict, **kwargs):
  """
  Replaces the {{ fromName('entity_name') }} with the actual id of the entity.

  :param PdpEntity entity_type: Is the type of the entity where you're going to replace the name.
  :param list[dict] entities: The list of entities where your going to replace the names.
  :param dict names: A dict that contains the name of previous entities associated with their respective ids.
  :rtype: dict
  :return: Return the names with the names of entities associated too.
  """
  suppress_warnings = kwargs.get('suppress_warnings', False)
  for entity in entities:
    if 'id' in entity.keys():
      name, _id = associate_values_from_entities(entity, 'name', entity, 'id', entity_type, True)
      if name is not None:
        # We let associate a name with a None to let the associate_value print a warning on further steps
        names[name] = _id
      # If _id is not None, we associate the id with the id to avoid covert the case
      # where an entity has a reference to the id and not to the name of the entity
      if _id is not None:
        names[_id] = _id
    replace_value(entity_type, entity, names, from_field='name', format='{0}', suppress_warnings=suppress_warnings)

  return names


def replace_ids_for_names(entity_type: PdpEntity, entities: list[dict], ids: dict, **kwargs):
  """
  Replaces all the ids in an entity referencing another entity for the name of the another
  entity with the format {{ fromName('{name}') }}. And also adds his id with his name to
  the ids' dict.

  :param PdpEntity entity_type: A class representing the type of the entity.
  :param list[dict] entities: A list with all the entities to replaces de ids.
  :param dict ids: A dictionary containing the ids as keys and the names as values.
  :rtype: dict
  :return: The ids dict with his id and his name as a key-value.
  """
  suppress_warnings = kwargs.get('suppress_warnings', False)
  for entity in entities:
    success, response = handle_and_continue(associate_values_from_entities, {'show_exception': True}, entity, 'id',
                                            entity, 'name', entity_type, suppress_warnings)
    if success:
      _id, name = response
      if _id is not None:
        ids[_id] = name
    replace_value(entity_type, entity, ids, suppress_warnings=suppress_warnings)
  return ids


def show_expected_error_if_value_not_found(entity_type: PdpEntity, entity: dict, from_field: str,
                                           value: str):
  """
  Show a warning when the value that you need to replace is not on the values of "replace_value" function.
  It shows two types of message, when teh entity is identifiable and when is not.

  :param PdpEntity entity_type: The type of the entity. Used to show the file associated to.
  :param dict entity: The entity to identify where the value is missing.
  :param str from_field: The name of the field where the replace_value function was trying to replace.
  :param str value: The value that does not have any value associated on the replace_value function.
  """
  name = identify_entity(entity, default=entity_type.associated_file_name)
  if name != entity_type.associated_file_name:
    print_warning(f'Value "{value}" does not exist while attempting to replace in field "{from_field}". '
                  f'Entity {name} in file {entity_type.associated_file_name}.'
                  f' That could means that the name "{value}" do not exists or the entity "{value}" do not have an Id.')
  else:
    print_warning(f'Value "{value}" does not exist while attempting to replace in field "{from_field}". '
                  f'Entity has no name and no id in file {entity_type.associated_file_name}.'
                  f' That could means that the name "{value}" do not exists or the entity "{value}" do not have an Id.')


def replace_value(entity_type: PdpEntity, entity: any, values: dict, **kwargs):
  """
  Replaces the value from a field of an entity, for the value of another field with a
  specified format.

  :param PdpEntity entity_type: A class that represents the type of entity passed.
  :param any entity: The entity with the fields to ve replaced.
  :param dict values: A dict where the keys are the value of the from_field and the value for the key
                      is the value of the to_field.
  :key from_field: The name of the field where you want to get the value to replace.
                   Used just to show a correct message.
  :key to_fields: A list of the names fields where you want to replace the value with the given format
                  and the value of from_field.
  :key format: The format that the replaced field will have. Default is "{{ fromName('{0}')"
  """
  from_field: str = kwargs.get('from_field', 'id')
  to_fields: list[str] = kwargs.get('to_fields', [ent.reference_field for ent in ENTITIES])
  format_str: str = kwargs.get('format', FROM_NAME_FORMAT)
  parent: dict = kwargs.get('parent', None)
  suppress_warnings = kwargs.get('suppress_warnings', False)

  if not isinstance(to_fields, (list, tuple, set)):
    to_fields = [to_fields]

  if entity is None:
    return

  # Calls the function recursively for each entity in the array.
  if isinstance(entity, (list, tuple, set)):
    for index, nested_entity in enumerate(entity):
      replace_value(entity_type, nested_entity, values, **{**kwargs, 'index': index})
    return

  # Calls the function recursively for each nested entity in the dictionary
  if type(entity) is dict:
    for key in entity.keys():
      replace_value(entity_type, entity.get(key, None), values, **{**kwargs, 'parent': entity,
                                                                   'from_field': key, 'index': None})
    return

  # Manage the entity when it's a string, integer or any other primitive type

  # Here from_field takes the value of the property associated with the primitive value
  # that's why from_field most be in to_fields
  if from_field not in to_fields:
    return
  entity = clear_from_name_format(entity)
  value = values.get(entity, None)

  if value is None:
    if entity not in values.keys() and not suppress_warnings:
      show_expected_error_if_value_not_found(entity_type, parent, from_field, entity)
    return

  index = kwargs.get('index', None)
  if index is not None:
    parent[from_field][index] = format_str.format(value)
    return

  parent[from_field] = format_str.format(value)


def export_all_entities(api_url: str, path: str, extract: bool = True, **kwargs):
  """
  Export all entities for a given product. (INGESTION, DISCOVERY or CORE)
  Downloads the zip and is extracted to the given path.

  :param str api_url: The url where will try to download the zip with the entities.
  :param str path: The path where the zip will be downloaded.
  :param bool extract: If is True the zip will be extracted and deleted.
  :key dict ids: A dictionary with ids of already extracted entities. Default is {}.
  """
  ids = kwargs.get('ids', {})
  zip_path = os.path.join(path, 'export.zip')
  product_export_response = get(URL_EXPORT_ALL.format(api_url))
  with open(zip_path, 'wb') as zip_file:
    zip_file.write(product_export_response)

  if extract:
    with zipfile.ZipFile(zip_path, 'r') as zip_ref:
      zip_ref.extractall(path)

    if os.path.exists(zip_path):
      os.remove(zip_path)

    success, new_ids = handle_and_continue(replace_ids, {'show_exception': True}, path, ids)

    if success:
      return {**ids, **new_ids}

  return ids


def clear_from_name_format(value: any, regex: str = r"\{\{\s*fromName\('(.+?)'\)\s*\}\}"):
  """
  It takes a template function, default is {{ fromName('param') }} and returns only the param.
  If the value doesn't match with the regex, then value is returned.
  :param any value: The value to clear the format. If value is not a str it returns the value itself.
  :param str regex: The regex to match and clear to retrieve the param.
  :rtype: str
  :return: The param of the given template function.
  """
  if type(value) is not str:
    return value

  search = re.search(f"{regex}", value)

  if search is None or len(search.groups()) < 1:
    return value

  return search.group(1)


def order_products_to_deploy(products: list[str] = None):
  """
  Orders a products list in order that is convenient to deployment.
  The only order that matters is that CORE must be before INGESTION
  :param list[str] products: A list of the products names to order.
  :rtype: list[str]
  :return: The list of products ordered.
  """
  # Actually the only order that matters is that CORE must be before INGESTION
  if products is None:
    return []

  if INGESTION not in products or CORE not in products:
    return products

  # If INGESTION and CORE are in products, then we have to assure that CORE is before INGESTION
  core_index = products.index(CORE)
  ingestion_old_index = products.index(INGESTION)
  products.insert(core_index + 1, products.pop(ingestion_old_index))

  return products


def create_or_update_entity(product_url: str, type: str, entity: dict, **kwargs):
  """
  Will deploy a new entity in the given product if the entity does not exist or if it has not
  an id field. If it has an id field and the id exists in the given product, then will be updated.
  :param str product_url: The URL for the product to deploy the entity.
  :param str type: The type of the entity to be deployed.
                   Must be the same that the "create new" endpoint expects.
  :param dict entity: The entity to be deployed.
  :key bool verbose: Will define the printing message strategy.
  """
  is_verbose = kwargs.get('verbose', False)
  successful_message = '{type} {entity} has been {action} successfully with id {id}.'

  try:
    entity_id = entity.get('id', None)

    # If the entity has an id then we need to verify if that id already exists
    if entity_id is not None:
      old_entity = get(URL_GET_BY_ID.format(product_url, entity=type, id=entity_id))

      # if an entity already exists with the same id, then we're going to update it
      if old_entity is not None:
        res = put(URL_UPDATE.format(product_url, entity=type, id=entity_id), json=entity)
        styled_action = click.style('updated', fg='blue')
        styled_id = click.style(entity_id, fg='green')
        verbose(
          verbose_func=lambda: print_console(
            successful_message.format(
              type=type.title(),
              entity=identify_entity(entity, ['name', 'id', 'description', 'type', 'expression']),
              action=styled_action,
              id=styled_id
            )
          ),
          verbose=is_verbose
        )
        return json.loads(res).get('id', None)

    # If the entity does not have an id or the entity does not exist on the API
    # we need to create it
    res = post(URL_CREATE.format(product_url, entity=type), json=entity)
    entity_id = json.loads(res).get('id', None)
    styled_id = click.style(entity_id, fg='green')
    styled_action = click.style('created', fg='blue')
    verbose(
      verbose_func=lambda: print_console(
        successful_message.format(
          type=type.title(),
          entity=identify_entity(entity, ['name', 'id', 'description', 'type', 'expression']),
          action=styled_action,
          id=styled_id
        )
      ),
      verbose=is_verbose
    )
    return entity_id
  except PdpException as exception:
    message = "\"{ref}\""
    print_error(
      f'Could not create entity {identify_entity(entity, default=type, format=message)} due to:\n'
      f'\t{exception.content.get("errors")}'
    )


def get_entity_type_by_name(entity_type_name: str) -> PdpEntity | None:
  """
  Returns the entity type based on the name of an entity.
  :param str entity_type_name: The name of the entity to infer the entity type.
  :rtype: PdpEntity
  :return: The entity type inferred by the given name.
  """
  if type(entity_type_name) is not str:
    return None

  if entity_type_name.lower() == INGESTION_PROCESSOR['name']:
    return INGESTION_PROCESSOR['entity']

  if entity_type_name.lower() == DISCOVERY_PROCESSOR['name']:
    return DISCOVERY_PROCESSOR['entity']

  filtered = [entity_type for entity_type in ENTITIES if entity_type.type == entity_type_name]
  if len(filtered) <= 0:
    return None
  return filtered.pop()


def get_all_entities_names_ids(project_path: str, entities: list[dict]) -> dict:
  """
  Reads all the entities on a project and returns a dictionary with the ids and names of the entities.
  :param str project_path: The path to the pdp project.
  :param list[dict] entities: A list of entities where the read entities will be added.
  :rtype: dict
  :return: A dictionary with the ids and names of the read entities.
  """
  raise_file_not_found_error(project_path)
  if not has_pdp_project_structure(project_path):
    return {}
  ids_names = {}
  products = [product for product in PRODUCTS['list'] if product != STAGING]

  for product in order_products_to_deploy(products):
    entity_types = PRODUCTS.get(product, {'entities': []}).get('entities')
    for entity_type in entity_types:
      file_path = os.path.join(project_path, product.title(), entity_type.associated_file_name)
      success, _entities = handle_and_continue(read_entities, {'show_exception': True}, file_path)
      if not success:
        continue
      entities += _entities

      replace_names_by_ids(entity_type, entities, ids_names, suppress_warnings=True)

  return ids_names


def are_same_pdp_entity(entity: dict, entity_clone: dict) -> bool:
  """
  Compares two PDP entities to determine if they are the same entity.
  :param dict entity: The first entity to compare.
  :param dict entity_clone: The second entity to compare.
  :rtype: bool
  :return: True if they are the same entity, False in other case.
  """
  if entity == entity_clone:
    return True

  entity_id = entity.get('id', None)
  entity_clone_id = entity_clone.get('id', None)
  if entity_id is not None and entity_clone_id is not None and entity_clone_id == entity_id:
    return True

  if [key for key in entity.keys() if key != 'id'] != [key for key in entity_clone.keys() if key != 'id']:
    return False

  return all(entity[key] == entity_clone.get(key, None) for key in entity.keys() if key != 'id')


def json_to_pdp_entities(entities_json: str, **kwargs) -> list[dict]:
  """
  Parse a str with JSON format and returns a list with the respective JSON objects as dictionaries.
  :param str entities_json: The str containing the list of entities in JSON format.
  :key message: A custom message if the parsing from JSON fails.
  :rtype: list[dict]
  :return: A list of dictionaries, or an empty list if no entities were on the entities_json.
  """
  message = kwargs.get('message', 'JSONDecodeError: Could not parse the JSON text. '
                                  'Please check the file has a valid JSON format.'
                       )
  _, entities = handle_and_exit(json.loads,
                                {
                                  'show_exception': True,
                                  'exception': PdpException(
                                    message=message)
                                },
                                entities_json)
  if type(entities) is not list:
    entities = [entities]
  return entities


def delete_pdp_entity(config: dict, entity_type: PdpEntity, entity: dict, cascade: bool, local: bool) -> bool:
  """
  Delete a given entity from the API and from the local project if the "local" param is true.
  :param dict config: Contains the project path and the PDP products' url.
  :param PdpEntity entity_type: The entity type of the entity to delete.
  :param dict entity: The entity to delete. Used to get the id and references to other entities.
  :param bool cascade: Will delete the entities referenced on "entity" in cascade.
  :param bool local: Will delete the configuration of the entity from the PDP project.
  :rtype: bool
  :return: True if the entity was deleted, False in other case.
  """
  product = entity_type.product
  entity_id = entity.get('id', None)
  if entity_id is None:
    return False

  acknowledge = delete(
    URL_DELETE.format(config[product], entity=entity_type.type, id=entity_id),
    params={'cascade': cascade}
  )
  was_deleted = json.loads(acknowledge).get('acknowledged', False)
  if was_deleted and local:
    delete_entity_from_pdp_project(config, entity_type, entity, cascade)
  return was_deleted


def delete_entity_from_pdp_project(config: dict, entity_type: PdpEntity, entity: dict, cascade: bool):
  """
  Deletes an entity from the configuration files on a pdp project.
  :param dict config: Contains the project path and the PDP products' url.
  :param PdpEntity entity_type: The entity type of the entity to delete.
  :param dict entity: The entity to delete. Used to get the id and references to other entities.
  :param bool cascade: Will delete the entities referenced on "entity" in cascade.
  """
  path = config['project_path']
  if not has_pdp_project_structure(path):
    return

  product = entity_type.product
  file_path = os.path.join(path, product.title(), entity_type.associated_file_name)
  entities = read_entities(file_path)
  none = {'none': None}
  index = 0
  for _entity in entities:
    entity_id = _entity.get('id', none)
    if entity_id is not none and entity_id == entity.get('id'):
      entities = entities[:index] + entities[index + 1:]
      break
    else:
      index += 1

  write_entities(file_path, entities)

  if not cascade:
    return

  delete_references_from_entity(config, entity_type, entity, cascade)


def delete_references_from_entity(config: dict, entity_type: PdpEntity, entity: dict, cascade: bool):
  """
  Deletes an entity from the configuration files on a pdp project.
  :param dict config: Contains the project path and the PDP products' url.
  :param PdpEntity entity_type: The entity type of the entity to delete.
  :param dict entity: The entity to delete. Used to get the id and references to other entities.
  :param bool cascade: Will delete the entities referenced on "entity" in cascade.
  """
  from commands.config.get import get_entity_by_id
  path = config['project_path']
  entity_references = entity_type.get_references()
  for reference_field in entity_references.keys():
    success, ids_found = handle_and_continue(search_value_from_entity, {}, reference_field, entity)
    if not success:
      continue
    referenced_ids = flat_list(ids_found)
    for _id in referenced_ids:
      if _id is None:
        continue
      entity_type = entity_references[reference_field]
      success, res = handle_and_continue(get_entity_by_id, {}, config, _id, [entity_type])
      if not success or (success and res[1] is not None):
        continue
      _, _entity = res
      handle_configuration = {
        'message': f'Could not delete entity {_id}.',
        'show_exception': True
      }
      _, entity = handle_and_continue(get_entity_by_id_from_local_entities, handle_configuration, path, entity_type,
                                      _id)
      if entity is None:
        continue
      handle_and_continue(delete_entity_from_pdp_project, {}, config, entity_type, entity, cascade)


def get_entity_by_id_from_local_entities(project_path: str, entity_type: PdpEntity, entity_id: str) -> dict | None:
  """
  Reads all the entities from a PDP project, and returns the entity with the same given id.
  :param str project_path: The path to the PDP project.
  :param PdpEntity entity_type: The entity type of the entity to search.
  :param str entity_id: The id to search on the local entities.
  :rtype: dict | None
  :return: A dictionary if the entity was found, None in other case.
  """
  if not has_pdp_project_structure(project_path):
    return None
  file_path = os.path.join(project_path, entity_type.product.title(), entity_type.associated_file_name)
  _, entities = handle_and_exit(read_entities, {}, file_path)
  for _entity in entities:
    if _entity.get('id', None) == entity_id:
      ids_names = get_all_entities_names_ids(project_path, [])
      replace_names_by_ids(entity_type, [_entity], ids_names)
      return _entity
  return None


def search_value_from_entity(entity_property: str, entity: dict | list) -> list[any]:
  """
  Search the value for the specified property.
  :param str entity_property: The name of the property to search for.
  :param dict | list entity: The entity or list of entities to search the value.
  :rtype: list[any]
  :return: A list of possible values found. None could be one of those values.
  """
  if isinstance(entity, (list, set, tuple)):
    values = []
    for nested_entity in entity:
      values += search_value_from_entity(entity_property, nested_entity)
    return values

  if not isinstance(entity, dict):
    raise DataInconsistency(
      message=f"The given vale is type {type(entity).__name__}.",
      content={'error': 'invalid type'}
    )
  values = []
  for key in entity.keys():
    nested_entity = entity[key]
    if key == entity_property:
      values += [nested_entity]
      break

    if isinstance(nested_entity, (list, set, tuple, dict)):
      values += search_value_from_entity(entity_property, nested_entity)
      continue

  return values
