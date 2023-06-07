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
import sys

import click

from commands.config._import import run as run_import
from commands.config.create import run as run_create
from commands.config.delete import run as run_delete
from commands.config.deploy import run as run_deploy
from commands.config.export import run as run_export
from commands.config.get import run as run_get
from commands.config.init import run as run_init
from commons.console import print_error
from commons.constants import ENTITIES, PRODUCTS, STAGING
from commons.custom_classes import DataInconsistency, PdpException
from commons.file_system import get_templates_directory, list_directories, list_files, replace_file_extension
from commons.pdp_products import get_entity_type_by_name, order_products_to_deploy
from commons.raisers import raise_for_inconsistent_product, raise_for_pdp_data_inconsistencies


@click.group()
@click.pass_context
def config(ctx):
  """
  Contains all the commands to help you manage the entities on PDP.
  You can create, update, delete, deploy and more.\n
  Use --help on each command for more detailed information.
  """


TEMPLATE_NAMES = [directory.lower() for directory in
                  list_directories(os.path.join(get_templates_directory(), 'projects'))]


@config.command()
@click.option('-n', '--name', default='my-pdp-project',
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
def init(ctx, name: str, empty: bool, products_url: list[(str, str)], force: bool, template):
  """
  Creates a new project from existing sources or from scratch. Will create the folder structure for a PDP project.
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
    template = 'random-generator'
  elif not empty:
    template = None

  successfully_executed = run_init(name, config, force, template)
  color = 'green'
  message = 'Project {project_name_styled} created successfully.\n' \
            'Recommended next commands:\n' \
            '\tcd {project_name}\n' \
            '\tpdp config deploy'
  if not successfully_executed:
    color = 'red'
    message = 'Could not create the project {project_name_styled}.'
  name_styled = click.style(name, fg=color)
  click.echo(message.format(project_name=name, project_name_styled=name_styled))
  sys.exit(0 if successfully_executed else 1)


@config.command()
@click.option('--product', 'products', default=[product for product in PRODUCTS['list'] if product != STAGING],
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
def deploy(ctx, products: list[str], is_verbose: bool, ignore_ids: bool, quiet: bool):
  """
  Deploys project configurations to the target products.
  Must be run within the directory from a project created with the 'init' command.
  Will replace any name reference with IDs. Names are case-sensitive. If the "id" field is missing from an entity,
  assumes this is a new instance.
  """
  path = ctx.obj['project_path']
  config = ctx.obj['configuration']
  raise_for_pdp_data_inconsistencies(path, {"ignore_ids": ignore_ids})
  run_deploy(config, path, products, is_verbose and not quiet, ignore_ids, quiet)


@config.command()
@click.option('-t', '--entity-type', 'entity_type_name', required=True,
              type=click.Choice([
                entity_type.type if entity_type.type != 'processor'
                else f'{entity_type.product}{entity_type.type.title()}'
                for entity_type in ENTITIES],
                case_sensitive=False
              ),
              help='This is the type of the entity that will be created. The entity types supported at'
                   'the moment are: seed, ingestionProcessor, pipeline, Scheduler, Endpoint, discoveryProcessor.')
@click.option('--entity-template', default=None,
              help="This is the template's name of the entity to use. Default is None.")
@click.option('--deploy', 'has_to_deploy', default=False, is_flag=True,
              help='It will deploy the entity configuration to the corresponding product. Default is False.')
@click.option('--path', '_file', default=None,
              help='The path to the file that contains the configuration for the entity or entities. If the '
                   'configuration contains an id property it will be updated instead. Default is the established '
                   'configuration for each entity.')
@click.option('-j', '--json', is_flag=True, default=False,
              help='This is a Boolean flag. Will print the results in JSON format. Default is False.')
@click.option('--interactive', is_flag=True, default=False,
              help='This is a Boolean flag. Will launch your default text editor to allow you to modify the entity '
                   'configuration. Default is False.')
@click.option('-g', '--ignore-ids/--no-ignore-ids', 'ignore_ids', default=False,
              help='Will cause existing ids to be ignored, hence everything will be created as a new instance. This '
                   'is useful when moving configs from one instance to another. Default is False.')
@click.pass_context
def create(ctx, entity_type_name: str, entity_template: str, _file: str, has_to_deploy: bool, json: bool,
           ignore_ids: bool, interactive: bool):
  """
  Add a new entity configuration to the entities on the current project. The configuration for each entity it will have
  default values depending on the template name provided, or you can specify your own entity configuration with the
  --file and/or --interactive flags. You can also deploy the entities to their respective product.
  """
  _config = ctx.obj['configuration']
  project_path = ctx.obj['project_path']
  entity_type = get_entity_type_by_name(entity_type_name)
  if _file is None:
    if entity_template is not None:
      entity_type_templates_path = os.path.join(get_templates_directory(), 'entities', entity_type.product,
                                                entity_type.type)
      entity_templates = [replace_file_extension(file_name, '') for file_name in list_files(entity_type_templates_path)]

      if entity_template not in entity_templates:
        raise PdpException(message=f'Entity template "{entity_template}" not supported. Please provide one of the '
                                   f'following templates: {",".join(entity_templates)}.')
      _file = os.path.join(entity_type_templates_path, replace_file_extension(entity_template, '.json'))

    if not interactive and entity_template is None:
      raise PdpException(message="Entity template not provided. You must provide at least one flag to get the entity "
                                 "properties. Allowed flags: --entity-template, --file")

  run_create(_config, project_path, entity_type, _file, has_to_deploy, json, ignore_ids, interactive)


@config.command()
@click.pass_obj
@click.option('--product', default=None,
              type=click.Choice(PRODUCTS['list'],
                                case_sensitive=False
                                ),
              help="Will filter the entities based on the name entered (Ingestion, Core or Discovery). Default is All.")
@click.option('-t', '--entity-type', 'entity_type_name', default=None,
              type=click.Choice([
                entity_type.type if entity_type.type != 'processor'
                else f'{entity_type.product}{entity_type.type.title()}'
                for entity_type in ENTITIES],
                case_sensitive=False
              ),
              help="Will filter and only show the entities of the type entered. Default is All.")
@click.option('-i', '--entity-id', 'entity_ids', default=[], multiple=True,
              help="Will only retrieve information for the component specified by the ID. Default is None. "
                   "The command allows multiple flags of -i.")
@click.option('-j', '--json', 'is_json', is_flag=True, default=False,
              help="This is a boolean flag. Will print the results in JSON format. Default is False.")
@click.option('-v', '--verbose', 'is_verbose', is_flag=True, default=False,
              help='Will show more information about the deployment results. Default is False.')
@click.option('-f', '--filter', 'filters', multiple=True, default=[], type=(str, str),
              help='Will show more information about the deployment results. Default is False.')
@click.option('-p', '--page', 'page', default=0, type=int,
              help='The number of the page to show. Min 0. Default is 0.')
@click.option('-s', '--size', 'size', default=25, type=int,
              help='The size of the page to show. Range 1 - 100. Default is 25.')
@click.option('--asc', default=[], multiple=True,
              help='The name of the property to sort in ascending order. Multiple flags are supported. Default is [].')
@click.option('--desc', default=[], multiple=True,
              help='The name of the property to sort in descending order. Multiple flags are supported. Default is [].')
def get(obj, product: str, entity_type_name: str, entity_ids: list[str], is_json: bool, is_verbose: bool,
        filters: list[(str, str)], page: int, size: int, asc: list[str], desc: list[str]):
  """
  Retrieves information about all the entities deployed on PDP Products. You can search by products, entity types or
  even by id. And you can filter the results by giving a property and the expected value to match the entities.
  """
  products = PRODUCTS['list']
  entity_type = get_entity_type_by_name(entity_type_name)
  if product is not None:
    raise_for_inconsistent_product(entity_type, product)
    products = [product]

  sort = []
  for asc_property in asc:
    sort += [f'{asc_property},asc']
  for desc_property in desc:
    sort += [f'{desc_property},desc']
  query_params = {
    "page": page,
    "size": size,
    "sort": sort
  }
  run_get(obj['configuration'], products, entity_type, entity_ids, filters, query_params, is_json,
          is_verbose and not is_json)


@config.command()
@click.pass_obj
@click.option('--product', default=None,
              type=click.Choice(PRODUCTS['list'],
                                case_sensitive=False
                                ),
              help="Will filter the entities based on the name entered "
                   "(Ingestion, Core or Discovery). Default is All.")
@click.option('-t', '--entity-type', 'entity_type_name', default=None,
              type=click.Choice([
                entity_type.type if entity_type.type != 'processor'
                else f'{entity_type.product}{entity_type.type.title()}'
                for entity_type in ENTITIES],
                case_sensitive=False
              ),
              help="Will filter and only show the entities of the type entered. Default is All.")
@click.option('-i', '--entity-id', 'entity_ids', default=[], multiple=True,
              help="Will only retrieve information for the component specified by the ID. Default is None. "
                   "The command allows multiple flags of -i.")
@click.option('-a', '--all', 'delete_all', default=False, is_flag=True,
              help="Will try to delete entities based on the given flags, that is, if the id is not provided by the "
                   "user, it will attempt to delete all entities of the type provided by the user, and if the type of "
                   "entity is not entered by the user, then it will attempt to delete all types of entities from "
                   "a product, and so on.")
@click.option('--cascade', 'cascade', default=False, is_flag=True,
              help="Will try to delete entity on cascade. For example: If you try to delete a pipeline, then pdp will"
                   "try to delete all the processors associated to the pipeline.")
@click.option('--local', 'local', default=False, is_flag=True,
              help="Will delete the configuration of the entities from the PDP Project files.")
@click.option('-y', '--yes', default=False, is_flag=True,
              help='Will automatically confirm the execution of the command without write "YES" or "CASCADE".')
def delete(obj, product, entity_type_name, entity_ids: list[str], delete_all, cascade: bool, local: bool, yes: bool):
  """
  Will attempt to delete the entity or entities from the product and the configuration files.
  If an entity is referenced by another can’t be deleted.
  """
  configuration = obj['configuration']
  configuration['project_path'] = obj['project_path']
  if len(entity_ids) <= 0 and not delete_all:
    raise DataInconsistency(message="You must to provide at least one entity-id or the -a flag to delete all entities.")

  products = []
  if product is None:
    products = [product for product in PRODUCTS['list'] if product != STAGING]
  else:
    products = [product]

  entity_types = []
  if entity_type_name is None:
    products = order_products_to_deploy(products)
    for product in products:
      entity_types += PRODUCTS[product]['entities']
    entity_types.reverse()
  else:
    entity_type = get_entity_type_by_name(entity_type_name)
    raise_for_inconsistent_product(entity_type, product)
    entity_types = [entity_type]

  sure_to_delete = 'YES' if yes else click.prompt(
    f"Type {click.style('YES', fg='green')} if you are sure to {click.style('delete', fg='red')} the entities",
    default=None)
  if sure_to_delete != 'YES':
    print_error("Command aborted by user.", True)
  if cascade and not yes:
    sure_to_cascade = click.prompt(
      f"Type {click.style('CASCADE', fg='green')} if you are sure "
      f"to {click.style('delete', fg='red')} the entities in cascade",
      default=None
    )
    if sure_to_cascade != 'CASCADE':
      print_error("Command aborted by user.", True)

  run_delete(configuration, entity_types, [*entity_ids], cascade, local)


@config.command()
@click.pass_obj
@click.option('--product', default=None,
              type=click.Choice(PRODUCTS['list'],
                                case_sensitive=False
                                ),
              help="Will filter the entities based on the name entered "
                   "(Ingestion, Core or Discovery). Default is All.")
@click.option('-t', '--entity-type', 'entity_type_name', default=None,
              type=click.Choice([
                entity_type.type if entity_type.type != 'processor'
                else f'{entity_type.product}{entity_type.type.title()}'
                for entity_type in ENTITIES],
                case_sensitive=False
              ),
              help="Will filter and only show the entities of the type entered. Default is All.")
@click.option('-i', '--entity-id', 'entity_id', default=None,
              help="Will only export the component specified by the ID. Default is None.")
@click.option('--include-dependencies/--no-include-dependencies', 'include_dependencies', default=False, is_flag=True,
              help="Will include those entities which are dependencies for "
                   "the entity identified with the given id. Default is False")
def export(obj: dict, product: str, entity_type_name: str, entity_id: str, include_dependencies: bool):
  """
  Will create a .zip file with the configuration files for the entities given by the user.
  """
  configuration = obj['configuration']
  configuration['project_path'] = obj['project_path']
  entity_type = None
  if entity_type_name is not None:
    entity_type = get_entity_type_by_name(entity_type_name)
    if entity_id is None:
      raise DataInconsistency(message="You must provide the --entity-id flag when you specified a --entity-type flag.")

  if entity_id is not None and entity_type_name is None:
    raise DataInconsistency(message="You must provide the --entity-type flag when you specified a --entity-id flag.")

  if product is not None:
    raise_for_inconsistent_product(entity_type, product)

  run_export(configuration, product, entity_type, entity_id, include_dependencies)


@config.command('import')
@click.pass_obj
@click.option('--product', default=None, required=True,
              type=click.Choice(PRODUCTS['list'],
                                case_sensitive=False
                                ),
              help="Will import the given file to the specified product."
                   "(Ingestion, Core or Discovery).")
@click.option('--path', '_zip', default=None, required=True,
              help="The path to the zip that will be imported.")
def _import(obj, product: str, _zip: str):
  """
  Will import a .zip to a given product. The commands assume that the zip contains the files and structure
  necessary for each product.
  """
  configuration = obj['configuration']
  if not _zip.endswith('.zip'):
    raise DataInconsistency(message=f'The path "{_zip}" is not a .zip file')
  run_import(configuration, product, _zip)
