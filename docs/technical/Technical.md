# PDP-CLI - ** Technical Documentation **

This client uses [Click](https://click.palletsprojects.com/en/8.1.x/#documentation) under the hood to make it easy to develop and document new commands. You can also use [Pydoc](https://docs.python.org/3/library/pydoc.html) to get the docstring of the CLI and has a better understanding.

> **_NOTE:_** We'll goint to refer to command options (@click.option) as flags, and as arguments to the positional arugments.

## Dependencies

* Python 3.10.9

## Development guidelines

### Usage and commands documentations

Click provides an auto-generated documenation for each command, the user can access by passing the --help flag. But as a developer we have the responsability to provide more details to these documentation through the docstrings and "help" parameters of Click.

For example:

```python
@click.command()
@click.option('-f','--foo',help='Explanation about this option.')
def bar(foo):
    """
    This is an example of the documentation usage for the command 'foo'.
    """
    pass
```

### Arguments and Flags

Try to avoid the use of positional arguments and use flags instead. Although using flags requires the user to write more, it is actually better because it does not force the user to memorize the order of the arguments. Additionally, by using meaningful names for each flag, it is easier for the user.

#### Flag/Argument naming conventions

##### Names with two or more words
* Words must be separated by '-'. For example, you can name a flag that specifies the project name as --project-name.
* Try to add a short flag for simplicity. Example -n, --project-name.
* For those flags who can be passed multiple times, try to use short names and short flags.

##### Short flags 
* Always use the same short flag for the same flag. That means, if you used -d as short flag of --dry-run you can't use -d, even in other commands, for another flag. 
* Try always to use a meaningful short flag. A good aproach could be use the first letter of the flag name, while this doesn't affect the previous convention.

## Developer Getting started

### Virtual environment

Create and activate a virtual environment:
```bash
virtualenv venv
.\venv\Scripts\activate

#For Linux used t he below
#.venv/bin/activate
```
You can use [conda](https://docs.conda.io/en/latest/) if you prefer to.

Pull the dependencies:
```bash
pip install -r requirements.txt
```

Invoke the health check command to verify that the CLI can run correctly.
```bash
cd ./src/
python pdp.py health
```

### Getting a .exe file for the CLI

## Testing

### Unit testing

For unit testing we'll be using [Pytest](https://www.bing.com/search?q=pytest&cvid=999475cf79b6410488992254b2f2dd92&aqs=edge.0.69i59l4j0l2j69i60l3j69i11004.3619j0j4&FORM=ANAB01&PC=U531). In the root directory there is a directory called tests, here is where we have all the tests. The test directory must have the same structure than the src directory.

Each test must start with the same name of the file that you want to test and finish with "_test.py".

### Integration testing
