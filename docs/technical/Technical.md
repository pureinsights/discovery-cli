# PDP Command Line Interface
Thin client around PDP Admin UI for common tasks. Intended for performing quick changes and as the foundation of more automation.

This client uses [Click](https://click.palletsprojects.com/en/8.0.x/) under the hood to make it easy to develop and document new commands.

## Dependencies
* python 3.9
* pip 21.1.2
* virtualenv (run `pip install virtualenv --user`)

## Developer Getting started

Create and activate a virtual environment:
```bash
virtualenv venv
.\venv\Scripts\activate

#For Linux used t he below
#.venv/bin/activate
```

Pull the dependencies:
```bash
pip install -r requirements.txt
```

Then invoke the entry point to view available commands:
```
py pdp.py help
```

## Getting a .exe file for the CLI
```bash
pip install pyinstaller
pyinstaller .\pdp.py
```

This will create a 'dist/pdp' folder, whose contents are the necessary run time dependencies for the CLI.

This can be zipped and shared, and then to install, just add that folder to your Windows PATH.

## Configuring Intellij

Follow the guide [here](https://www.jetbrains.com/help/idea/creating-virtual-environment.html) so that Intellij can
recognize the virtual env.