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
    title = "Pureinsights Discovery Platform: Command Line Interface"
    version = "v1.5.0"
    url = "https://pureinsights.com/"
    click.echo(f"{ascii_art_pdp_cli}{title}\n{version}")
    click.echo(click.style(url, fg="blue", underline=True, bold=True))

#Register all the commands
pdp.add_command(config)

if __name__ == '__main__':
    pdp()