import click
import json
import requests

session = requests.Session()


@click.group()
@click.pass_context
def seed(ctx):
    """
     Subcommands for getting seed information or interacting with seeds.
    """
    pass


@seed.command()
@click.pass_context
@click.option('-a', '--action', required=True,
              type=click.Choice(['start', 'halt', 'pause', 'resume'], case_sensitive=False),
              help="The action to be taken.")
@click.option('--scan-type',
              type=click.Choice(['full', 'incremental'], case_sensitive=False),
              help="If --action is 'start', then select which type of execution to perform.")
@click.option('--all-seeds/--no-all-seeds', default=False, show_default=True,
              help="If it should halt all seeds. Takes precedence over the --seed-id option.")
@click.option('-s', '--seed-id',
              help="The seed ID that will be halted.")
def control(ctx, action, scan_type, all_seeds, seed_id):
    """
     Instructs a seed to perform a given action.
    """
    admin_api_url = ctx.obj['configuration'].get('AdminApiUrl')

    if action.upper() == 'start' and not scan_type:
        raise click.BadParameter("if action is start then --scan-type must be set.", param_hint="scan-type")

    # Use the seed ID provided
    if not all_seeds:
        if not seed_id:
            raise click.BadParameter("if --all-seeds is empty or set to --no-all-seeds (default), you must provide a "
                                     "non empty seed ID.", param_hint="seed-id")

        control_seed(seed_id, action, scan_type, admin_api_url)
        return

    # If not, then get all the seeds that exist and try all of them
    for seed_list in get_seeds(admin_api_url, True):
        for seed_response in seed_list['content']:
            if seed_response['status'] != 'ENQUEUED':
                seed_id_response = seed_response['seedId']
                control_seed(seed_id_response, action, scan_type, admin_api_url)


@seed.command()
@click.pass_context
@click.option('-s', '--size', default=25, show_default=True,
              help="The number of seeds to get.")
@click.option('-p', '--page', default=0, show_default=True,
              help="The page to retrieve.")
def get(ctx, size, page):
    """
    Gets a list of seeds
    """
    admin_api_url = ctx.obj['configuration'].get('AdminApiUrl')
    seed_list = get_seeds_request(admin_api_url, size, page, False)
    click.echo(f"Found {seed_list['totalSize']} seeds, and {seed_list['totalPages']} pages.")

    for seed_response in seed_list['content']:
        click.echo(json.dumps(seed_response, indent=4, sort_keys=True))


def get_seeds_request(admin_api_url, size, page, active):
    url = f'{admin_api_url}/seed'
    if active:
        url = f'{url}/execution'

    response = session.get(url, params={'size': size, 'offset': page * size})

    if response.status_code == requests.codes.bad:
        click.echo(f'\nFailed to fetch seeds with error: \n{json.dumps(response.json(), indent=4, sort_keys=True)}\n')
        response.raise_for_status()

    return response.json()


def get_seeds(admin_api_url, active):
    first_page = get_seeds_request(admin_api_url, 100, 0, active)
    yield first_page

    num_pages = first_page['totalPages']

    for page in range(1, num_pages):
        next_page = get_seeds_request(admin_api_url, 100, page, active)
        yield next_page


def control_seed(seed_id, action, scan_type, admin_api_url):
    if action.upper() == 'START':
        response = requests.post(
            f'{admin_api_url}/seed/{seed_id}?scanType={scan_type.upper()}',
            data='{}',
            headers={'Content-type': 'application/json'}
        )
    else:
        response = requests.put(
            f'{admin_api_url}/seed/{seed_id}/control?action={action.upper()}',
            data='{}',
            headers={'Content-type': 'application/json'}
        )

    if response.status_code == requests.codes.bad:
        click.echo(
            f'\nAction {action} failed to be applied on seed {seed_id} with error: \n{json.dumps(response.json(), indent=4, sort_keys=True)}\n')
        response.raise_for_status()

    click.echo(f'Action {action} executed on seed {seed_id}')
