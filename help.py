import os


def description():
    return "help: prints help options"


def print_help():
    print("Gives a summary of all help options")


def run(argv, commands, configuration):
    if len(argv) > 2:
        command = argv[2]
        commands[command].print_help()
    else:
        print("""
Type 'pdp help {command}' to get information about that specific command. See the complete list of available commands
below:
        """)

        for command in commands.values():
            print(command.description())
