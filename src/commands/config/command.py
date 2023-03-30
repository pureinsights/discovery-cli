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

from commands.config.create import run as run_create
from commands.config.deploy import run as run_deploy
from commands.config.init import run as run_init
from commons.console import print_error
from commons.constants import ENTITIES, PRODUCTS, STAGING, TEMPLATES_DIRECTORY
from commons.custom_classes import PdpException
from commons.file_system import list_directories, list_files, replace_file_extension
from commons.pdp_products import get_entity_type_by_name


@click.group()
@click.pass_context
def config(ctx):
  """
  Contains all the commands to help you manage the entities of the PDP.
  You can create, update, delete, deploy and more.\n
  Use --help on each command for more detailed information.
  """


TEMPLATE_NAMES = [directory.lower() for directory in list_directories(os.path.join(TEMPLATES_DIRECTORY, 'projects'))]


@config.command()
@click.option('-n', '--project-name', default='my-pdp-project',
              help='The name of the resulting directory, will try to fetch existing configurations from the APIs '
                   'referenced in ~/.pdp. Notice that imported configs have id fields, don`t change those. Default is '
                   'my-pdp-project.')
@click.option('--empty/--no-empty', default=True, help='If it should only create an empty directory structure with '
                                                       'basic handlebars for starting a new project. Default is True.')
@click.option('-u', '--product-url', 'products_url', multiple=True, default=[], type=(str, str),
              help='The base URL for the given product API. The '
                   'product URL must be provided with the following '
                   'format PRODUCT_NAME:URL. The command allows '
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
  Creates a new project from existing sources or from scratch. It will create the folder structure for a PDP project.
  """
  config = ctx.obj['configuration']
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
@click.option('--target', 'targets', default=[product for product in PRODUCTS['list'] if product != STAGING],
              multiple=True,
              type=click.Choice([product for product in PRODUCTS['list'] if product != STAGING], case_sensitive=False),
              help='The name of the product where you want to deploy the entities.  The command allows multiple flags '
                   'to define multiple targets. Default are [ingestion, core, discovery]')
@click.option('-v', '--verbose', 'is_verbose', is_flag=True, default=False,
              help='It will show more information about the deployment results. Default is False.')
@click.option('-g', '--ignore-ids/--no-ignore-ids', 'ignore_ids', default=False,
              help='Will cause existing ids to be ignored, hence everything will be created as a new instance. This '
                   'is useful when moving configs from one instance to another. Default is False.')
@click.option('-q', '--quiet', is_flag=True, default=False,
              help='Display only the seed ids. Warnings and Errors will not be shown neither. Default is False.')
@click.pass_context
def deploy(ctx, targets: list[str], is_verbose: bool, ignore_ids: bool, quiet: bool):
  """
  Deploys project configurations to the target products.
  Must be run within the directory from a project created with the 'init' command.
  Will replace any name reference with IDs. Names are case-sensitive. If the "id" field is missing from an entity,
  assumes this is a new instance.
  """
  path = ctx.obj['project_path']
  config = ctx.obj['configuration']
  run_deploy(config, path, targets, is_verbose and not quiet, ignore_ids, quiet)


@config.command()
@click.option('-t', '--entity-type', 'entity_type_name', required=True,
              type=click.Choice([
                entity_type.type if entity_type.type != 'processor'
                else f'{entity_type.product}{entity_type.type.title()}'
                for entity_type in ENTITIES],
                case_sensitive=False
              ),
              help='This is the type of the entity that should be created. The entity type supported at the moment are:'
                   ' seed, ingestionProcessor, pipeline, scheduler, discoveryProcessor and endpoint.')
@click.option('--entity-template', default=None, help='This is the name of the template of the entity to use.')
@click.option('--deploy', 'has_to_deploy', default=False,
              help='It will deploy the entity configuration to the corresponding product. Default is False.')
@click.option('--file',
              help='The path to the file that contains the configuration for the entity or entities. If the '
                   'configuration contains an id property it will be updated instead. Default is the established '
                   'configuration for each entity.')
@click.option('-j', '--json', is_flag=True,
              help='This is a boolean flag. It will print the results in JSON format. Default is False.')
@click.pass_context
def create(ctx, entity_type_name: str, entity_template: str, file: str, has_to_deploy: bool, json: bool):
  """
  Add a new entity configuration to the entities on the current project. The configuration for each entity it will have
  default values. You can change those values and deploy them later.
  """
  config = ctx.obj['configuration']
  project_path = ctx.obj['project_path']
  entity_type = get_entity_type_by_name(entity_type_name)
  if file is None:
    if entity_template is None:
      raise PdpException(message="Entity template not provided. You must provide at least one flag to get the entity "
                                 "properties. Allowed flags: --entity-template, --file")

    entity_type_templates_path = os.path.join(TEMPLATES_DIRECTORY, 'entities', entity_type.product, entity_type.type)
    entity_templates = [replace_file_extension(file_name, '') for file_name in list_files(entity_type_templates_path)]

    if entity_template not in entity_templates:
      raise PdpException(message=f'Entity template "{entity_template}" not supported. Please provide one of the '
                                 f'following templates: {",".join(entity_templates)}.')

    file = os.path.join(entity_type_templates_path, replace_file_extension(entity_template, '.json'))
  run_create(config, project_path, entity_type, file, has_to_deploy, json)
