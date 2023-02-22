from click.testing import CliRunner
from commands.config.command import init


cli = CliRunner()

def test_init():
    project_name = "Hello world"
    response = cli.invoke(init, ["-n",project_name])
    assert response.exit_code == 0
    assert f"Project {project_name} creted successfully." in response.output