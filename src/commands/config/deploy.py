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
import os.path

import click

from commons.console import create_spinner, spinner_change_text, spinner_ok
from commons.constants import CORE, INGESTION, PRODUCTS
from commons.file_system import read_entities, write_entities
from commons.handlers import handle_and_continue
from commons.pdp_products import create_or_update_entity, replace_names_by_ids
from commons.raisers import raise_file_not_found_error


def run(config: dict, path: str, target_products):
  raise_file_not_found_error(path)
  products = [product for product in PRODUCTS if product in target_products]
  if INGESTION in products:
    products.insert(products.index(INGESTION), CORE)
  names_ids = {}
  for product in products:
    create_spinner()
    spinner_change_text(f'Deploying {product} entities...')
    entity_types = PRODUCTS.get(product, []).get('entities')
    for entity_type in entity_types:
      file_path = os.path.join(path, product.title(), entity_type.associated_file_name)
      success, entities = handle_and_continue(read_entities, {'show_exception': True}, file_path)
      if not success:
        continue

      replace_names_by_ids(entity_type, entities, names_ids)
      if product not in target_products or len(entities) <= 0:
        continue

      for entity in entities:
        product_url = config.get(product)
        _, _id = handle_and_continue(create_or_update_entity, {'show_exception': True},
                                     product_url, entity_type.type, entity)
        name = entity.get('name', None)
        if _id is not None and name is not None:
          entity['id'] = _id
          names_ids[name] = _id
      write_entities(file_path, entities)
      spinner_ok(f'{product.title()} {click.style(entity_type.type + "s", fg="cyan")}:', icon='->')

  # entities = read_and_parse_to_deploy_entities_from(path, target_products)
  # entities_to_write = []
  # names_ids = {}
  # for product in entities.keys():
  #   entity_types = ENTITIES.get(product, [])
  #   if len(entities[product]) <= 0:
  #     continue
  #   print(f'{product.title()}: ')
  #   for type in entities[product].keys():
  #     entity_type = [entity_type for entity_type in entity_types if entity_type.type == type]
  #     entity_types.remove(entity_type)  # Just to improve performance
  #     entities_per_product = entities[product][type]
  #     for entity in entities_per_product:
  #       success, _id = handle_and_continue(create_or_update_entity, {'show_exception': True}, ctx.get(product), type,
  #                                          entity)
  #       if success:
  #         entity['id'] = _id
  #         entities_to_write += [entity]
  #     # TODO: Here you must write every single type of file to update any possible change of ID
  #     # But think about this more, because you already have all the entities in memory so looks like
  #     # you have to find another way to update the ids, probably just by calling the method
  #
  # print('finished')
