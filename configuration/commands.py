import click
from configuration import init as initialize
from configuration import deploy as deployment


@click.group()
@click.pass_context
def config(ctx):
    """
    Provides common configuration operations, such as export, import or creating a new project from scratch.
    """
    pass


@config.command()
@click.pass_context
@click.option('-i', '--ignore-ids/--no-ignore-ids', default=False, show_default=True,
              help="will cause existing ids to be ignored, hence everything will be created as a new instance. This "
                   "is useful when moving configs from one instance to another.")
def deploy(ctx, ignore_ids):
    """
    Deploys project configurations to the target Admin API. Must be run within the directory from a project created
    with the 'init' command.

    Entities are created in this order: credentials, processors, pipelines, seeds and then cron_jobs.

    Will replace any name reference with IDs. Names are case insensitive.

    If the "id" field is missing from an entity, assumes this is a new instance.
    """
    deployment.run(ctx, ignore_ids)


@config.command()
@click.option('-p', '--project-name', default='my-pdp-project',
              help="the name of the resulting directory, will try to fetch "
                   "existing configurations from the Admin API referenced "
                   "in ~/.pdp. Notice that imported configs have id "
                   "fields, don't change those.")
@click.option('-e', '--empty/--no-empty', default=False, help='if it should only create an empty directory structure with '
                                                        'basic handlebars for starting a new project.')
@click.option('-a', '--admin-api-url', default="http://localhost:8080", show_default=True,
              help='the base URL for the Admin API')
def init( project_name, empty, admin_api_url):
    """
    Creates a new project from existing sources or from scratch.
    """
    initialize.run(project_name, empty, admin_api_url)
