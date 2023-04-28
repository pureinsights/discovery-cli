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
import os

from commons.custom_classes import PdpEntity

# URLs
INGESTION_API_URL = 'http://localhost:8080'
STAGING_API_URL = 'http://localhost:8081'
CORE_API_URL = 'http://localhost:8082'
DISCOVERY_API_URL = 'http://localhost:8088/admin'

# Endpoints
GENERIC_URL = '{0}/{entity}/{id}'
URL_EXPORT = '{0}/export/{entity}'
URL_IMPORT = '{0}/import'
URL_GET_BY_ID = GENERIC_URL
URL_GET_ALL = '{0}/{entity}'
URL_UPDATE = GENERIC_URL
URL_CREATE = '{0}/{entity}'
URL_DELETE = GENERIC_URL
URL_SEARCH = '{0}/search'
URL_UPLOAD_FILE = '{0}/files'
URL_DOWNLOAD_FILE = '{0}/files/download'

# Must be all in lower case
# Products
INGESTION = 'ingestion'
DISCOVERY = 'discovery'
CORE = 'core'
STAGING = 'staging'

# Entities
SCHEDULER = PdpEntity(INGESTION, 'scheduler', 'cron_jobs.json')
INGESTION_PROCESSOR_ENTITY = PdpEntity(INGESTION, 'processor', 'processors.json')
PIPELINE = PdpEntity(INGESTION, 'pipeline', 'pipelines.json')
SEED = PdpEntity(INGESTION, 'seed', 'seeds.json')
CREDENTIAL = PdpEntity(CORE, 'credential', 'credentials.json')
ENDPOINT = PdpEntity(DISCOVERY, 'endpoint', 'endpoints.json')
DISCOVERY_PROCESSOR_ENTITY = PdpEntity(DISCOVERY, 'processor', 'processors.json', 'processors')
# Must be in order (based on which has fewer dependencies to another entities)
ENTITIES = CREDENTIAL, INGESTION_PROCESSOR_ENTITY, PIPELINE, SEED, SCHEDULER, DISCOVERY_PROCESSOR_ENTITY, ENDPOINT

INGESTION_PROCESSOR = {'name': 'ingestionprocessor', 'entity': INGESTION_PROCESSOR_ENTITY}
DISCOVERY_PROCESSOR = {'name': 'discoveryprocessor', 'entity': DISCOVERY_PROCESSOR_ENTITY}

# Configurations
DEFAULT_CONFIG = {
  INGESTION: INGESTION_API_URL,
  DISCOVERY: DISCOVERY_API_URL,
  CORE: CORE_API_URL,
  STAGING: STAGING_API_URL
}

# The entity types are in order to deploy
PRODUCTS = {
  'list': [CORE, INGESTION, DISCOVERY, STAGING],
  INGESTION: {
    'entities': [INGESTION_PROCESSOR_ENTITY, PIPELINE, SEED, SCHEDULER],
  },
  DISCOVERY: {
    'entities': [DISCOVERY_PROCESSOR_ENTITY, ENDPOINT],
  },
  CORE: {
    'entities': [CREDENTIAL],
  },
  STAGING: {
    'entities': []
  },
}

# Common messages
WARNING_FORMAT = '[WARNING]: {message}'
ERROR_FORMAT = '[ERROR]: {message}'
EXCEPTION_FORMAT = 'Some thing went wrong due to: {exception}.{error}'

FROM_NAME_FORMAT = "{{{{ fromName('{0}') }}}}"

# Severities
ERROR_SEVERITY = 'error'
WARNING_SEVERITY = 'warning'

# File paths
TEMPLATES_DIRECTORY = os.path.join(os.path.dirname(__file__), '..', 'commands', 'config', 'templates')
