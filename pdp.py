import cluster
import deploy
import help
import init
import sys

import configparser

commands = {
    'init': init,
    'deploy': deploy,
    'cluster': cluster,
    'help': help
}


def print_help():
    print("""
PDP Command Line Interface. Type 'pdp help' for information about available commands.
    """)


def load_config(config_name):
    # Implement profiles me with configparser (i.e. the idea is to be able to chose between profiles easily
    # like we do on kubectl or aws-cli). Reference: https://docs.python.org/3/library/configparser.html
    config = configparser.ConfigParser()
    config.read(config_name)
    return config['DEFAULT']


def main():
    if len(sys.argv) < 2:
        print_help()
        sys.exit()

    # Get configuration
    configuration = load_config('pdp.ini')

    command = sys.argv[1]
    commands[command].run(sys.argv, commands, configuration)


main()
