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
import os

import click

from commands.config.deploy import run as run_deploy
from commands.config.init import run as run_init
from commons.console import print_error
from commons.constants import PRODUCTS, STAGING
from commons.file_system import list_directories


@click.group()
@click.pass_context
def config(ctx):
  """
  Contains all the commands to help you manage the entities on PDP.
  You can create, update, delete, deploy and more.\n
  Use --help on each command for more detailed information.
  """


TEMPLATE_NAMES = [directory.lower() for directory in list_directories(
  os.path.join(os.path.dirname(__file__), 'templates', 'projects'))]


@config.command()
@click.option('-n', '--project-name', default='my-pdp-project',
              help='The name of the resulting directory, will try to fetch existing configurations from the APIs '
                   'referenced in ~/.pdp. Notice that imported configurations have id fields, don`t change those. '
                   'Default is my-pdp-project.')
@click.option('--empty/--no-empty', default=True, help='If it should only create an empty directory structure with '
                                                       'basic handlebars for starting a new project. Default is True.')
@click.option('-u', '--product-url', 'products_url', multiple=True, default=[], type=(str, str),
              help='The base URL for the given product API. The '
                   'product URL must be provided with the following '
                   'format PRODUCT_NAME URL. The command allows '
                   'multiple flags to define multiples products.\n '
                   'Default are ingestion http://localhost:8080,'
                   'staging http://localhost:8081,'
                   'core http://localhost:8082,'
                   'discovery http://localhost:8088.')
@click.option('--force/--no-force', default=False,
              help='If there is a project with the same name it will to override it. '
                   'Default is False.')
@click.option('--template', default=None, help='Choose the template with the project will be created.',
              type=click.Choice(TEMPLATE_NAMES,
                                case_sensitive=False))
@click.pass_context
def init(ctx, project_name: str, empty: bool, products_url: list[(str, str)], force: bool, template):
  """
  Creates a new project from existing sources or from scratch. Will create the folder structure for a PDP project.
  """
  config = ctx.obj['configuration']
  if config.get('load_config'):
    from pdp import load_config
    path = os.path.join('.', project_name, 'pdp.ini')
    config = load_config(path, ctx.obj['profile'])

  for product_url in products_url:
    product: str
    url: str
    product, url = product_url
    if config.get(product.lower(), None) is None:
      print_error(f'Unrecognized product "{product}".', True)
    else:
      config[product.lower()] = url

  if empty and template is None:
    template = 'random_generator'
  elif not empty:
    template = None

  successfully_executed = run_init(project_name, config, force, template)
  color = 'green'
  message = 'Project {project_name_styled} created successfully.\n' \
            'Recommended next commands:\n' \
            '\tcd {project_name}\n' \
            '\tpdp config deploy'
  if not successfully_executed:
    color = 'red'
    message = 'Could not create the project {project_name_styled}.'
  project_name_styled = click.style(project_name, fg=color)
  click.echo(message.format(project_name=project_name, project_name_styled=project_name_styled))
  exit(0 if successfully_executed else 1)


@config.command()
@click.option('-d', '--dir', 'path', default='.', help='The path to a directory with the structure and the pdp.ini '
                                                       'that init command creates. Default is ./.')
@click.option('--target', 'targets', default=[product for product in PRODUCTS['list'] if product != STAGING],
              multiple=True,
              type=click.Choice([product for product in PRODUCTS['list'] if product != STAGING], case_sensitive=False),
              help='The name of the product where you want to deploy the entities.  The command allows multiple flags '
                   'to define multiple targets. Default are [ingestion, core, discovery]')
@click.option('-v', '--verbose', 'is_verbose', is_flag=True, default=False,
              help='Will show more information about the deployment results. Default is False.')
@click.option('-g', '--ignore-ids/--no-ignore-ids', 'ignore_ids', default=False,
              help='Will cause existing ids to be ignored, hence everything will be created as a new instance. This '
                   'is useful when moving configs from one instance to another. Default is False.')
@click.option('-q', '--quiet', is_flag=True, default=False,
              help='Display only the seed ids. Warnings and Errors will not be shown. Default is False.')
@click.pass_context
def deploy(ctx, targets: list[str], path: str, is_verbose: bool, ignore_ids: bool, quiet: bool):
  """
  Deploys project configurations to the target products.
  Must be run within the directory from a project created with the 'init' command.
  Will replace any name reference with IDs. Names are case-sensitive. If the "id" field is missing from an entity,
  assumes this is a new instance.
  """
  config = ctx.obj['configuration']
  if config.get('load_config'):
    from pdp import load_config
    config_path = os.path.join(path, 'pdp.ini')
    config = load_config(config_path, ctx.obj['profile'])
  run_deploy(config, path, targets, is_verbose and not quiet, ignore_ids, quiet)
