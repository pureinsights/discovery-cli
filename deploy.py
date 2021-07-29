import common
import constants
import json
import requests

from jinja2 import Template, TemplateSyntaxError

name_to_id = {}


def description():
    return 'deploy: push configurations to a target environment'


def print_help():
    print("""
Deploys project configurations to the target Admin API. Must be run within the directory from a project created with the 'init'
command.

Entities are created in this order: credentials, processors, pipelines, seeds and then cron_jobs.

Will replace any name reference with IDs. Names are case insensitive.

If the "id" field is missing from an entity, assumes this is a new instance.

Usage:
    pdp deploy [--ignore-ids]

    *--ignore-ids: will cause existing ids to be ignored, hence everything will be created as a new 
                   instance. This is useful when moving configs from one instance to another.
    """)


def from_name(name):
    name = name.lower()

    if name not in name_to_id:
        raise Exception(f'Woops, this should not happen, but seems there is a name "{name}" that has no id mapping')

    return name_to_id[name]


def run(argv, commands, configuration):
    admin_api_url = configuration.get('AdminApiUrl')

    id_to_name = {}

    for entity_name in constants.entity_names:
        with open(entity_name[0], 'r+') as file:
            data = file.read()

            try:
                # Replace names for IDs if any
                try:
                    template = Template(data)
                    template_fields = {'fromName': from_name}
                    rendered = template.render(**template_fields)
                    entities = json.loads(rendered)

                    for entity in entities:
                        entity_id = entity.get('id', None)

                        if '--ignore-ids' in argv:
                            entity_id = None

                        if entity_id:
                            # Update
                            response = requests.put(
                                f'{admin_api_url}/{entity_name[1]}/{entity_id}',
                                data=json.dumps(entity),
                                headers={'Content-type': 'application/json'}
                            )

                            response.raise_for_status()
                            print(f'Updated entity of type {entity_name[1]} with id {entity_id}')
                        else:
                            # Create
                            response = requests.post(
                                f'{admin_api_url}/{entity_name[1]}',
                                data=json.dumps(entity),
                                headers={'Content-type': 'application/json'}
                            )

                            response.raise_for_status()
                            entity_id = response.json()['id']
                            entity['id'] = entity_id
                            print(f'Created new entity of type {entity_name[1]} with id {entity_id}')

                        # Cron jobs don't have a name
                        if 'name' in entity:
                            name_to_id[entity['name'].lower()] = entity_id

                    # Replace new IDs in files
                    common.replace_ids(entities, id_to_name, entity_name)

                    file.seek(0)
                    json.dump(entities, file, indent=2)
                    file.truncate()

                except TemplateSyntaxError as template_error:
                    raise Exception(f'Error: evaluating file {entity_name[0]} with detail {template_error}')

            except ValueError as error:
                raise Exception(f'File {entity_name[0]} is not a valid JSON. Please fix formatting issues and try again.')
