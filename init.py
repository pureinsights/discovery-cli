import common
import configparser
import constants
import json
import os
import requests
import shutil
import zipfile


def description():
    return 'init: start here, creates a new project from existing sources or from scratch.'


def print_help():
    print("""
Creates a new project form existing sources or from scratch.    

Usage:
    pdp init [projectName] [--empty] --adminApiUrl=http://localhost:8080
    
    * projectName: the name of the resulting directory, will try to fetch existing configurations from the Admin 
                   API referenced in ~/.pdp. **Notice that imported configs have id fields, don't change those**.
    * adminApiUrl: the base URL for the Admin API, defaults to http://localhost:8080
    * --empty: if it should only create an empty directory structure with basic handlebars for starting
               a new project. 
    """)


def run(argv, commands, configuration):
    if len(argv) < 3:
        print_help()
        return

    project_name = argv[2]
    admin_api_url = admin_api_url_from_args(argv)

    if '--empty' in argv:
        create_empty_project(project_name)
    else:
        # Create empty directory
        project_path = os.path.join('./', project_name)
        os.mkdir(project_path)

        # Download and extract zip file
        export_all(project_name, admin_api_url)

        # Write ini file
        project_configuration = configparser.RawConfigParser()
        project_configuration['DEFAULT'] = {'AdminApiUrl': admin_api_url}

        with open(f'{project_name}/pdp.ini', 'w') as file:
            project_configuration.write(file)


def create_empty_project(project_name):
    try:
        # Create sample files
        shutil.copytree('templates', project_name)
        return
    except OSError as error:
        print(f'Failed to init project due {error}')


def export_all(project_name, admin_api_url):
    # Get export file
    response = requests.get(f'{admin_api_url}/export/all')
    response.raise_for_status()

    zip_file_name = f'{project_name}/export.zip'

    with open(zip_file_name, 'wb') as file:
        file.write(response.content)

    # Extract zip files
    with zipfile.ZipFile(zip_file_name, 'r') as zip_ref:
        zip_ref.extractall(project_name)

    os.remove(zip_file_name)

    # Then replace IDs with names and pretty print
    id_to_name = {}

    for entity_name in constants.entity_names:
        with open(f'{project_name}/{entity_name[0]}', 'r+') as file:
            data = json.load(file)

            common.replace_ids(data, id_to_name, entity_name[1])

            file.seek(0)
            json.dump(data, file, indent=2)
            file.truncate()


def admin_api_url_from_args(argv):
    for arg in argv:
        if arg.startswith('--adminApiUrl'):
            return arg[arg.index('=') + 1:len(arg)]

    return 'http://localhost:8080'
