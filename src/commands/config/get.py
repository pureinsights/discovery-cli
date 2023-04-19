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
from builtins import enumerate

import click
from tabulate import tabulate

from commons.console import create_spinner, print_console, print_error, spinner_change_text, spinner_fail, spinner_ok
from commons.constants import DISCOVERY, PRODUCTS, STAGING, URL_GET_ALL, URL_GET_BY_ID
from commons.custom_classes import PdpEntity
from commons.handlers import handle_and_continue
from commons.http_requests import get
from commons.pdp_products import identify_entity
from commons.uuid_commons import is_hex_uuid, is_valid_uuid


def get_all_entities(config: dict, entity_types: list[PdpEntity], query_params: dict,
                     is_verbose: bool) -> dict:
  """
  Fetch and return all the entities, based on the given entity types, from the PDP products.
  :param dict config: The configuration containing the product's url.
  :param list[PdpEntity] entity_types: A list of entity type to filter the results.
  :param bool is_verbose: Will show useful information to the user.
  :rtype: dict
  :return: A dict containing as keys the name of the entity type and as vale a list with the entities.
  """
  entities = {}
  for entity_type in entity_types:
    if is_verbose:
      create_spinner()
      spinner_change_text(
        f"Searching for {click.style(entity_type.user_facing_type_name(), fg='cyan')} on "
        f"{click.style(entity_type.product.title(), fg='blue')}..."
      )
    product = entity_type.product
    entities[product] = entities.get(product, {})
    entities_found = get(URL_GET_ALL.format(config[product], entity=entity_type.type), params=query_params)

    if entities_found is None:
      if is_verbose:
        spinner_fail(
          f"No {click.style(entity_type.user_facing_type_name(), fg='cyan')} entities found "
          f"on {click.style(entity_type.product.title(), fg='blue')}.",
          icon=click.style("0", fg='red')
        )
      continue
    entities_found = json.loads(entities_found).get('content', [])
    entities[product][entity_type.type] = entities[product].get(entity_type.type, []) + entities_found
    if is_verbose:
      spinner_ok(
        f"{click.style(entity_type.user_facing_type_name(), fg='cyan')} entities found "
        f"on {click.style(entity_type.product.title(), fg='blue')}.",
        icon=f'{click.style(str(len(entities_found)), fg="green")}'
      )
  return entities


def get_entities_by_ids(config: dict, ids: list[str], entity_types: list[PdpEntity], query_params: dict,
                        is_verbose: bool = False) -> dict:
  """
  Fetch and returns all the entities filtered by the given list of ids.
  :param dict config: The configuration containing the product's url.
  :param list[str] ids: A list with the ids of the entities to fetch.
  :param list[PdpEntity] entity_types: A list of entity types to filter
  :param dict query_params: A dictionary containing the query params to be sent.
  :param bool is_verbose: Will show useful information to the user.
  :rtype: dict
  :return: A dict containing as keys the name of the entity type and as vale a list with the entities.
  """
  entities = {}
  for entity_id in ids:
    entity_found_in = None
    if not is_valid_uuid(entity_id):
      print_error(f"The id {entity_id} is not a valid uuid.", False)
      continue
    styled_id = click.style(entity_id, fg='blue')
    if is_verbose:
      create_spinner()
      spinner_change_text(f"Searching for entity {styled_id}")
    for entity_type in entity_types:
      product = entity_type.product
      # Ids with hex format are only supported by Discovery API
      if product != DISCOVERY and is_hex_uuid(entity_id):
        continue

      entities[product] = entities.get(product, {})
      product_url = config[entity_type.product]
      _, res = handle_and_continue(get, {},
                                   URL_GET_BY_ID.format(product_url, entity=entity_type.type, id=entity_id),
                                   params=query_params, status_404_as_error=False)
      if res is None:
        continue  # The entity doesn't exist
      entity = json.loads(res)
      type_name = entity_type.type
      entities[product][type_name] = entities[product].get(type_name, []) + [entity]
      entity_found_in = entity_type
      break
    if is_verbose and entity_found_in is not None:
      spinner_ok(f"Entity {styled_id} found in {click.style(entity_found_in.user_facing_type_name(), fg='cyan')}.")
    elif is_verbose:
      spinner_fail(
        f"Entity {styled_id} not found."
      )

  return entities


