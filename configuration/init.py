import click
from configuration import common
import configparser
from configuration import constants
import json
import os
import requests
import shutil
import zipfile


def run(project_name, empty, admin_api_url):
    if empty:
        create_empty_project(project_name)
    else:
        # Create directory
        project_path = os.path.join('../', project_name)
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
        shutil.copytree('configuration/templates', project_name)
        return
    except OSError as error:
        click.echo(f'Failed to init project due {error}')


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
