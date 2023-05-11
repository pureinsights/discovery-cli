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

import configparser
import os
import shutil

import click

from commons.console import create_spinner, print_exception, spinner_change_text, spinner_fail, spinner_ok
from commons.constants import CORE, FILES_FOLDER, PRODUCTS, STAGING, TEMPLATES_DIRECTORY
from commons.custom_classes import PdpException
from commons.handlers import handle_and_continue, handle_and_exit
from commons.pdp_products import export_entities


def run(project_name: str, apis: dict, force: bool, template: str = None):
  """
  This is the entry point for the command 'init'. Creates a project from
  scratch or from existing sources based on empty.

  :param str project_name: This will be the name of the folder that will contain the structure of the project.
  :param dict apis: A dictionary containing the url for each product api. (ingestion, discovery, core and staging).
  :param bool force: If it is True will try to override the project where you want to create it, if there is a
                      folder with the same name.
  :param str template: The name of the template to use, if None then will create a non-empty project.
  :rtype: bool
  :return: True if the project was created successfully, False if any error happen.
  :raises Exception: If a project with the same name already exists.
  """
  if force and os.path.exists(project_name):
    handler_params = {'message': f'Can not remove {project_name.title()}.'}
    handle_and_exit(shutil.rmtree, handler_params, project_name)

  created_successfully = False
  if template is not None:
    template_path = os.path.join(TEMPLATES_DIRECTORY, 'projects', template)
    created_successfully = create_project_from_template(project_name, template_path)
  else:
    created_successfully = create_project_from_existing_sources(project_name, apis)

  if not created_successfully:
    return False

  # Creates the pdp.ini configuration
  project_configuration = configparser.RawConfigParser()
  project_configuration['DEFAULT'] = apis

  with open(f'{project_name}/pdp.ini', 'w') as file:
    project_configuration.write(file)

  return True


def create_project_from_template(project_name: str,
                                 relative_template_path: str = os.path.join(TEMPLATES_DIRECTORY, 'projects',
                                                                            'random-generator')):
  """
  Copies the folder structure of the given template on a new directory called as the given project_name
  parameter.

  :param str project_name: The name of the folder.
  :param str relative_template_path: The path to the template that will be used.
  :rtype: bool
  :return: True if the project was created, False if some error happen.
  """
  try:
    if os.path.exists(project_name):
      raise PdpException(message=f'Project {project_name} already exists.\n\tUse --force flag to override the project.',
                         handled=True)
    # Copy the sample files from the templates folder to the project folder
    templates_path = relative_template_path
    shutil.copytree(templates_path, project_name)
    return True
  except Exception as error:
    print_exception(error)
    return False


def create_project_folder_structure(project_path: str,
                                    folders: list[str] = None):
  """
  Create empty folders. The default list of folders are ['Discovery', 'Core', 'Core/Files', 'Ingestion']
  but you can pass a custom folder structure.

  :param str project_path: The path where the project will be created.
  :param list[str] folders: A list of relative paths to the project_path.
  """
  if folders is None:
    folders = [product.title() for product in PRODUCTS['list'] if product != STAGING] + \
              [os.path.join(CORE.title(), FILES_FOLDER)]
  if not os.path.exists(project_path):
    os.mkdir(project_path)
    for folder in folders:
      os.mkdir(os.path.join(project_path, folder))
  else:
    splitted_path = os.path.split(project_path)
    raise PdpException(message=f'Project {splitted_path[len(splitted_path) - 1]} already exists.'
                               '\n\tUse --force flag to override the project.', handled=True)


def create_project_from_existing_sources(project_name: str, apis: dict):
  """
  Creates the project folder and imports all the entities of all products.

  :param str project_name: The name of the folder.
  :param dict apis: A dictionary containing the url for each product api. (ingestion, discovery, core and staging).
  :rtype: bool
  :return: True if the project was created, False if some error happen.
  """
  try:
    project_path = os.path.join('.', project_name)
    create_project_folder_structure(project_path)

    # Exports from all products (excepting STAGING)
    products = [product for product in PRODUCTS['list'] if product != STAGING]
    ids = {}
    successful_import = False
    for product in products:
      create_spinner()
      product_dir_name = product.title()
      spinner_change_text(f'Importing {product_dir_name} entities...')
      zip_path = os.path.join(project_path, product_dir_name)
      product_api_url = apis.get(product)
      success, new_ids = handle_and_continue(export_entities,
                                             {'show_exception': True},
                                             product_api_url, zip_path, True, ids=ids)
      if new_ids is not None:
        ids = new_ids
      if not success:
        spinner_fail(click.style(f'Could not import the {product_dir_name} entities.', fg='red'), suffix='\n')
      else:
        spinner_ok(click.style(f'{product_dir_name} entities imported.', fg='green'))
        successful_import = True

    return successful_import
  except Exception as exception:
    print_exception(exception)
    return False