def print_stage(entities: dict, is_verbose: bool, is_json: bool):
  """
  Prints all the entities given, depending on the flags.
  :param dict entities: A dict containing the products, entity types and inside a list with entities.
  :param bool is_json: Print the entities just with JSON format.
  """
  if is_json:
    print_console(entities)
    return

  length = 0
  indentation = 0
  print_console(f"{click.style('Entities', fg='green')}")
  for product in entities.keys():
    if len(entities[product].keys()) <= 0:
      continue

    length_product = 0
    print_console(
      f"--------------------|{click.style(product.title(), fg='blue')}|--------------------",
    )
    for index, entity_type in enumerate(entities[product].keys()):
      if len(entities[product][entity_type]) <= 0:
        continue

      print_console(f"{click.style(entity_type.title(), fg='cyan')}:", prefix='  ' * (indentation + 2))
      if is_verbose:
        print_as_table(entities[product][entity_type], indentation + 3)
      else:
        print_entities(entities[product][entity_type], indentation + 3)

      length_product += len(entities[product][entity_type])
      print_console(
        f'{click.style("Entities found:", fg="cyan")} {len(entities[product][entity_type])}',
        prefix='  ' * (indentation + 2), suffix='\n' if index < len(entities[product].keys()) - 1 else ''
      )

    length += length_product
    print_console(
      f'--------------------|{click.style("Entities found:", fg="blue")} {length_product}|--------------------',
    )

  print_console(
    f'{click.style("Entities found:", fg="green")} {length}',
    prefix='  ' * indentation
  )


def print_entities(entities: list[dict], indentation: int = 0):
  """
  Shows just the id of the entity.
  :param list[dict] entities: The list of entities to show.
  :param int indentation: A number indicating how many indentation should be.
  """
  for entity in entities:
    print_console(
      "Entity {id} found it!".format(
        id=identify_entity(entity)
      ), prefix='  ' * indentation
    )


def print_as_table(entities: list[dict], indentation: int):
  """
  Shows the entities in a table with a few properties.
  :param list[dict] entities: The list of entities to show.
  :param int indentation: A number indicating how many indentation should be.
  """
  headers = ['id', 'name']
  other_columns = ['type', 'active', 'description']
  for key in entities[0].keys():
    if key in other_columns:
      headers += [key]
  table_values = [headers]
  for entity in entities:
    entity_values = []
    for header in headers:
      none_dict = {'empty': True}  # Represent None, we will use its reference to check if the entity has a value or not
      entity_value = entity.get(header, none_dict)  # The entity can contain a None as value, so we can't return None
      if entity_value is not none_dict:  # We compare by reference
        if entity_value is None:
          entity_value = 'None'
        entity_values += [entity_value]
    table_values += [entity_values]

  table_str = tabulate(table_values, headers='firstrow', showindex='always', tablefmt='presto')
  table_indented = ""
  lines = table_str.split('\n')
  for index, line in enumerate(lines):
    table_indented += f'{" " * indentation}{line}'
    if index < len(lines) - 1:
      table_indented += '\n'
  print_console(table_indented)


def filter_entities(filters: list[(str, any)], entities: dict) -> dict:
  """
  Returns all the entities who match the key-value list of filters.
  :param list[(str, any)] filters: A list containing the property and the expected value to each property.
  :param dict entities: A dictionary containing a list of entities to filter, categorized by product and entity type.
  """
  if len(filters) <= 0:
    return entities

  filtered_entities = {}
  for product in entities.keys():
    filtered_entities[product] = {}
    for entity_type in entities[product].keys():
      filtered_entities[product][entity_type] = []
      for entity in entities[product][entity_type]:
        for _key, value in filters:
          entity_value = str(entity.get(_key, None))
          if entity_value == value:
            filtered_entities[product][entity_type] += [entity]
  return filtered_entities


def run(config: dict, products: list[str], entity_type: PdpEntity | None, entity_ids: list[str],
        filters: list[(str, str)], query_param: dict, is_json: bool, is_verbose: bool):
  """
  Fetch information of entities and printed to console.
  :param dict config: The configuration containing the pdp products' url.
  :param list[str] products: A list of products to limit the search to entity types of those products.
  :param PdpEntity entity_type: An entity type to limit the search to entities of the same type.
                                If None searches for all entity types.
  :param list[str] entity_ids: A list of ids to search just for those ids. If empty search for all entities.
  :param list[(str, str)] filters: A list with key-value, the keys are properties and the value are the expected value
                                   of the property. Just entities who match those properties and
                                   values will be returned.
  :param dict query_param: Are the query param accepted by the "get all entities" endpoint of each product.
  :param bool is_json: Will show the entities in a JSON format.
  :param bool is_verbose: Will show more information to the user.
  """
  entities = {}
  entity_types = []
  if entity_type is not None:
    entity_types = [entity_type]
  else:
    if DISCOVERY in products:
      products.remove(DISCOVERY)
      products.insert(0, DISCOVERY)

    for product in products:
      if product != STAGING:
        entity_types += PRODUCTS[product]['entities']

  if len(entity_ids) > 0:
    entities = get_entities_by_ids(config, entity_ids, entity_types, query_param, is_verbose)
  else:
    entities = get_all_entities(config, entity_types, query_param, is_verbose)

  entities = filter_entities(filters, entities)
  print_stage(entities, is_verbose, is_json)
