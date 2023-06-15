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

from commons.console import create_spinner, get_number_errors_exceptions, print_console, print_warning, \
  spinner_change_text, spinner_ok, \
  suppress_errors, suppress_warnings, verbose
from commons.constants import CORE, INGESTION, PRODUCTS
from commons.custom_classes import PdpEntity
from commons.file_system import read_entities, write_entities
from commons.handlers import handle_and_continue
from commons.pdp_products import create_or_update_entity, identify_entity, replace_ids, \
  replace_ids_for_names, \
  replace_names_by_ids
from commons.raisers import raise_file_not_found_error


def inject_dependencies(products: list[str]):
  """
  Inject the product's dependencies if needed and not on the list.
  :param list[str] products: The list of products
  :rtype: list[str]
  :return: The list with dependencies included
  """
  if INGESTION in products:
    if CORE in products:
      products.remove(CORE)
    products.insert(products.index(INGESTION), CORE)
  return products


def deploy_entities(config: dict, product: str, entity_type: PdpEntity, entities: list[dict], names_ids: dict,
                    seeds: list[dict], is_target: bool, is_verbose: bool, ignore_ids: bool):
  """
  This is the last auxiliary function. It iterates for each entity of the given entity_type.
  And deploy the entity.
  """
  for entity in entities:
    product_url = config.get(product)
    # We store the id to recover it if we have to ignore the ids and the entity could not be deployed.
    id_backup = None
    if ignore_ids and is_target:
      id_backup = entity.pop('id', None)
    _, new_id = handle_and_continue(
      create_or_update_entity, {'show_exception': True}, product_url, entity_type.type,
      entity, verbose=is_verbose
    )
    # If _id is None means the entity was not created, so we use the backup.
    _id = id_backup if new_id is None else new_id
    name = entity.get('name', None)
    if _id is not None:
      entity['id'] = _id  # Updates the id of the entity, the product could create a new one
      if name is not None:
        names_ids[name] = _id

    # If the entity is a seed, we store the entity to show the ids later.
    if entity_type.type == 'seed' and new_id is not None:
      seeds += [entity]

  return seeds


def deploy_entity_types(config: dict, product: str, path: str, ids_names: dict, seeds: list[dict], is_target: bool,
                        is_verbose: bool, ignore_ids: bool):
  """
  This is an auxiliary function. It iterates for each entity_type of a product and calls another auxiliary function.
  """
  entity_types = PRODUCTS.get(product, {'entities': []}).get('entities')
  for index, entity_type in enumerate(entity_types):
    file_path = os.path.join(path, product.title(), entity_type.associated_file_name)

    if is_target and is_verbose:
      create_spinner()
      spinner_change_text(f'Deploying {entity_type.type}s to {product}...')

    success, entities = handle_and_continue(read_entities, {'show_exception': True}, file_path)
    if not success:
      continue

    replace_names_by_ids(entity_type, entities, ids_names)
    # If the product is not a target then we don't deploy the entities
    if not is_target:
      continue

    if len(entities) <= 0:
      print_warning(
        verbose(
          verbose_func='There are not entities to deploy.',
          verbose=is_verbose
        )
      )

    deploy_entities(config, product, entity_type, entities, ids_names, seeds, is_target, is_verbose, ignore_ids)

    # Once the entities are deployed, we replace the ids for the names to keep it simple to the user.
    handle_and_continue(
      replace_ids_for_names, {'message': 'Could not replace the ids by the entity names.'
                                         'You might want to doit your self', 'warning': True},
      entity_type, entities, ids_names, suppress_warnings=True
    )
    write_entities(file_path, entities)
    message, icon = verbose(
      verbose_func=(f'{click.style(entity_type.type.title() + "s", fg="cyan")}:', f'{index + 1})'),
      not_verbose_func=('', None),  # Do not print anything but finish the spinner
      verbose=is_verbose
    )
    if is_target and is_verbose:
      spinner_ok(message, icon=icon)


def deploy_products(config: dict, products: list[str], target_products: list[str], path: str, seeds: list[dict],
                    is_verbose: bool, ignore_ids: bool):
  """
  This is an auxiliary function. It iterates for each product and calls another auxiliary function.
  """
  new_line = ''
  ids_names = {}
  for index, product in enumerate(products):
    is_target = product in target_products

    if is_target:
      print_console(
        verbose(
          verbose_func=f'--------------------|{product.title()} entities|--------------------',
          verbose=is_verbose
        ),
        prefix=new_line
      )
      new_line = '\n'

    deploy_entity_types(config, product, path, ids_names, seeds, is_target, is_verbose, ignore_ids)
    handle_and_continue(
      replace_ids,
      {},
      path, ids_names
    )


def run(config: dict, path: str, target_products: list[str], is_verbose: bool = False, ignore_ids: bool = False,
        quiet: bool = False):
  """
  Deploy all the entities of the specified products with a specific behavior based on the given flags.

  :param dict config: The configuration that contains the product's url.
  :param str path: The path to the root of a pdp project.
  :param list[str] target_products: The list of the product names to be deployed.
  :param bool is_verbose: Will show more information while deploy the entities.
  :param bool ignore_ids: Will ignore the ids of the entities and try to create a new one.
  :param bool quiet: Will suppress all the warnings and errors. Only shows the ids of the seeds, separated by \n.
                     If and error occurs then no ids will be showed.
  """
  suppress_warnings(quiet)
  suppress_errors(quiet)

  raise_file_not_found_error(path)
  products = [product for product in PRODUCTS if product in target_products]
  seeds = []

  # Ensures that CORE is before than INGESTION, if ingestion is in target_product.
  # Since CORE is a dependency for INGESTION.
  products = inject_dependencies(products)

  deploy_products(config, products, target_products, path, seeds, is_verbose, ignore_ids)

  # Will show the seed ids only if, seeds is not empty,
  # is verbose or not verbose or if is quiet but no errors happened.
  if len(seeds) > 0 and (not quiet or (quiet and get_number_errors_exceptions() <= 0)):
    print_format = 'Seed {name} with id {id}.'
    if not quiet:
      print_console(f'{click.style("Seeds", fg="cyan")} ids:', prefix='\n' if is_verbose else '')
    for seed in seeds:
      _id = seed.get("id")
      if not quiet:
        _id = click.style(seed.get("id"), fg="green")
      else:
        print_format = '{id}'
      print_console(
        print_format.format(name=click.style(identify_entity(seed), fg="blue"), id=_id)
      )
