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

from commands.config.get import print_stage as print_entities
from commons.console import create_spinner, print_console, spinner_change_text, spinner_fail, spinner_ok
from commons.constants import CORE, URL_SEARCH
from commons.http_requests import post


def print_stage(entities: list[dict], is_json: bool, pretty: bool):
  """
  Print the entities as JSON or as table.
  :param list[dict] entities: A list of entities to print.
  :param bool is_json: Will print the entities in JSON format.
  :param bool pretty: If True, the result will be showed in  a human-readable JSON format.
  """
  if len(entities) <= 0:
    return print_console("No entities were found...")
  product_entities = {}
  for entity in entities:
    product = entity.get('product', 'core')
    entity_type = entity.get('type', 'other')
    product_entities[product] = product_entities.get(product, {})
    product_entities[product][entity_type] = product_entities[product].get(entity_type, []) + [entity]

  print_entities(product_entities, not is_json, is_json, pretty)


def search_entity(core_url: str, query_params: dict):
  """
  Calls to the CORE API search endpoint and returns the result.
  :param str core_url: The url to the Core API.
  :param dict query_params: A dictionary containing the query params of the request.
  """
  res = post(URL_SEARCH.format(core_url), params=query_params)
  if res is None:
    return None

  return json.loads(res).get('content', [])


def run(config: dict, query_params: dict, is_json: bool, pretty: bool):
  """
  Will search entities based on the query_params given.
  :param dict config: A dictionary containing the product's url.
  :param dict query_params: The criteria for search the entities.
  :param bool is_json: Will print the entities in JSON format.
  :param bool pretty: If True, the result will be showed in  a human-readable JSON format.
  """
  create_spinner()
  spinner_change_text("Searching for entities...")
  entities = search_entity(config[CORE], query_params)
  if entities is None:
    return spinner_fail("No entities match the given criteria.")

  if not is_json:
    spinner_ok('Some entities found...')
  print_stage(entities, is_json, pretty)
