import click

@click.command()
@click.pass_context
def seed(ctx):
    click.echo(f"namespace is {ctx.obj['namespace']}")
