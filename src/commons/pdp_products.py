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

from commons.console import print_console, print_error, print_exception, print_warning
from commons.constants import CORE, ENTITIES, FROM_NAME_FORMAT, INGESTION, PRODUCTS, URL_CREATE, URL_EXPORT_ALL, \
  URL_GET_BY_ID, URL_UPDATE, WARNING_SEVERITY
from commons.custom_classes import DataInconsistency, PdpEntity, PdpException
from commons.file_system import read_entities
from commons.handlers import handle_and_continue
from commons.http_requests import get, post, put
from commons.raisers import raise_file_not_found_error


def identify_entity(entity: dict, fields=['name', 'id', 'description', 'type'], **kwargs):
  """
  Returns a string which makes easier for the user identify the entity.

  :param dict entity: The entity to identify.
  :param list[str] fields: The fields to try to return.
  :rtype: str
  :return: The name of the field and the value. If none of the fields are in the entity itself will be returned.
  """
  format = kwargs.get('format', '{field} "{ref}"')
  if type(entity) is not dict:
    return entity
  default_value = kwargs.get('default', entity)
  for field in fields:
    ref = entity.get(field, None)
    if ref is not None:
      return format.format(field=field, ref=ref)
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


def replace_ids(path: str, ids=None):
  """
  Reads all the files of all entities and reformat the file to be readable for humans.
  Also, it replaces the ids referenced in other entities by {{ fromName('{name}') }}.

  :param str path: A path to the entity files.
  :param str ids: A dict with ids as keys and names as values.
  :rtype: dict
  :return: The same ids dict but updated.
  """
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
                                               entity_type, entities, ids)
        if success:
          ids = new_ids
        file.seek(0)
        json.dump(entities, file, indent=2)
        file.truncate()
  return ids


# def replace_names_for_ids_of_products(project_path: str, products: list[str]):
#   raise_file_not_found_error(project_path)
#   names_ids = {}
#   if INGESTION in products and CORE not in products:
#     products.insert(0, CORE)
#   for product in products:
#     entity_types = PRODUCTS.get(product, None)
#     for entity_type in entity_types:
#       file_path = os.path.join()


def replace_names_by_ids(entity_type: PdpEntity, entities: list[dict], names: dict):
  for entity in entities:
    if 'id' not in entity:
      name, _id = associate_values_from_entities(entity, 'name', entity, 'id', entity_type, True)
      if name is not None:
        names[name] = _id
    replace_value(entity_type, entity, names, from_field='name', format='{0}')
  return names


def replace_ids_for_names(entity_type: PdpEntity, entities: list[dict], ids: dict):
  """
  Replaces all the ids in an entity referencing another entity for the name of the another
  entity with the format {{ fromName('{name}') }}. And also adds his id with his name to
  the ids dict.

  :param PdpEntity entity_type: A class representing the type of the entity.
  :param list[dict] entities: A list with all the entities to replaces de ids.
  :param dict ids: A dictionary containing the ids as keys and the names as values.
  :rtype: dict
  :return: The ids dict with his id and his name as a key-value.
  """
  for entity in entities:
    success, response = handle_and_continue(associate_values_from_entities, {'show_exception': True}, entity, 'id',
                                            entity, 'name', entity_type)
    if success:
      _id, name = response
      if _id is not None:
        ids[_id] = name
    replace_value(entity_type, entity, ids)
  return ids


def show_expected_error_if_value_not_found(entity_type: PdpEntity, entity: dict, from_field: str,
                                           value: str):
  name = identify_entity(entity, default=entity_type.associated_file_name)
  if name != entity_type.associated_file_name:
    print_warning(f'Value "{value}" does not exist while attempting to replace in {from_field}. '
                  f'Entity "{name}" in file {entity_type.associated_file_name}.')
  else:
    print_warning(f'Value "{value}" does not exist while attempting to replace in {from_field}. '
                  f'Entity has no name and no id in file {entity_type.associated_file_name}.')


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
    if entity not in values.keys():
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

  :param str api_url: The url where it will try to download the zip with the entities.
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


def get_entity_type_per_file_name(file_name: str, path_to_file: str = None):
  entity_type = None
  for entity_type in ENTITIES:
    if file_name.lower() == entity_type.associated_file_name.lower():
      return entity_type
  if path_to_file is not None:
    raise PdpException(message=f'File {file_name} on {path_to_file} is not recognized as one PDP products file.',
                       severity=WARNING_SEVERITY)
  raise PdpException(message=f'File {file_name} is not recognized as one PDP products file.', severity=WARNING_SEVERITY)


def clear_from_name_format(value: any):
  if type(value) is not str:
    return value
  search = re.search(r"\{\{\sfromName\('(.+?)'\)\s\}\}", value)
  if search is None or len(search.groups()) < 1:
    return value
  return search.group(1)


def order_products_to_deploy(products: list[str] = []):
  # Actually the only order that matters is that CORE must be before INGESTION
  if INGESTION not in products:
    return products

  if CORE not in products:
    return products

  # If INGESTION and CORE are in products, then we have to assure that CORE is before INGESTION
  core_index = products.index(CORE)
  ingestion_old_index = products.index(INGESTION)
  products.insert(core_index + 1, products.pop(ingestion_old_index))

  return products


def read_and_parse_to_deploy_entities_from(project_path: str, products: list[str] = []):
  names_with_dependencies = {}
  dependencies = []
  entities = {}

  if INGESTION in products and CORE not in products:
    dependencies += [CORE]

  def manage_entity_type_iteration(entity_type):
    """
      This function was made to handle_and_continue all the logic of the inner for
      (the for on the next function).
    """
    entity_file_path = os.path.join(project_path, product.title(), entity_type.associated_file_name)
    raise_file_not_found_error(entity_file_path)
    entities_temp = read_entities(entity_file_path)
    if product in products:
      entities[product][entity_type.type] = entities_temp

  def manage_product_iteration():
    """
      This function was made to handle_and_continue all the logic of the outter for
      (the for below this function).
    """
    entity_types = PRODUCTS[product]['entities']
    for entity_type in entity_types:
      handle_and_continue(manage_entity_type_iteration, {'show_exception': True}, entity_type)

  for product in order_products_to_deploy([*dependencies, *products]):
    if product in products:
      entities[product] = {}
    handle_and_continue(manage_product_iteration, {'show_exception': True})

  return entities


def create_or_update_entity(product_url: str, type: str, entity: dict):
  successful_message = '{type} {entity} has been {action} successfully with id {id}.'
  try:
    entity_id = entity.get('id', None)
    if entity_id is not None:
      old_entity = get(URL_GET_BY_ID.format(product_url, entity=type, id=entity_id))
      if old_entity is not None:
        res = put(URL_UPDATE.format(product_url, entity=type, id=entity_id), json=entity)
        styled_action = click.style('updated', fg='blue')
        styled_id = click.style(entity_id, fg='green')
        print_console(
          successful_message.format(type=type.title(), entity=identify_entity(entity), action=styled_action,
                                    id=styled_id)
        )
        return json.loads(res).get('id', None)
    res = post(URL_CREATE.format(product_url, entity=type), json=entity)
    entity_id = json.loads(res).get('id', None)
    styled_id = click.style(entity_id, fg='green')
    styled_action = click.style('created', fg='blue')
    print_console(
      successful_message.format(type=type.title(), entity=identify_entity(entity), action=styled_action, id=styled_id)
    )
    return entity_id
  except PdpException as exception:
    format = "\"{ref}\""
    print_error(
      f'Could not create entity {identify_entity(entity, default=type, format=format)} due to:\n'
      f'\t{exception.content.get("errors")}'
    )
