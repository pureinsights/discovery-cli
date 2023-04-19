#  Copyright (c) 2023 Pureinsights Technology Ltd. All rights reserved.
#
#  Permission to use, copy, modify or distribute this software and its
#  documentation for any purpose is subject to a licensing agreement with
#  Pureinsights Technology Ltd.
#
#  All information contained within this file is the property of
#  Pureinsights Technology Ltd. The distribution or reproduction of this
#  file or any information contained within is strictly forbidden unless
#  prior written permission has been granted by Pureinsights Technology Ltd.

from commons.constants import DEFAULT_CONFIG
from pdp import pdp
from pdp_test import cli


def test_config(snapshot):
  """
  Test the command defined in :func:`src.commands.config.command.config`.
  """
  response = cli.invoke(pdp, ["config"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_config.snapshot')


def test_init_success(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.config.command.init`,
  without arguments.
  """
  init_run_mocked = mocker.patch('commands.config.command.run_init', return_value=True)
  project_name = "my-pdp-project"
  response = cli.invoke(pdp, ["config", "init", "--empty", "--template", "empty"])
  init_run_mocked.assert_called()
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_init_success.snapshot')


def test_init_could_not_create(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.config.command.init`,
  when some error happens.
  """
  init_run_mocked = mocker.patch('commands.config.command.run_init', return_value=False)
  response = cli.invoke(pdp, ["config", "init"])
  init_run_mocked.assert_called()
  assert response.exit_code == 1
  snapshot.assert_match(response.output, 'test_init_could_not_create.snapshot')


def test_init_parse_options(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.config.command.init`,
  with all the arguments provided.
  """
  init_run_mocked = mocker.patch('commands.config.command.run_init', return_value=True)
  project_name = "my-pdp-project"
  no_empty = '--no-empty'
  expected_config = {
    'ingestion': 'http://ingestion-fake',
    'discovery': 'http://ingestion-fake',
    'core': 'http://ingestion-fake',
    'staging': 'http://ingestion-fake'
  }
  force = '--force'
  response = cli.invoke(pdp,
                        ["config", "init", "-n", project_name, no_empty, force, '--template', 'empty', '-u',
                         'ingestion',
                         'http://ingestion-fake', '-u', 'discovery', 'http://ingestion-fake', '-u', 'core',
                         'http://ingestion-fake', '-u', 'staging', 'http://ingestion-fake'])

  init_run_mocked.assert_called_once_with(project_name, expected_config, True, None)
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_init_parse_options.snapshot')


def test_init_incorrect_option_product(snapshot):
  """
  Test the command defined in :func:`src.commands.config.command.init`,
  with an unrecognized product of the argument product-url.
  """
  project_name = "my-pdp-project"
  response = cli.invoke(pdp, ["config", "init", '-u', 'fake-product',
                              'http://ingestion-fake'])
  assert response.exit_code == 1
  snapshot.assert_match(response.output, 'test_init_incorrect_option_product.snapshot')


def test_init_without_load_config_on_init_command(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.config.command.init`,
  when the configuration 'load_config' is False.
  """
  mocker.patch("pdp.os.path.exists", returned_value=False)
  response = cli.invoke(pdp, ["config", "init", "-n", "fake-name", "--force"])
  assert response.exit_code == 1
  snapshot.assert_match(response.output, 'test_init_without_load_config_on_init_command.snapshot')


def test_deploy_success(mocker, snapshot, test_project_path):
  """
  Test the command defined in :func:`src.commands.config.command.deploy`,
  without arguments.
  """
  project_path = test_project_path()
  run_deploy_mock = mocker.patch("commands.config.command.run_deploy")
  response = cli.invoke(pdp, ["-d", project_path, "config", "deploy"])
  assert response.exit_code == 0
  targets = ('core', 'ingestion', 'discovery')
  run_deploy_mock.assert_called_once_with(DEFAULT_CONFIG, project_path, targets, False, False, False)
  snapshot.assert_match(response.output, 'test_deploy_success.snapshot')


def test_deploy_without_load_config_on_deploy_command(mocker, snapshot, test_project_path):
  """
  Test the command defined in :func:`src.commands.config.command.deploy`,
  without arguments.
  """
  project_path = test_project_path()
  mocker.patch("pdp.os.path.exists", returned_value=False)
  run_deploy_mock = mocker.patch("commands.config.command.run_deploy")
  response = cli.invoke(pdp, ["-d", project_path, "config", "deploy"])
  assert response.exit_code == 0
  targets = ('core', 'ingestion', 'discovery')
  run_deploy_mock.assert_called_once_with(DEFAULT_CONFIG, project_path, targets, False, False, False)
  snapshot.assert_match(response.output, 'test_deploy_without_load_config_on_deploy_command.snapshot')


def test_create_successfully(mocker, snapshot, test_project_path):
  """
  Test the command defined in :func:`src.commands.config.command.create`,
  when the flag --file was provided.
  """
  mocker.patch("commands.config.create.create_spinner")
  mocker.patch("commands.config.create.create_or_update_entity", return_value="newId")
  mocker.patch("commands.config.create.raise_for_pdp_data_inconsistencies")
  mocker.patch("commands.config.create.write_entities")
  response = cli.invoke(pdp, ["-d", test_project_path(), "config", "create", "--entity-type", "pipeline", "--file",
                              test_project_path('custom_pipeline.json')])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_create_successfully.snapshot')


def test_create_with_entity_template(mocker, snapshot, test_project_path):
  """
  Test the command defined in :func:`src.commands.config.command.create`,
  when a template for an entity is provided.
  """
  mocker.patch("commands.config.create.create_spinner")
  mocker.patch("commands.config.create.create_or_update_entity", return_value="newId")
  mocker.patch("commands.config.create.raise_for_pdp_data_inconsistencies")
  response = cli.invoke(pdp, ["-d", test_project_path(), "config", "create", "--entity-type", "pipeline",
                              "--entity-template", "empty_pipeline", "--deploy", "--json"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_create_with_entity_template.snapshot')


def test_create_with_entity_template_and_no_file(mocker, snapshot, test_project_path):
  """
  Test the command defined in :func:`src.commands.config.command.create`,
  when a template and a file were not provided.
  """
  mocker.patch("commands.config.create.create_spinner")
  mocker.patch("commands.config.create.raise_for_pdp_data_inconsistencies")
  response = cli.invoke(pdp, ["-d", test_project_path(), "config", "create", "--entity-type", "pipeline"])
  assert response.exit_code == 1
  snapshot.assert_match(response.output, 'test_create_with_entity_template_and_no_file.snapshot')


def test_create_entity_template_not_supported(mocker, snapshot, test_project_path):
  """
  Test the command defined in :func:`src.commands.config.command.create`,
  when the template provided is not supported.
  """
  mocker.patch("commands.config.create.create_spinner")
  mocker.patch("commands.config.create.raise_for_pdp_data_inconsistencies")
  response = cli.invoke(pdp, ["-d", test_project_path(), "config", "create", "--entity-type", "pipeline",
                              "--entity-template", "fake_template"])
  assert response.exit_code == 1
  snapshot.assert_match(response.output, 'test_create_entity_template_not_supported.snapshot')


def test_create_with_entity_template_and_no_file_but_is_interactive(mocker, snapshot, test_project_path):
  """
  Test the command defined in :func:`src.commands.config.command.create`,
  when a template and a file were not provided, but the flag interactive is True and also the deployment.
  """
  mocker.patch("commands.config.create.raise_for_pdp_data_inconsistencies")
  mocker.patch("commands.config.create.click.edit", return_value='{ "name": "Pipeline", "active": true, "steps": [] }')
  mocker.patch("commands.config.create.create_spinner")
  mocker.patch("commands.config.create.create_or_update_entity", return_value="newId")
  response = cli.invoke(pdp, ["-d", test_project_path(), "config", "create", "--entity-type", "pipeline",
                              "--interactive", "--ignore-ids", "--deploy"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_create_with_entity_template_and_no_file_but_is_interactive.snapshot')


def test_get_all_entities(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.config.command.get`,
  without flags.
  """
  mocker.patch("commands.config.get.get")
  mocker.patch("commands.config.get.json.loads", return_value={"content": [{'id': 'fake-id', 'name': 'fake-name'}]})
  response = cli.invoke(pdp, ["config", "get"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_get_all_entities.snapshot')


def test_get_all_entities_without_entities(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.config.command.get`,
  without flags but without entities on the products.
  """
  mocker.patch("commands.config.get.get")
  mocker.patch("commands.config.get.json.loads", return_value={"content": []})
  response = cli.invoke(pdp, ["config", "get"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_get_all_entities_without_entities.snapshot')


def test_get_entities_from_product_verbose(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.config.command.get`,
  when a product was provided, with --verbose flag activated.
  """
  mocker.patch("commands.config.get.create_spinner")
  mocker.patch("commands.config.get.get")
  mocker.patch("commands.config.get.json.loads", return_value={"content": [
    {'id': 'fake-id', 'name': 'fake-name', 'description': None}, {'id': 'fake-id', 'name': 'fake-name', "active": True}
  ]})
  response = cli.invoke(pdp, ["config", "get", "--product", "ingestion", "-v"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output.replace('\r', '\n'), 'test_get_entities_from_product_verbose.snapshot')


def test_get_entities_with_ids_and_types_json_flag(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.config.command.get`,
  when a type and ids were provided, with --json flag activated.
  """
  mocker.patch("commands.config.get.get")
  mocker.patch("commands.config.get.json.loads", return_value={"content": [{'id': 'fake-id', 'name': 'fake-name'}]})
  response = cli.invoke(pdp, ["config", "get", "--entity-type", "pipeline", "-j"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_get_entities_with_ids_and_types_json_flag.snapshot')


def test_get_entities_by_ids_and_filtered_by_active(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.config.command.get`,
  when an id and a filter was provided.
  """
  mocker.patch("commands.config.get.create_spinner")
  mocker.patch("commands.config.get.get")
  mocker.patch(
    "commands.config.get.json.loads",
    return_value={'id': '6376af03-1af2-41a2-aef6-62aefc73a870', 'name': 'fake-name1', 'description': None}
  )
  response = cli.invoke(pdp, ["config", "get", "-i", "6376af03-1af2-41a2-aef6-62aefc73a870", "-i",
                              "fake-id", "-f", "active", "True",
                              "--asc", "name", "--desc", "id", "-v"])
  assert response.exit_code == 0
  snapshot.assert_match(response.output, 'test_get_entities_by_ids_and_filtered_by_active.snapshot')


def test_get_invalid_type_for_product(mocker, snapshot):
  """
  Test the command defined in :func:`src.commands.config.command.get`,
  when the given entity type don't belong to the given product.
  """
  mocker.patch("commands.config.get.create_spinner")
  response = cli.invoke(pdp, ["config", "get", "--product", "discovery", "--entity-type", "credential"])
  assert response.exit_code == 1
  snapshot.assert_match(str(response.exception), 'test_get_invalid_type_for_product.snapshot')
