from pdp_test import cli
from commands.config.command import config,init



def test_config():
    """
    Should end with an exit code 0.
    """
    response = cli.invoke(config,[])
    assert response.exit_code == 0

def test_init():
    project_name = "Hello world"
    response = cli.invoke(init, ["-n",project_name])
    assert response.exit_code == 0
    assert f"Project {project_name} creted successfully." in response.output