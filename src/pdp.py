import click 
import pyfiglet

from commands.config.command import config

@click.group()
def pdp():
    """
    This is the official PureInsights Discovery Platform CLI.
    """
    pass

@pdp.command()
def health():
    """
    This command is used to know if PDP-CLI has been installed successfully.
    """
    ascii_art_pdp_cli = pyfiglet.figlet_format("PDP - CLI")
    click.echo(f"{ascii_art_pdp_cli}Thank you for use PDP-CLI.")

#Register all the commands
pdp.add_command(config)

if __name__ == '__main__':
    pdp()