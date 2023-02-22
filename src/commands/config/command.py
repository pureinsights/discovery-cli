import click

@click.group()
def config():
    pass

@config.command()
@click.option('-n','--project-name', default='my-pdp-project', help='The name of the resulting directory, will try to fetch existing configurations from the APIs referenced in ~/.pdp. Notice that imported configs have id fields, don`t change those. Default is my-pdp-project.')
def init(project_name):
    """
    Creates a new project from existing sources or from scratch.
    """
    click.secho(f"Project {project_name} creted successfully.")