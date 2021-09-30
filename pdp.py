import click
import configparser

from cluster import commands as cluster_commands
from configuration import commands as config_commands
from seed import commands as seed_commands


def load_config(config_name):
    # Implement profiles me with configparser (i.e. the idea is to be able to chose between profiles easily
    # like we do on kubectl or aws-cli). Reference: https://docs.python.org/3/library/configparser.html
    config = configparser.ConfigParser()
    config.read(config_name)
    return config['DEFAULT']

@click.group()
@click.option('--namespace', default="pdp", help='Namespace in which the PDP components are running. Default is "pdp".')
@click.pass_context
def cli(ctx, namespace):
    # ensure that ctx.obj exists and is a dict (in case `cli()` is called
    # by means other than the `if` block below)
    ctx.ensure_object(dict)
    ctx.obj['namespace'] = namespace
    ctx.obj['configuration'] = load_config('pdp.ini')


cli.add_command(cluster_commands.cluster)
cli.add_command(config_commands.config)
cli.add_command(seed_commands.seed)


if __name__ == '__main__':
    cli(obj={})
