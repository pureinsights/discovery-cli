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

import os.path
from configparser import ConfigParser

import click
import pyfiglet

from commands.config.command import config
from commons.constants import CORE, DEFAULT_CONFIG, DISCOVERY, INGESTION, STAGING
from commons.custom_classes import DataInconsistency
from commons.handlers import handle_exceptions


def ensure_configurations(config: dict):
  """
  It assures all configurations were loaded and if not, it uses defaults values.
  :param dict config: The configurations to analyze.
  :rtype: dict
  :return: The config dict but with defaults values on those missing configurations.
  """
  properties: list[str] = [INGESTION, DISCOVERY, CORE, STAGING]

  for property in properties:
    if config.get(property, None) is None:
      config[property] = DEFAULT_CONFIG.get(property, None)

  return config


def load_config(config_name: str, profile: str = 'DEFAULT'):
  """
  Implement profiles me with configparser (i.e. the idea is to be able to chose between profiles easily
  like we do on kubectl or aws-cli). Reference: https://docs.python.org/3/library/configparser.html
  """
  config = ConfigParser()
  configuration = { }
  if os.path.exists(config_name):
    config.read(config_name)
    if config.has_section(profile) or profile == 'DEFAULT':
      configuration = config[profile]
    else:
      raise DataInconsistency(message=f'Configuration profile {profile} was not found.')
  return ensure_configurations(configuration)


@click.group()
@click.option('--namespace', default='pdp', help='Namespace in which the PDP components are running. Default is "pdp".')
@click.option('--profile',
              help='Configuration profile to load specific configurations from pdp.ini. Default is "DEFAULT"')
@click.pass_context
def pdp(ctx, namespace: str, profile: str | None):
  """
  This is the official PureInsights Discovery Platform CLI.
  """
  # ensure that ctx.obj exists and is a dict (in case `cli()` is called
  # by means other than the `if` block below)
  ctx.ensure_object(dict)
  ctx.obj['namespace'] = namespace
  ctx.obj['configuration'] = load_config('pdp.ini', profile)


@pdp.command()
def health():
  """
  This command is used to know if PDP-CLI has been installed successfully.
  """
  ascii_art_pdp_cli = pyfiglet.figlet_format("PDP - CLI")
  title = "Pureinsights Discovery Platform: Command Line Interface"
  version = "v1.5.0"
  url = "https://pureinsights.com/"
  click.echo(f"{ascii_art_pdp_cli}{title}\n{version}")
  click.echo(click.style(url, fg="blue", underline=True, bold=True))


# Register all the commands
pdp.add_command(config)

if __name__ == '__main__':
  # TODO: Delete all the unnecessary code here
  # TODO: Document all the functions
  # handle_exceptions(pdp, ["--profile", "FAKE", "config", "init", "--no-empty", "--force"])
  handle_exceptions(pdp)  # pragma: no cover
