# PDP Command Line Interface
Thin client around PDP Admin UI for common tasks. Intended for performing quick changes and as the foundation of more automation.

## Dependencies
 * python 3.9
 * pip 21.1.2

## Installing

Download the latest distributable ZIP file from here: https://github.com/pureinsights/pdp-cli/releases, unzip, and add to your Windows PATH environment variable.

## Getting started

### Sync configurations

If there are already configurations in the target environment, then you can fetch them to create your local work dir.

```bash
pdp init hello-world --adminApiUrl=http://localhost:8080
```

The above command will fetch the configurations (seeds, credentials, processors, pipelines, etc) from the target environment and create a
local directory with the layout for all existing configurations. **Notice that imported configs have id fields, don't change those**

If you would rather start a new project, you can do something like:
```bash
pdp init hello-word --empty
```

This will create an empty directory with the PDP layout for you to start working. Adjust the corresponding Admin API URL on pdp.ini.

Once you have done changes locally, you can do:

```bash
pdp config deploy
```

Which will validate and upload your changes. This will perform basic tagging to make sure you are not overriding someone else's changes. 

## Documentation

Access documentation by running the help command:

```bash
pdp help
```

## Getting a .exe file for the CLI
```bash
pip install pyinstaller
pyinstaller .\pdp.py
```

This will create a 'dist/pdp' folder, whose contents are the necessary run time dependencies for the CLI.

This can be zipped and shared, and then to install, just add that folder to your Windows PATH.

## Code changes
Pull the dependencies:
```bash
pip install -r requirements.txt
```

Then simply invoke the entry point:
```
py pdp.py help
```

