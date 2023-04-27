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
from commons.constants import PRODUCTS, STAGING
from commons.custom_classes import PdpEntity
from commons.pdp_products import export_entities


def run(config: dict, product: str | None, entity_type: PdpEntity | None, entity_id: str | None,
        include_dependencies: bool):
  """
  Downloads a .zip from PDP APIs containing the configuration for the entities.
  :param dict config: A dictionary containing the API urls and the path where the command was called.
  :param str product: The name of the product where the zip will be downloaded. If None, then will download a zip per
                      product.
  :param PdpEntity entity_type: The entity type that will be exported from the API.
  :param str entity_id: The id of the entity that will be exported.
  :param bool include_dependencies: Will include in the .zip those entities related with the specified by the id.
  """
  project_path = config['project_path']
  if product is None and entity_type is None:
    products = [product for product in PRODUCTS['list'] if product != STAGING]
    for product in products:
      run(config, product, None, None, include_dependencies)
    return

  if entity_type is not None:
    export_entities(config[entity_type.product], project_path, False, entity_type=entity_type, entity_id=entity_id,
                    zip_name=f'{entity_type.type.title()}.zip', include_dependencies=include_dependencies, verbose=True)
    return

  export_entities(config[product], project_path, False, zip_name=f'{product.title()}.zip', verbose=True)
