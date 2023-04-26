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
#
#  Permission to use, copy, modify or distribute this software and its
#  documentation for any purpose is subject to a licensing agreement with
#  Pureinsights Technology Ltd.
#
#  All information contained within this file is the property of
#  Pureinsights Technology Ltd. The distribution or reproduction of this
#  file or any information contained within is strictly forbidden unless
#  prior written permission has been granted by Pureinsights Technology Ltd.
import errno
import os
from typing import Callable

from commons.constants import PRODUCTS, STAGING
from commons.custom_classes import DataInconsistency, PdpEntity


def raise_file_not_found_error(path: str):
  """
  Raises a FileNotFoundError if the given path does not exist.
  :param str path: The path to validate if exists.
  :raises FileNotFoundError: If path doesn't exist.
  """
  if not os.path.exists(path):
    err_no = errno.ENOENT
    raise FileNotFoundError(err_no, os.strerror(err_no), path)


def unique_fields(**kwargs):
  """
  Validates duplicated values among the entities.
  :key dict entity: The entity to validate.
  :key dict aux: A dictionary containing useful information to make the validation.
  """
  entity = kwargs.get('entity', None)
  aux = kwargs.get('aux', None)
  ignore_ids = aux.get('ignore_ids', False)
  _unique_fields = ['id', 'name'] if not ignore_ids else ['name']
  for field in _unique_fields:
    entity_field = entity.get(field, None)
    if entity_field is not None:
      repeated_field = aux.get(entity_field, None)
      if repeated_field is not None:
        raise DataInconsistency(
          message=f'Field "{field}" must be unique. More than one entity has the same {field}  "{entity_field}".',
          handled=False
        )
      aux[entity_field] = entity

  return True


def validate_pdp_entities(requirements: list[Callable], entities: dict, aux: dict = None):
  """
  Check if each given entity meets all the given requirements.
  :param list[Callable] requirements: The list of requirements who each entity must meet.
  :param dict entities: A dictionary containing all the entities to check.
  :param dict aux: A dictionary containing useful information to make the validation.
  :rtype: bool
  :return: True if all the entities meets all the requirements, False in other case
  :raise: Some requirements can raise exceptions.
  """
  if aux is None:
    aux = {}
  for entity_type in entities.keys():
    for entity in entities[entity_type]:
      for requirement in requirements:
        if not requirement(entity=entity, entities=entities, aux=aux, entity_type=entity_type):
          return False
  return True


def raise_for_pdp_data_inconsistencies(project_path: str, aux: dict = None):
  """
  Check if a PDP project meets all the requirements to be a valid PDP project.
  :param str project_path: The path to the project to check.
  :param dict aux: A dictionary containing useful information to check the project.
  :raises FileNotFoundError: If path to the project doesn't exist.
  :raises PdpException: If the project doesn't have a valid project structure.
  """
  from commons.pdp_products import order_products_to_deploy
  from commons.file_system import has_pdp_project_structure, read_entities
  raise_file_not_found_error(project_path)
  has_pdp_project_structure(project_path, show='error')
  entities = {}
  requirements = [unique_fields]
  products = [product for product in PRODUCTS['list'] if product != STAGING]
  for product in order_products_to_deploy(products):
    entity_types = PRODUCTS[product].get('entities', [])
    for entity_type in entity_types:
      file_path = os.path.join(project_path, product.title(), entity_type.associated_file_name)
      entities[entity_type] = read_entities(file_path)
  validate_pdp_entities(requirements, entities, aux)


def raise_for_inconsistent_product(entity_type: PdpEntity, product: str):
  if product is not None:
    if entity_type is not None and product != entity_type.product:
      raise DataInconsistency(
        message=f"The entity type \"{entity_type.user_facing_type_name()}\" doesn't belong to \"{product.title()}\".")
