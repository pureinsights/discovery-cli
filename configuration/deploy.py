import click
from configuration import common
from configuration import constants
import json
import requests

from jinja2 import Template, TemplateSyntaxError

name_to_id = {}


def from_name(name):
    name = name.lower()

    if name not in name_to_id:
        raise common.DataInconsistencyException(f'Woops, this should not happen, but seems there is a name "{name}" '
                                                f'that has no id mapping')

    return name_to_id[name]


def run(ctx, ignore_ids):
    admin_api_url = ctx.obj['configuration'].get('AdminApiUrl')

    id_to_name = {}

    for entity_name in constants.entity_names:
        with open(entity_name[0], 'r+') as file:
            data = file.read()

            try:
                # Replace names for IDs if any
                rendered = replace_names_for_ids(data, entity_name)
                entities = json.loads(rendered)

                for entity in entities:
                    entity_id = entity.get('id', None)

                    if ignore_ids:
                        entity_id = None

                    if entity_id:
                        # Update
                        response = requests.put(
                            f'{admin_api_url}/{entity_name[1]}/{entity_id}',
                            data=json.dumps(entity),
                            headers={'Content-type': 'application/json'}
                        )

                        if response.status_code == requests.codes.bad:
                            click.echo(
                                f'\nError while updating entity {entity_id}: \n{json.dumps(response.json(), indent=4, sort_keys=True)}\n')
                            response.raise_for_status()

                        click.echo(f'Updated entity of type {entity_name[1]} with id {entity_id}')
                    else:
                        # Create
                        response = requests.post(
                            f'{admin_api_url}/{entity_name[1]}',
                            data=json.dumps(entity),
                            headers={'Content-type': 'application/json'}
                        )

                        if response.status_code == requests.codes.bad:
                            click.echo(
                                f'\nError while creating entity: \n{json.dumps(response.json(), indent=4, sort_keys=True)}\n')
                            response.raise_for_status()

                        entity_id = response.json()['id']
                        entity['id'] = entity_id
                        click.echo(f'Created new entity of type {entity_name[1]} with id {entity_id}')

                    # Cron jobs don't have a name
                    if 'name' in entity:
                        name_to_id[entity['name'].lower()] = entity_id

                # Replace new IDs in files
                common.replace_ids(entities, id_to_name, entity_name)

                file.seek(0)
                json.dump(entities, file, indent=2)
                file.truncate()
            except ValueError as error:
                click.echo(f'File {entity_name[0]} is not a valid JSON. Please fix formatting issues and try again.')
                raise error


def replace_names_for_ids(data, entity_name):
    try:
        template = Template(data)
        template_fields = {'fromName': from_name}
        return template.render(**template_fields)
    except TemplateSyntaxError as template_error:
        click.echo(f'Error: evaluating file {entity_name[0]} with detail {template_error}')
        raise template_error
