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
from configparser import ConfigParser

import click
import pyfiglet

from commands.config.command import config
from commands.core.command import core
from commands.core.file.command import file
from commands.execution.command import seed_exec
from commands.staging.bucket.command import bucket_command as bucket
from commands.staging.command import staging
from commands.staging.item.command import item
from commands.staging.transaction.command import transaction
from commons.constants import DEFAULT_CONFIG, PRODUCTS
from commons.custom_classes import DataInconsistency
from commons.handlers import handle_exceptions


def ensure_configurations(config: dict):
  """
  Assures all configurations were loaded and if not, uses defaults values.
  :param dict config: The configurations to analyze.
  :rtype: dict
  :return: The config dict but with defaults values on those missing configurations.
  """
  properties: list[str] = PRODUCTS['list']
  for property in properties:
    if config.get(property, None) is None:
      config[property] = DEFAULT_CONFIG.get(property, None)

  return config


def load_config(config_name: str, profile: str = 'DEFAULT'):
  """
  Implement profiles with configparser (i.e. the idea is to be able to chose between profiles easily
  like we do on kubectl or aws-cli). Reference: https://docs.python.org/3/library/configparser.html
  """
  config = ConfigParser()
  configuration = {}
  if os.path.exists(config_name):
    config.read(config_name)
    if config.has_section(profile) or profile == 'DEFAULT':
      configuration = {**config[profile]}
    else:
      raise DataInconsistency(message=f'Configuration profile {profile} was not found.')
  return ensure_configurations(configuration)


@click.group()
@click.option('--profile', default='DEFAULT',
              help='Configuration profile to load specific configurations from pdp.ini. Default is "DEFAULT"')
@click.option('--path', default='.', help='The path to a directory with the structure and the pdp.ini '
                                          'that init command creates. Default is ./.')
@click.pass_context
def pdp(ctx, path: str, profile: str):
  """
  This is the official Pureinsights Discovery Platform CLI.
  """
  # ensure that ctx.obj exists and is a dict (in case `cli()` is called
  # by means other than the `if` block below)
  ctx.ensure_object(dict)
  ctx.obj['profile'] = profile
  ctx.obj['project_path'] = path
  config_path = os.path.join(path, 'pdp.ini')
  ctx.obj['configuration'] = load_config(config_path, profile)


@pdp.command()
def health():
  """
  This command is used to know if PDP-CLI has been installed successfully.
  """
  ascii_art_pdp_cli = pyfiglet.figlet_format("PDP - CLI")
  title = "Pureinsights Discovery Platform: Command Line Interface"
  url = "https://pureinsights.com/"
  click.echo(f"{ascii_art_pdp_cli}{title}\nv1.5.0")
  click.echo(click.style(url, fg="blue", underline=True, bold=True))


# Register all the commands
pdp.add_command(config)
pdp.add_command(core)
pdp.add_command(seed_exec)
pdp.add_command(staging)

core.add_command(file)
staging.add_command(item)
staging.add_command(bucket)
staging.add_command(transaction)

if __name__ == '__main__':
  os.system("")  # pragma: no cover
  handle_exceptions(pdp)  # pragma: no cover
