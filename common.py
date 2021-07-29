import constants


def replace_ids(data, id_to_name, entity_name):
    for entity in data:
        entity_id = entity['id']
        id_to_name[entity_id] = entity.get('name', entity_id)

        context = {
            'id': entity_id,
            'type': entity.get('type', entity_name[1]),
            'fileName': entity_name[0]
        }

        for reference_field in constants.reference_fields:
            replace_id(entity, reference_field, id_to_name, context)


def replace_id(data, k, id_to_name, context):
    if not type(data) is dict:
        return

    for key in data.keys():
        if key == k:
            referenced_id = data[key]
            if referenced_id not in id_to_name:
                print(f'Error: Id {referenced_id} does not exist while attempting to replace in {data} with {context}')
                return

            data[key] = f"{{{{ fromName('{id_to_name[referenced_id]}') }}}}"
        elif type(data[key]) is dict:
            replace_id(data[key], k, id_to_name, context)
        elif type(data[key]) is list:
            [replace_id(nested, k, id_to_name, context) for nested in data[key]]